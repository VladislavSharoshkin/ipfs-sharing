package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"ipfs-sharing/internal"
)

type ChatWindow struct {
	Cont  *fyne.Container
	Inter *internal.Internal
}

func NewChatWindow(inter *internal.Internal) *ChatWindow {

	addressE := widget.NewEntry()

	var messages []string
	list := widget.NewList(
		func() int {
			return len(messages)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {

			o.(*widget.Label).SetText("test")
		})

	messageE := widget.NewEntry()
	sendB := widget.NewButton("Send", func() {

	})
	sendC := container.NewBorder(nil, nil, nil, sendB, messageE)

	cont := container.NewBorder(addressE, sendC, nil, nil, list)

	return &ChatWindow{cont, inter}
}
