package controllers

import (
	"encoding/json"
	"github.com/anacrolix/log"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"ipfs-sharing/gen/model"
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
	control.gu.SearchW.TreeAdd(content)
	log.Println("Searched", content.Name, content.Cid)
}

func (control *Controller) SearchDht(r iface.PubSubMessage) {
	content, err := control.inter.SearchMyContent(string(r.Data()))
	if err != nil {
		return
	}
	control.gu.SearchW.TreeAdd(content)
	_, err = control.inter.PostJson(r.From().String()+"/search/answer", content)
	if err != nil {
		log.Println(err)
		return
	}
}

func (control *Controller) GetChildren(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	contents, err := control.inter.GetChildrenContents(int32(id))
	if err != nil {
		log.Println(err)
	}

	control.Respond(w, contents)
}
