package controllers

import (
	"ipfs-sharing/misk"
	"ipfs-sharing/models"
	"net/http"
)

func (control *Controller) CheckUpdate(w http.ResponseWriter, r *http.Request) {
	control.Respond(w, models.Update{Version: misk.Version, Cid: misk.UpdateCid})
}
