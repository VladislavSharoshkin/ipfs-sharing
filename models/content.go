package models

import (
	iface "github.com/ipfs/interface-go-ipfs-core"
	"ipfs-sharing/gen/model"
)

type ContentStatus int64

func NewContent(Name string, Cid string, parentID *int32) model.Contents {
	return model.Contents{Name: Name, Cid: Cid, ParentID: parentID}
}

func ContentFromEntry(entry iface.DirEntry) model.Contents {
	return NewContent(entry.Name, entry.Cid.String(), nil)
}

func ContentFromEntryList(dirs []iface.DirEntry) []model.Contents {
	contents := make([]model.Contents, 0, len(dirs))
	for _, dir := range dirs {
		contents = append(contents, ContentFromEntry(dir))
	}

	return contents
}
