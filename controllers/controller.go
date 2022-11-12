package controllers

import (
	"encoding/json"
	"ipfs-sharing/gui"
	"ipfs-sharing/internal"
	"net/http"
)

type Controller struct {
	gu    *gui.Gui
	inter *internal.Internal
}

func NewController(gu *gui.Gui, inter *internal.Internal) *Controller {
	return &Controller{gu, inter}
}

func (control *Controller) Respond(w http.ResponseWriter, data interface{}) {
	//pp.Print(data)
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(data)
}
