package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/cenkalti/log"
	"ipfs-sharing/internal"
	"os/exec"
	"runtime"
)

type DashboardWindow struct {
	Cont  *fyne.Container
	Inter *internal.Internal
}

func NewDashboardWindow(inter *internal.Internal) *DashboardWindow {

	addressL := widget.NewLabel("My address: " + inter.Node.IpfsNode.Identity.String())

	shareB := widget.NewButtonWithIcon("Open share folder", theme.FolderOpenIcon(), func() {
		cmd := "open"
		if runtime.GOOS == "windows" {
			cmd = "explorer"
		}
		err := exec.Command(cmd, inter.Options.ShareDir).Start()
		if err != nil {
			log.Errorln(err)
		}
	})

	updateShareB := widget.NewButtonWithIcon("Scan share folder", theme.ViewRefreshIcon(), func() {
		err := inter.SyncFilesAndDatabase(inter.Options.ShareDir, nil)
		if err != nil {
			log.Errorln(err)
		}
	})

	cont := container.NewVBox(addressL, shareB, updateShareB)

	return &DashboardWindow{cont, inter}
}
