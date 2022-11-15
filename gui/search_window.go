package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ipfs/go-cid"
	"ipfs-sharing/gen/model"
	"ipfs-sharing/internal"
	"ipfs-sharing/misk"
	"log"
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
			node.(*widget.Label).SetText(misk.SPrintValues(content.Status, content.Name))
		},
	)
	var selectedContent model.Contents
	sw.Tree.OnSelected = func(uid widget.TreeNodeID) {
		selectedContent = sw.TreeData[uid].Content
	}

	sw.Tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		selectedNode := sw.TreeData[uid]
		if len(selectedNode.Children) != 0 {
			return
		}

		var err error
		var children []model.Contents
		if selectedNode.Content.From == inter.ID {
			children, err = inter.GetChildrenContents(selectedNode.Content.ID)
			if err != nil {
				return
			}
		} else {
			children, err = sw.Inter.GetChildren(selectedNode.Content.From, selectedNode.Content.ID)
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
			children, err := sw.Inter.GetChildren(searchE.Text, 0)
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

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DownloadIcon(), func() {
			err := inter.Download(selectedContent)
			if err != nil {
				log.Println(err)
				return
			}
		}),
	)

	sw.Cont = container.NewBorder(searchContainer, toolbar, nil, nil, sw.Tree)

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
