package gui

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"ipfs-sharing/gen/model"
	"ipfs-sharing/internal"
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

	treeData := make(map[string]ContentNode)
	tree := widget.NewTree(
		func(uid string) (children []string) {
			children = treeData[uid].Children
			return
		},
		func(uid string) bool {
			return true
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(uid string, branch bool, node fyne.CanvasObject) {
			node.(*widget.Label).SetText(treeData[uid].Content.Name)
		},
	)
	var selectedContent model.Contents
	tree.OnSelected = func(uid widget.TreeNodeID) {
		selectedContent = treeData[uid].Content
	}

	tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		selectedContent = treeData[uid].Content

		r, err := inter.PostJson(fmt.Sprint(*selectedContent.From, "/content/children?id=", selectedContent.ID), nil)
		if err != nil {
			return
		}

		var contents []model.Contents
		err = json.NewDecoder(r.Body).Decode(&contents)
		if err != nil {
			log.Println(err)
			return
		}

		newChildren := make([]string, 0, len(contents))
		for _, cont := range contents {
			cont.From = selectedContent.From
			newChildren = append(newChildren, string(cont.ID))
			treeData[string(cont.ID)] = ContentNode{cont, nil}
			tree.Refresh()
		}
		treeData[uid] = ContentNode{selectedContent, newChildren}
	}

	entrySearch := widget.NewEntry()

	searchButton := widget.NewButton("Search", func() {
		err := inter.Search(entrySearch.Text)
		if err != nil {
			log.Println(err)
			return
		}
	})
	searchContainer := container.NewBorder(nil, nil, nil, searchButton, entrySearch)

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.DownloadIcon(), func() {
			err := inter.Download(selectedContent)
			if err != nil {
				log.Println(err)
				return
			}
		}),
	)

	cont := container.NewBorder(searchContainer, toolbar, nil, nil, tree)

	return &SearchWindow{cont, tree, treeData, inter}
}

func (sw *SearchWindow) TreeAdd(namedCid model.Contents) {
	parent := ""
	if namedCid.ParentID != nil {
		parent = string(sw.TreeData[string(*namedCid.ParentID)].Content.ID)
	}

	sw.TreeData[parent] = ContentNode{namedCid, []string{string(namedCid.ID)}}
	sw.TreeData[string(namedCid.ID)] = ContentNode{Content: namedCid}
	sw.Tree.Refresh()
}
