package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"ipfs-sharing/internal"
	"ipfs-sharing/misk"
	"log"
)

type DashboardWindow struct {
	Cont  *fyne.Container
	Inter *internal.Internal
}

func NewDashboardWindow(inter *internal.Internal) *DashboardWindow {

	addressL := widget.NewMultiLineEntry()
	addressL.Text = misk.SPrintValues("My address:", inter.Node.IpfsNode.Identity.String(), "\n",
		"Version", misk.Version)

	shareB := widget.NewButtonWithIcon("Open share folder", theme.FolderOpenIcon(), func() {
		misk.OpenFolder(inter.Opt.ShareDir)
	})

	updateShareB := widget.NewButtonWithIcon("Scan share folder", theme.ViewRefreshIcon(), func() {
		err := inter.Sync()
		if err != nil {
			log.Println(err)
		}
	})

	checkUpdatesB := widget.NewButton("Check updates", func() {
		inter.Update()
	})

	cont := container.NewVBox(addressL, shareB, updateShareB, checkUpdatesB)

	return &DashboardWindow{cont, inter}
}
