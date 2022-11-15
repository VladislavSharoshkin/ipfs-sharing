package controllers

import (
	"encoding/json"
	"ipfs-sharing/models"
	"log"
	"net/http"
)

func (control *Controller) NewMessage(w http.ResponseWriter, r *http.Request) {
	mes := models.Message{}
	err := json.NewDecoder(r.Body).Decode(&mes)
	if err != nil {
		log.Println(err)
		return
	}

	control.gu.ChatW.MessageAdd(mes)
}
