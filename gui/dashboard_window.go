package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"ipfs-sharing/internal"
	"log"
	"os/exec"
	"runtime"
)

type DashboardWindow struct {
	Cont  *fyne.Container
	Inter *internal.Internal
}

func NewDashboardWindow(inter *internal.Internal) *DashboardWindow {

	addressL := widget.NewMultiLineEntry()
	addressL.Text = "My address: " + inter.Node.IpfsNode.Identity.String()

	shareB := widget.NewButtonWithIcon("Open share folder", theme.FolderOpenIcon(), func() {
		cmd := "open"
		if runtime.GOOS == "windows" {
			cmd = "explorer"
		}
		err := exec.Command(cmd, inter.Opt.ShareDir).Start()
		if err != nil {
			log.Println(err)
		}
	})

	updateShareB := widget.NewButtonWithIcon("Scan share folder", theme.ViewRefreshIcon(), func() {
		err := inter.Sync()
		if err != nil {
			log.Println(err)
		}
	})

	cont := container.NewVBox(addressL, shareB, updateShareB)

	return &DashboardWindow{cont, inter}
}
