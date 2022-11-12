package internal

import (
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"path/filepath"
)

type Options struct {
	Instance      string `short:"i" long:"instance" description:"Instance name" default:"data"`
	DataDir       string
	IpfsDir       string
	WorkDir       string
	OrbitDir      string
	ShareDir      string
	StaticDir     string
	MigrationsDir string
	DatabasePath  string
}

func NewOptions() *Options {
	var Opt Options
	_, err := flags.Parse(&Opt)
	if err != nil {
		log.Fatalln(err)
	}

	ex, err := os.Executable()
	if err != nil {
		log.Fatalln(err)
	}

	ex, err = os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	Opt.WorkDir = ex //filepath.Dir(ex)

	Opt.DataDir = filepath.Join(Opt.WorkDir, Opt.Instance)
	Opt.StaticDir = filepath.Join(Opt.WorkDir, "/static")
	Opt.MigrationsDir = filepath.Join(Opt.StaticDir, "/migrations")
	Opt.IpfsDir = filepath.Join(Opt.DataDir, "/ipfs")
	Opt.OrbitDir = filepath.Join(Opt.DataDir, "/orbit")
	Opt.ShareDir = filepath.Join(Opt.DataDir, "/share")
	Opt.DatabasePath = filepath.Join(Opt.DataDir, "database.db")
	os.MkdirAll(Opt.ShareDir, os.ModePerm)

	err = os.MkdirAll(Opt.OrbitDir, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}

	return &Opt
}
