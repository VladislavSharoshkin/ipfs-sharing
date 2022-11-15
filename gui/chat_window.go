package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"ipfs-sharing/internal"
	"ipfs-sharing/models"
	"log"
)

type ChatWindow struct {
	Cont      *fyne.Container
	Inter     *internal.Internal
	ListData  []models.Message
	List      *widget.List
	RoomsList *widget.List
}

func NewChatWindow(inter *internal.Internal) *ChatWindow {

	cw := ChatWindow{Inter: inter}

	addressE := widget.NewEntry()
	cw.ListData = append(cw.ListData, models.NewMessage("", "", "hi"))
	cw.List = widget.NewList(
		func() int {
			return len(cw.ListData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			mes := cw.ListData[i]
			o.(*widget.Label).SetText(mes.Text)
		})

	cw.RoomsList = widget.NewList(
		func() int {
			return len(cw.ListData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			mes := cw.ListData[i]
			o.(*widget.Label).SetText(mes.Text)
		})

	messageE := widget.NewEntry()
	sendB := widget.NewButton("Send", func() {
		mes := models.NewMessage(inter.Node.IpfsNode.Identity.String(), addressE.Text, messageE.Text)
		_, err := inter.PostJson(mes.ToID+"/message/new", mes)
		if err != nil {
			log.Println(err)
		}

		cw.MessageAdd(mes)
	})
	sendC := container.NewBorder(nil, nil, nil, sendB, messageE)

	cw.Cont = container.NewBorder(addressE, sendC, cw.RoomsList, nil, cw.List)

	return &cw
}

func (cw *ChatWindow) MessageAdd(message models.Message) {
	cw.ListData = append(cw.ListData, message)
	cw.List.Refresh()
}
