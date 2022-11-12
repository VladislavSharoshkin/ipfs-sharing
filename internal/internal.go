package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/rain/torrent"
	"github.com/go-jet/jet/v2/qrm"
	. "github.com/go-jet/jet/v2/sqlite"
	"github.com/ipfs/go-cid"
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
	Options *Options
	Node    *Node
	DB      *Database
	Sub     iface.PubSubSubscription
	Cl      *http.Client
	TorrSes *torrent.Session
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

	node, err := NewNode(context.Background(), opt.IpfsDir)
	if err != nil {
		log.Println(err)
		return nil
	}

	db := NewDatabase(opt)

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

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(node.IpfsNode.PeerHost))
	cl := &http.Client{Transport: tr}

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

	internal := Internal{opt, node, db, sub, cl, torrSes}
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

func (inter *Internal) Download(content model.Contents) error {

	file, err := inter.Node.CoreAPI.Unixfs().Get(context.Background(), path.New(content.Cid))
	if err != nil {
		return err
	}

	savePath := filepath.Join(inter.Options.ShareDir, content.Name)

	err = files.WriteTo(file, savePath)
	if err != nil {
		return err
	}

	log.Println("Downloaded CID", content.Cid)

	return nil
}

func (inter *Internal) PostJson(url string, structData interface{}) (*http.Response, error) {
	postBody, err := json.Marshal(structData)
	if err != nil {
		return nil, err
	}
	resp, err := inter.Cl.Post("libp2p://"+url, "application/json", bytes.NewBuffer(postBody))
	if err != nil {
		return nil, err
	}

	return resp, err
}

func (inter *Internal) SyncFilesAndDatabase(input string, parentID *int32) error {
	dirEntries, _ := os.ReadDir(input)
	for _, dirEntry := range dirEntries {
		name := input + "/" + dirEntry.Name()
		println(name)

		parentIDExpression := Contents.ParentID.IS_NULL()
		if parentID != nil {
			parentIDExpression = Contents.ParentID.EQ(Int32(*parentID))
		}

		stmt := SELECT(Contents.AllColumns).FROM(Contents).
			WHERE(Contents.Name.EQ(String(dirEntry.Name())).AND(parentIDExpression)).
			LIMIT(1)

		err := stmt.Query(inter.DB.DB, &model.Contents{})
		if err != nil && err != qrm.ErrNoRows {
			return err
		}
		if err == nil {
			return nil
		}

		uploadCid := cid.Cid{}
		if !dirEntry.IsDir() {
			uploadCid, err = inter.Node.Upload(name)
			if err != nil {
				return err
			}
		}

		newCont := models.NewContent(dirEntry.Name(), uploadCid.String(), parentID)
		err = inter.DB.InsertContent(&newCont)
		if err != nil {
			return err
		}

		if dirEntry.IsDir() {
			inter.SyncFilesAndDatabase(name, &newCont.ID)
		}
	}

	return nil
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
	err := stmt.Query(inter.DB.DB, &content)
	if err != nil {
		return content, err
	}

	return content, nil
}

func (inter *Internal) GetChildrenContents(id int32) ([]model.Contents, error) {
	// search for some text
	stmt := SELECT(Contents.AllColumns).FROM(Contents).
		WHERE(Contents.ParentID.EQ(Int32(id)))

	var contents []model.Contents
	err := stmt.Query(inter.DB.DB, &contents)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
