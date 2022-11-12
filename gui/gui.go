package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cenkalti/rain/torrent"
	"ipfs-sharing/gui/elements"
	"ipfs-sharing/internal"
	"log"
	"time"
)

type Gui struct {
	SearchW *SearchWindow
	Inter   *internal.Internal
	CurW    fyne.Window
	TorrW   *TorrentWindow
	DashW   *DashboardWindow
	chatW   *ChatWindow
}

func MakeGui(inter *internal.Internal) *Gui {

	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(theme.FyneLogo())
	MakeTray(a)

	w := a.NewWindow("Fyne Demo")

	//w.SetMainMenu(MakeMenu(a, w))
	w.SetMaster()

	torrW := NewTorrentWindow(inter)
	searchW := NewSearchWindow(inter)
	dashW := NewDashboardWindow(inter)
	chatW := NewChatWindow(inter)

	peersL := elements.NewToolbarLabel("")

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
			err := inter.SyncFilesAndDatabase(inter.Options.ShareDir, nil)
			if err != nil {
				log.Println(err)
				return
			}
		}),
		widget.NewToolbarSpacer(),
		peersL,
	)

	go func() {
		for range time.Tick(5 * time.Second) {
			var toDownload int64
			for _, torr := range inter.TorrSes.ListTorrents() {
				if torr.Stats().Status == torrent.Downloading {
					toDownload += torr.Stats().Bytes.Incomplete
				}
			}
			peersL.SetText(inter.Status())
		}
	}()

	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", dashW.Cont),
		container.NewTabItem("Search", searchW.Cont),
		container.NewTabItem("Torrent", torrW.Cont),
		container.NewTabItem("Chat", chatW.Cont),
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	content := container.NewBorder(toolbar, nil, nil, nil, tabs)
	w.SetContent(content)

	w.Resize(fyne.NewSize(640, 460))

	return &Gui{searchW, inter, w, torrW, dashW, chatW}
}

func (gu *Gui) ShowAndRun() {
	gu.CurW.ShowAndRun()
}
