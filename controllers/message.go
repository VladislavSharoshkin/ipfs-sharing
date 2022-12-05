package controllers

import (
	"encoding/json"
	"ipfs-sharing/gen/model"
	"log"
	"net/http"
)

func (control *Controller) NewMessage(w http.ResponseWriter, r *http.Request) {
	mes := model.Messages{}
	err := json.NewDecoder(r.Body).Decode(&mes)
	if err != nil {
		log.Println(err)
		return
	}
	err = control.inter.DB.Save(&mes)
	if err != nil {
		return
	}

	control.gu.ChatW.MessageAdd(mes)
}
