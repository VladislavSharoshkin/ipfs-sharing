package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/ipfs/go-cid"
	"ipfs-sharing/gen/model"
	"ipfs-sharing/internal"
	"ipfs-sharing/models"
	"log"
)

type ChatWindow struct {
	Cont         *fyne.Container
	Inter        *internal.Internal
	MessagesData []model.Messages
	MessagesL    *widget.List
	RoomsL       *widget.List
	ChatL        *widget.Label
	RoomsData    []string
	Address      string
}

func NewChatWindow(inter *internal.Internal) *ChatWindow {

	cw := ChatWindow{Inter: inter}

	cw.MessagesL = widget.NewList(
		func() int {
			return len(cw.MessagesData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			mes := cw.MessagesData[i]
			o.(*widget.Label).SetText(mes.Text)
		})

	cw.RoomsL = widget.NewList(
		func() int {
			return len(cw.RoomsData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			room := cw.RoomsData[i]
			o.(*widget.Label).SetText(room)
		})
	cw.RoomsL.OnSelected = func(id widget.ListItemID) {
		address := cw.RoomsData[id]
		cw.SetChat(address)
	}
	cw.UpdateRooms()

	messageE := widget.NewEntry()
	sendB := widget.NewButton("Send", func() {
		mes := models.NewMessage(inter.ID, cw.Address, messageE.Text)
		_, err := inter.Hc.PostJson(mes.To+"/message/new", mes)
		if err != nil {
			log.Println(err)
		}

		err = inter.DB.Save(&mes)
		if err != nil {
			return
		}
		fmt.Println(mes.ID)

		cw.MessageAdd(mes)
		cw.UpdateRooms()
	})
	sendC := container.NewBorder(nil, nil, nil, sendB, messageE)

	cw.ChatL = widget.NewLabel("")
	chatC := container.NewBorder(cw.ChatL, sendC, nil, nil, cw.MessagesL)
	addressE := widget.NewEntry()
	addressE.PlaceHolder = "Address"
	addressE.OnChanged = func(s string) {
		_, err := cid.Parse(s)
		if err != nil {
			return
		}
		cw.SetChat(s)
	}
	roomC := container.NewBorder(addressE, widget.NewLabel("Conversations"), nil, nil, cw.RoomsL)
	roomC.Size()

	cw.Cont = container.NewBorder(nil, nil, roomC, nil, chatC)

	return &cw
}

func (cw *ChatWindow) MessageAdd(message model.Messages) {
	cw.MessagesData = append(cw.MessagesData, message)
	cw.MessagesL.Refresh()
}

func (cw *ChatWindow) UpdateRooms() {
	cw.RoomsData, _ = cw.Inter.DB.GetContacts()
	cw.RoomsL.Refresh()
}

func (cw *ChatWindow) UpdateMessages() {
	cw.MessagesData, _ = cw.Inter.DB.GetMessages(cw.Address)
	cw.MessagesL.Refresh()
}

func (cw *ChatWindow) SetChat(address string) {
	cw.Address = address
	cw.ChatL.SetText(address)
	cw.UpdateMessages()
}
