package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cenkalti/rain/torrent"
	"ipfs-sharing/internal"
	"ipfs-sharing/misk"
	"log"
)

type TorrentWindow struct {
	Cont  *fyne.Container
	List  *widget.List
	Inter *internal.Internal
}

func NewTorrentWindow(inter *internal.Internal) *TorrentWindow {
	torrents := inter.TorrSes.ListTorrents()

	list := widget.NewList(
		func() int {
			return len(torrents)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			torr := torrents[i]
			o.(*widget.Label).SetText(misk.SPrintValues(torr.Stats().Status, torr.Name()))
		})

	var selectedTorrent *torrent.Torrent
	list.OnSelected = func(id widget.ListItemID) {
		selectedTorrent = torrents[id]
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			err := selectedTorrent.Start()
			if err != nil {
				log.Println(err)
				return
			}
			torrents = inter.TorrSes.ListTorrents()
			list.Refresh()
		}),
		widget.NewToolbarAction(theme.MediaPauseIcon(), func() {
			err := selectedTorrent.Stop()
			if err != nil {
				log.Println(err)
				return
			}
			torrents = inter.TorrSes.ListTorrents()
			list.Refresh()
		}),
		widget.NewToolbarAction(theme.DeleteIcon(), func() {
			err := inter.TorrSes.RemoveTorrent(selectedTorrent.ID())
			if err != nil {
				log.Println(err)
				return
			}
			torrents = inter.TorrSes.ListTorrents()
			list.Refresh()
		}),
	)

	magnetL := widget.NewEntry()
	magnetL.PlaceHolder = "Magnet link"

	magnetB := widget.NewButton("Add", func() {
		_, err := inter.TorrSes.AddURI(magnetL.Text, nil)
		if err != nil {
			log.Println(err)
			return
		}
		torrents = inter.TorrSes.ListTorrents()
		list.Refresh()
	})
	magnetC := container.NewBorder(nil, nil, nil, magnetB, magnetL)

	cont := container.NewBorder(magnetC, toolbar, nil, nil, list)

	return &TorrentWindow{cont, list, inter}
}
