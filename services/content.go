package services

import (
	"ipfs-sharing/gen/model"
	"ipfs-sharing/internal"
)

func GetChildren(inter *internal.Internal, id int32, isRecursive bool) ([]model.Contents, error) {
	var contents []model.Contents
	var err error

	if isRecursive {
		err = inter.GetChildrenRecursive(id, &contents)
	} else {
		contents, err = inter.GetChildrenContents(id)
	}
	if err != nil {
		return nil, nil
	}

	return contents, err
}
