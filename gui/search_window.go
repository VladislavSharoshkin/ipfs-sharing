package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ipfs/go-cid"
	"ipfs-sharing/gen/model"
	"ipfs-sharing/internal"
	"ipfs-sharing/misk"
	"ipfs-sharing/models"
	"log"
	"path/filepath"
)

type SearchWindow struct {
	Cont     *fyne.Container
	Tree     *widget.Tree
	TreeData map[string]ContentNode
	Inter    *internal.Internal
}

type ContentNode struct {
	Content  model.Contents
	Children []string
}

func NewSearchWindow(inter *internal.Internal) *SearchWindow {

	sw := SearchWindow{Inter: inter}

	sw.TreeData = make(map[string]ContentNode)
	sw.Tree = widget.NewTree(
		func(uid string) (children []string) {
			children = sw.TreeData[uid].Children
			return
		},
		func(uid string) bool {
			// hot fix
			return len(sw.TreeData[uid].Content.Cid) < 4
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(uid string, branch bool, node fyne.CanvasObject) {
			content := sw.TreeData[uid].Content
			node.(*widget.Label).SetText(models.ContentToString(content))
		},
	)
	var selectedContent model.Contents
	menuC := container.NewHBox(
		widget.NewButton("Download", func() {
			err := inter.Download(selectedContent)
			if err != nil {
				log.Println(err)
				return
			}
		}),
		widget.NewButton("Open browser", func() {
			misk.OpenBrowser(fmt.Sprint("https://ipfs.io/ipfs/", selectedContent.Cid))
		}),
		widget.NewButton("Open folder", func() {
			misk.OpenFolder(filepath.Join(inter.Opt.ShareDir, selectedContent.Dir))
		}),
		widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {

		}),
	)
	menuC.Hidden = true

	sw.Tree.OnSelected = func(uid widget.TreeNodeID) {
		selectedContent = sw.TreeData[uid].Content
		menuC.Hidden = false
	}

	sw.Tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		selectedNode := sw.TreeData[uid]
		if len(selectedNode.Children) != 0 {
			return
		}

		var err error
		var children []model.Contents
		if selectedNode.Content.Status != "" {
			children, err = inter.GetChildrenContents(selectedNode.Content.ID)
			if err != nil {
				return
			}
		} else {
			children, err = sw.Inter.Hc.GetChildren(selectedNode.Content.From, selectedNode.Content.ID, false)
			if err != nil {
				return
			}
		}
		sw.TreeAdds(children)
	}

	searchE := widget.NewEntry()
	searchE.PlaceHolder = "Name or node address"

	searchButton := widget.NewButton("Search", func() {
		searchType := ""
		_, err := cid.Parse(searchE.Text)
		if err == nil {
			searchType = "cid"
		}
		if searchE.Text == "" {
			searchType = "my"
		}

		switch searchType {
		case "cid":
			children, err := sw.Inter.Hc.GetChildren(searchE.Text, 0, false)
			if err != nil {
				return
			}
			sw.TreeAdds(children)
		case "my":
			contents, err := inter.GetChildrenContents(0)
			if err != nil {
				return
			}
			sw.TreeAdds(contents)
		default:
			err = inter.Search(searchE.Text)
			if err != nil {
				log.Println(err)
				return
			}
		}
	})
	searchContainer := container.NewBorder(nil, nil, nil, searchButton, searchE)

	sw.Cont = container.NewBorder(searchContainer, menuC, nil, nil, sw.Tree)

	return &sw
}

func (sw *SearchWindow) TreeAdds(contents []model.Contents) {
	for _, cont := range contents {
		sw.TreeAdd(cont)
	}
}

func (sw *SearchWindow) TreeAdd(namedCid model.Contents) {
	parentID := ""
	if namedCid.ParentID != nil {
		parentID = string(*namedCid.ParentID)
	}

	parent := sw.TreeData[parentID]

	parent.Children = append(parent.Children, string(namedCid.ID))

	sw.TreeData[parentID] = parent
	sw.TreeData[string(namedCid.ID)] = ContentNode{Content: namedCid}
	sw.Tree.Refresh()
}
