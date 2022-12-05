package internal

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/go-jet/jet/v2/qrm"
	. "github.com/go-jet/jet/v2/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"ipfs-sharing/gen/model"
	. "ipfs-sharing/gen/table"
	"ipfs-sharing/models"
	"log"
)

var embedMigrations embed.FS

type Database struct {
	Conn *sql.DB
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

	ddl := `
        PRAGMA journal_mode = OFF;
		PRAGMA synchronous = 0;
		PRAGMA cache_size = 1000000;
-- 		PRAGMA locking_mode = EXCLUSIVE;
		PRAGMA temp_store = MEMORY;
    `

	_, err = db.Exec(ddl)
	if err != nil {
		return nil
	}

	return &Database{db}
}

func (db *Database) SmtpByDirAndName(name string, dir string) SelectStatement {
	smtp := Contents.SELECT(Bool(false)).
		WHERE(Contents.Name.EQ(String(name)).AND(Contents.Dir.EQ(String(dir))))

	return smtp
}

func (db *Database) SmtpUnfinishedDownloads() SelectStatement {
	smtp := Contents.SELECT(Contents.AllColumns).
		WHERE(Contents.Status.EQ(String(models.ContentStatusDownloading)))

	return smtp
}

func (db *Database) IsExist(smtp SelectStatement) (isExist bool, err error) {
	err = smtp.Query(db.Conn, &model.Contents{})
	if err != nil && err != qrm.ErrNoRows {
		return
	}
	if err == nil {
		isExist = true
	}

	return isExist, nil
}

func (db *Database) InsertContent(content *model.Contents) error {

	insertStmt := Contents.
		INSERT(Contents.MutableColumns).
		MODEL(content).
		RETURNING(Contents.AllColumns)

	err := insertStmt.Query(db.Conn, content)

	return err
}

func (db *Database) Query(stmt Statement, destination interface{}) error {
	return stmt.Query(db.Conn, destination)
}

func (db *Database) Save(data interface{}) error {

	var allColumns ColumnList
	var mutableColumns ColumnList
	var table Table
	insert := true
	var id int32

	switch v := data.(type) {
	case *model.Contents:
		table = Contents
		allColumns = Contents.AllColumns
		mutableColumns = Contents.MutableColumns
		id = v.ID
	case *model.Messages:
		table = Messages
		allColumns = Messages.AllColumns
		mutableColumns = Messages.MutableColumns
		id = v.ID
	default:
		log.Fatalln("I don't know about type %T!\n", v)
	}

	if id != 0 {
		insert = false
	}

	var stmt Statement
	stmt = table.UPDATE(mutableColumns).MODEL(data).RETURNING(allColumns).WHERE(RawInt("id").EQ(Int32(id)))
	if insert {
		stmt = table.INSERT(mutableColumns).MODEL(data).RETURNING(allColumns)
	}

	err := stmt.Query(db.Conn, data)

	return err
}

func (db *Database) SearchContent(name string) (model.Contents, error) {

	var content model.Contents
	stmt := SELECT(Contents.AllColumns).FROM(Contents).
		WHERE(Contents.Name.LIKE(String("%" + name + "%"))).
		ORDER_BY(Contents.Name).LIMIT(1)

	err := stmt.Query(db.Conn, &content)
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

	err := stmt.Query(db.Conn, &contents)
	if err != nil {
		return contents, err
	}

	return contents, nil
}

func (db *Database) Delete(table Table, id int32) error {
	stmt := table.DELETE().WHERE(RawInt("id").EQ(Int32(id)))

	_, err := stmt.Exec(db.Conn)
	return err
}

func (db *Database) DeleteMany(table Table, ids []int32) error {
	stmt := table.DELETE().WHERE(RawInt("id").IN(InInt(ids)...))

	_, err := stmt.Exec(db.Conn)
	return err
}

func (db *Database) Count(table Table) (int32, error) {
	var count int32
	stmt := table.SELECT(COUNT(RawInt("id")))

	err := stmt.Query(db.Conn, &count)
	return count, err
}

func InInt(ids []int32) []Expression {
	var ids2 []Expression
	for _, id := range ids {
		ids2 = append(ids2, Int32(id))
	}
	return ids2
}

func (db *Database) GetMessages(address string) (messages []model.Messages, err error) {
	stmt := Messages.SELECT(Messages.AllColumns).
		WHERE(Messages.From.EQ(String(address)).
			OR(Messages.To.EQ(String(address))))

	err = stmt.Query(db.Conn, &messages)
	return
}

func (db *Database) GetContacts() (contacts []string, err error) {
	stmt := Messages.SELECT(Messages.To).GROUP_BY(Messages.To)

	var messages []model.Messages
	err = stmt.Query(db.Conn, &messages)

	for _, message := range messages {
		contacts = append(contacts, message.To)
	}

	return
}

func (db *Database) ByID(id int32, data interface{}) error {
	var allColumns ColumnList
	var table Table

	switch v := data.(type) {
	case *model.Contents:
		table = Contents
		allColumns = Contents.AllColumns
	case *model.Messages:
		table = Messages
		allColumns = Messages.AllColumns
	default:
		log.Fatalln("I don't know about type %T!\n", v)
	}

	insertStmt := table.SELECT(allColumns).WHERE(RawInt("id").EQ(Int32(id)))

	err := insertStmt.Query(db.Conn, data)

	return err
}

func (db *Database) ByCid(contentCid string, date model.Contents) (err error) {
	stmt := Contents.SELECT(Contents.AllColumns).
		WHERE(Contents.Cid.EQ(String(contentCid))).
		LIMIT(1)

	exec, err := stmt.Exec(db.Conn)
	if err != nil {
		return
	}
	affected, err := exec.RowsAffected()
	if err != nil {
		return
	}
	fmt.Println(affected)

	return
}

func (db *Database) FirstByCid(contentCid string) SelectStatement {
	stmt := Contents.SELECT(Contents.AllColumns).
		WHERE(Contents.Cid.EQ(String(contentCid))).
		LIMIT(1)

	return stmt
}

func (db *Database) GetContentsDependencies(id int32) (contents []model.Contents, err error) {
	subordinates := CTE("subordinates")
	subordinates2 := CTE("subordinates2")

	stmt := WITH_RECURSIVE(
		subordinates.AS(
			SELECT(
				Contents.AllColumns,
			).FROM(
				Contents,
			).WHERE(
				Contents.ID.EQ(Int32(id)),
			).UNION(
				SELECT(
					Contents.AllColumns,
				).FROM(
					Contents.
						INNER_JOIN(subordinates, Contents.ID.From(subordinates).EQ(Contents.ParentID)),
				),
			),
		),
		subordinates2.AS(
			SELECT(
				Contents.AllColumns,
			).FROM(
				Contents,
			).WHERE(
				Contents.ID.EQ(Int32(id)),
			).UNION(
				SELECT(
					Contents.AllColumns,
				).FROM(
					Contents.
						INNER_JOIN(subordinates2, Contents.ParentID.From(subordinates2).EQ(Contents.ID)),
				),
			),
		),
	)(
		SELECT(
			subordinates.AllColumns(),
		).FROM(
			subordinates,
		).UNION(SELECT(
			subordinates2.AllColumns(),
		).FROM(
			subordinates2,
		)),
	)

	err = stmt.Query(db.Conn, &contents)
	return
}
