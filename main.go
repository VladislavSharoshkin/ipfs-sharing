package main

import (
	"context"
	"ipfs-sharing/controllers"
	"ipfs-sharing/gui"
	"ipfs-sharing/internal"
	"log"
	"net/http"
)

func main() {

	inter := internal.NewInternal()
	gu := gui.MakeGui(inter)
	control := controllers.NewController(gu, inter)
	http.HandleFunc("/search/answer", control.ContentSearchAnswer)
	http.HandleFunc("/content/children", control.GetChildren)
	http.HandleFunc("/message/new", control.NewMessage)
	http.HandleFunc("/update/check", control.CheckUpdate)
	go func() {
		for {
			mes, err := inter.Sub.Next(context.Background())
			if err != nil {
				log.Println(err)
				continue
			}

			if mes.From() == inter.Node.IpfsNode.Identity {
				continue
			}
			control.SearchDht(mes)
		}
	}()

	gu.ShowAndRun()
}
