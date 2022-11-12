package internal

import (
	"database/sql"
	"embed"
	. "github.com/go-jet/jet/v2/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"ipfs-sharing/gen/model"
	. "ipfs-sharing/gen/table"
	"log"
)

var embedMigrations embed.FS

type Database struct {
	DB *sql.DB
}

func NewDatabase(opt *Options) *Database {
	var err error
	db, err := sql.Open("sqlite3", opt.DatabasePath)
	if err != nil {
		log.Fatalln(err)
	}

	//goose.SetBaseFS(embedMigrations)
	err = goose.SetDialect("sqlite3")
	if err != nil {
		log.Fatalln(err)
	}
	err = goose.Up(db, opt.MigrationsDir)
	if err != nil {
		log.Fatalln(err)
	}

	return &Database{db}
}

func (db *Database) InsertContent(content *model.Contents) error {

	insertStmt := Contents.
		INSERT(Contents.MutableColumns).
		MODEL(content).
		RETURNING(Contents.AllColumns)

	err := insertStmt.Query(db.DB, content)

	return err
}

func (db *Database) SearchContent(name string) (model.Contents, error) {

	var content model.Contents
	stmt := SELECT(Contents.AllColumns).FROM(Contents).
		WHERE(Contents.Name.LIKE(String("%" + name + "%"))).
		ORDER_BY(Contents.Name).LIMIT(1)

	err := stmt.Query(db.DB, &content)
	if err != nil {
		return content, err
	}
	// qrm.ErrNoRows
	return content, nil
}

func (db *Database) GetContentInDir(dir string) ([]model.Contents, error) {
	var contents []model.Contents
	stmt := SELECT(Contents.AllColumns).FROM(Contents).
		WHERE(Contents.Name.LIKE(String(dir + "%"))).
		ORDER_BY(Contents.Name)

	err := stmt.Query(db.DB, &contents)
	if err != nil {
		return contents, err
	}

	return contents, nil
}
