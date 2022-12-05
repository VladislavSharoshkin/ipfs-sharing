package controllers

import (
	"encoding/json"
	"github.com/anacrolix/log"
	"github.com/go-jet/jet/v2/qrm"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"ipfs-sharing/gen/model"
	"ipfs-sharing/models"
	"ipfs-sharing/services"
	"net/http"
	"strconv"
)

func (control *Controller) ContentSearchAnswer(w http.ResponseWriter, r *http.Request) {
	content := model.Contents{}
	err := json.NewDecoder(r.Body).Decode(&content)
	if err != nil {
		log.Println(err)
		return
	}
	content.From = r.RemoteAddr
	content.Status = models.ContentStatusStopped

	var myContent model.Contents
	err = control.inter.DB.FirstByCid(content.Cid).
		Query(control.inter.DB.Conn, &myContent)
	if err != nil && err != qrm.ErrNoRows {
		return
	}
	if myContent.ID != 0 {
		content.Status = myContent.Status
	}

	control.gu.SearchW.TreeAdd(content)
	log.Println("Searched", content.Name, content.Cid)
}

func (control *Controller) SearchDht(r iface.PubSubMessage) {
	content, err := control.inter.SearchMyContent(string(r.Data()))
	if err != nil {
		return
	}
	control.gu.SearchW.TreeAdd(content)
	_, err = control.inter.Hc.PostJson(r.From().String()+"/search/answer", content)
	if err != nil {
		log.Println(err)
		return
	}
}

func (control *Controller) GetChildren(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	isRecursive, _ := strconv.ParseBool(r.URL.Query().Get("recursive"))

	children, err := services.GetChildren(control.inter, int32(id), isRecursive)
	if err != nil {
		return
	}

	control.Respond(w, children)
}

func (control *Controller) GetDependencies(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	contents, err := control.inter.DB.GetContentsDependencies(int32(id))
	if err != nil {
		return
	}

	control.Respond(w, contents)
}
