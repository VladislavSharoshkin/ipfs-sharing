package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/rain/torrent"
	. "github.com/go-jet/jet/v2/sqlite"
	files "github.com/ipfs/go-ipfs-files"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/ipfs/interface-go-ipfs-core/path"
	gostream "github.com/libp2p/go-libp2p-gostream"
	p2phttp "github.com/libp2p/go-libp2p-http"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"ipfs-sharing/gen/model"
	. "ipfs-sharing/gen/table"
	"ipfs-sharing/misk"
	"ipfs-sharing/models"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Internal struct {
	Opt     *Options
	Node    *Node
	DB      *Database
	Sub     iface.PubSubSubscription
	Hc      *HttpClient
	TorrSes *torrent.Session
	ID      string
}

func NewInternal() *Internal {
	opt := NewOptions()

	logFile, err := os.OpenFile(opt.DataDir+"/log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(mw)

	db := NewDatabase(opt)

	node, err := NewNode(context.Background(), opt.IpfsDir)
	if err != nil {
		log.Println(err)
		return nil
	}

	// libp2p http server
	listener, _ := gostream.Listen(node.IpfsNode.PeerHost, p2phttp.DefaultP2PProtocol)
	server := &http.Server{}
	go func() {
		defer listener.Close()
		err := server.Serve(listener)
		if err != nil {
			log.Println(err)
		}
	}()

	go func() {
		http.ListenAndServe("127.0.0.1:8080", nil)
	}()

	sub, err := node.CoreAPI.PubSub().Subscribe(context.Background(), "ipfs-sharing")
	if err != nil {
		log.Println(err)
		return nil
	}

	hc := NewHttpClient(node.IpfsNode.PeerHost)

	torConfig := torrent.DefaultConfig
	torConfig.DataDir = opt.ShareDir
	torConfig.Database = filepath.Join(opt.DataDir, "session.db")
	torConfig.DataDirIncludesTorrentID = false
	torConfig.RPCEnabled = false
	torConfig.DHTEnabled = false
	torrSes, err := torrent.NewSession(torConfig)
	if err != nil {
		log.Println(err)
		return nil
	}

	//if len(os.Args) > 1 {
	//	f, err := os.Open(os.Args[1])
	//	if err != nil {
	//		log.Println(err)
	//		return nil
	//	}
	//	_, err = torrSes.AddTorrent(f, nil)
	//	if err != nil {
	//		log.Println(err)
	//		return nil
	//	}
	//}

	internal := Internal{opt, node, db, sub, hc, torrSes, node.IpfsNode.Identity.String()}
	return &internal
}

func (inter *Internal) Search(query string) error {
	api := inter.Node.CoreAPI
	err := api.PubSub().Publish(context.Background(), "ipfs-sharing", []byte(query))
	if err != nil {
		return err
	}

	return nil
}

func (inter *Internal) DownloadCid(fullPath string, fileCid string) (err error) {
	file, err := inter.Node.CoreAPI.Unixfs().Get(context.Background(), path.New(fileCid))
	if err != nil {
		return
	}

	err = files.WriteTo(file, fullPath)
	if err != nil {
		return
	}

	err = inter.Node.GC()
	if err != nil {
		return
	}

	return
}

func (inter *Internal) Download(content model.Contents) (err error) {

	r, err := inter.Hc.PostJson(fmt.Sprint(content.From, "/content/dependencies?id=", content.ID), nil)
	if err != nil {
		return err
	}

	var contents []model.Contents
	err = json.NewDecoder(r.Body).Decode(&contents)
	if err != nil {
		log.Println(err)
		return
	}

	oldIDs := make(map[int32]int32)
	for i, cont := range contents {
		isExist := false
		isExist, err = inter.DB.IsExist(inter.DB.SmtpByDirAndName(cont.Name, cont.Dir))
		if err != nil {
			return err
		}
		if isExist {
			continue
		}

		oldID := cont.ID
		cont.ID = 0
		cont.Status = models.ContentStatusDownloading

		var parentContent model.Contents
		if cont.ParentID != nil {
			newID := oldIDs[*cont.ParentID]
			cont.ParentID = &newID

			err = inter.DB.ByID(*cont.ParentID, &parentContent)
			if err != nil {
				return err
			}
		}

		cont.Dir = filepath.Join(parentContent.Dir, parentContent.Name)

		err = inter.DB.Save(&cont)
		if err != nil {
			return err
		}

		oldIDs[oldID] = cont.ID
		contents[i] = cont
	}

	inter.StartDownloads(contents)

	return
}

func (inter *Internal) Status() string {
	var total int64
	var complete int64
	for _, torr := range inter.TorrSes.ListTorrents() {
		if torr.Stats().Status == torrent.Downloading {
			total += torr.Stats().Bytes.Total
			complete += torr.Stats().Bytes.Completed
		}
	}
	download := ""
	if total > 0 {
		download = fmt.Sprint("Download: ", complete*100/total, "% ",
			"Speed: ", inter.TorrSes.Stats().SpeedDownload/1000, "KB")
	}

	return misk.SPrintValues("Peers:", inter.Node.IpfsNode.Peerstore.Peers().Len(), download)
}

func (inter *Internal) SearchMyContent(name string) (model.Contents, error) {

	stmt := SELECT(Contents.AllColumns).FROM(Contents).
		WHERE(Contents.Name.LIKE(String("%" + name + "%")))

	var content model.Contents
	err := stmt.Query(inter.DB.Conn, &content)
	if err != nil {
		return content, err
	}

	return content, nil
}

func (inter *Internal) GetChildrenRecursive(id int32, destination *[]model.Contents) error {
	contents, err := inter.GetChildrenContents(id)
	if err != nil {
		return err
	}
	*destination = append(*destination, contents...)

	for _, cont := range contents {
		err = inter.GetChildrenRecursive(cont.ID, destination)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inter *Internal) GetChildrenContents(id int32) ([]model.Contents, error) {

	parentIdExpression := Contents.ParentID.EQ(Int32(id))
	if id == 0 {
		parentIdExpression = Contents.ParentID.IS_NULL()
	}

	stmt := SELECT(Contents.AllColumns).FROM(Contents).
		WHERE(parentIdExpression)

	var contents []model.Contents
	err := stmt.Query(inter.DB.Conn, &contents)
	if err != nil {
		return nil, err
	}

	return contents, nil
}

func (inter *Internal) Update() {
	r, err := inter.Hc.PostJson(misk.CheckUpdateUrl, nil)
	if err != nil {
		return
	}

	var update models.Update
	err = json.NewDecoder(r.Body).Decode(&update)
	if err != nil {
		log.Println(err)
		return
	}

	if update.Version <= misk.Version {
		return
	}

	updateFilePath := filepath.Join(inter.Opt.ShareDir, "ipfs-sharing.zip")
	err = inter.DownloadCid(updateFilePath, update.Cid)
	if err != nil {
		return
	}
}

func (inter *Internal) DownloadContent(content model.Contents) {
	fullDir := filepath.Join(inter.Opt.ShareDir, content.Dir)
	os.MkdirAll(fullDir, os.ModePerm)

	fullPath := filepath.Join(fullDir, content.Name)

	err := inter.DownloadCid(fullPath, content.Cid)
	if err != nil {
		log.Println(err)
		return
	}

	uploadedCid, err := inter.Node.Upload(fullPath)

	content.Cid = uploadedCid
	content.Status = string(models.ContentStatusSaved)

	err = inter.DB.Save(&content)
	if err != nil {
		fmt.Println(err)
	}
}

func (inter *Internal) StartDownloads(contents []model.Contents) {
	go func() {
		for _, cont := range contents {
			inter.DownloadContent(cont)
		}
	}()
}

func (inter *Internal) StartUnfinishedDownloads() (err error) {

	smtp := inter.DB.SmtpUnfinishedDownloads()

	var contents []model.Contents
	err = inter.DB.Query(smtp, &contents)
	if err != nil {
		return
	}

	inter.StartDownloads(contents)

	return
}
