package models

import (
	"fmt"
	"ipfs-sharing/gen/model"
	"time"
)

type ContentStatus string

const (
	ContentStatusSaved       ContentStatus = "saved"
	ContentStatusDownloading               = "downloading"
	ContentStatusStopped                   = "stopped"
)

func NewContent(
	Name string,
	Cid string,
	parentID *int32,
	From string,
	Dir string,
	Status ContentStatus,
) model.Contents {

	now := time.Now().UTC().String()
	return model.Contents{
		Name:      Name,
		Cid:       Cid,
		ParentID:  parentID,
		CreatedAt: now,
		From:      From,
		Dir:       Dir,
		Status:    string(Status),
	}
}

func ContentToString(cont model.Contents) string {
	return fmt.Sprint("(", cont.Status, ") ", cont.Name)
}
