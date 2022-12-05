package internal

import (
	"github.com/go-jet/jet/v2/qrm"
	. "github.com/go-jet/jet/v2/sqlite"
	"github.com/ipfs/go-cid"
	_ "github.com/mattn/go-sqlite3"
	"ipfs-sharing/gen/model"
	. "ipfs-sharing/gen/table"
	"ipfs-sharing/models"
	"os"
	"path/filepath"
)

type ForAdd struct {
	Dir      string
	ParentID *int32
	DirEntry os.DirEntry
}

func (inter *Internal) recursiveScan(dirPath string, parentID *int32, scannedIDs *[]int32, forAdds *[]ForAdd) error {
	dirEntries, _ := os.ReadDir(dirPath)
	for _, dirEntry := range dirEntries {
		fullPath := filepath.Join(dirPath, dirEntry.Name())

		parentIDExpression := Contents.ParentID.IS_NULL()
		if parentID != nil {
			parentIDExpression = Contents.ParentID.EQ(Int32(*parentID))
		}

		stmt := SELECT(Contents.AllColumns).FROM(Contents).
			WHERE(Contents.Name.EQ(String(dirEntry.Name())).AND(parentIDExpression)).
			LIMIT(1)

		content := model.Contents{}
		err := stmt.Query(inter.DB.Conn, &content)
		if err != nil && err != qrm.ErrNoRows {
			return err
		}
		if content.ID != 0 {
			*scannedIDs = append(*scannedIDs, content.ID)
			inter.recursiveScan(fullPath, &content.ID, scannedIDs, forAdds)
		} else {
			*forAdds = append(*forAdds, ForAdd{dirPath, parentID, dirEntry})
		}
	}

	return nil
}

func (inter *Internal) recursiveAdd(dirPath string, parentID *int32, dirEntry os.DirEntry) error {
	fullPath := filepath.Join(dirPath, dirEntry.Name())
	relativeDir := dirPath[len(inter.Opt.ShareDir):]
	var err error

	newCid := ""
	if !dirEntry.IsDir() {
		newCid, err = inter.Node.Upload(fullPath)
		if err != nil {
			return err
		}
	}
	content := models.NewContent(dirEntry.Name(), newCid, parentID, inter.ID, relativeDir, models.ContentStatusSaved)

	err = inter.DB.InsertContent(&content)
	if err != nil {
		return err
	}

	if dirEntry.IsDir() {
		dirEntries, _ := os.ReadDir(fullPath)
		for _, dirEntry := range dirEntries {
			inter.recursiveAdd(fullPath, &content.ID, dirEntry)
		}
	}

	return nil
}

func (inter *Internal) delete(scannedIDs []int32) error {
	var deletedContents []model.Contents
	deleteStmt := Contents.DELETE().WHERE(RawInt("id").NOT_IN(InInt(scannedIDs)...)).
		RETURNING(Contents.AllColumns)

	err := deleteStmt.Query(inter.DB.Conn, &deletedContents)
	if err != nil {
		return err
	}

	for _, content := range deletedContents {
		stmt := Contents.SELECT(Contents.AllColumns).
			WHERE(Contents.Cid.EQ(String(content.Cid))).
			LIMIT(1)

		err = stmt.Query(inter.DB.Conn, &model.Contents{})
		if err != nil && err != qrm.ErrNoRows {
			return err
		}
		if err == qrm.ErrNoRows {
			parse, _ := cid.Parse(content.Cid)

			err = inter.Node.Delete(parse)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (inter *Internal) Sync() error {
	var scannedIDs []int32
	var forAdds []ForAdd

	err := inter.recursiveScan(inter.Opt.ShareDir, nil, &scannedIDs, &forAdds)
	if err != nil {
		return err
	}

	err = inter.delete(scannedIDs)
	if err != nil {
		return err
	}

	for _, forAdd := range forAdds {
		err := inter.recursiveAdd(forAdd.Dir, forAdd.ParentID, forAdd.DirEntry)
		if err != nil {
			return err
		}
	}

	return nil
}
