package internal

import (
	"github.com/blevesearch/bleve/v2"
	"log"
)

func NewBleve(opt *Options) bleve.Index {
	index, err := bleve.Open(opt.DataDir + "/example.bleve")
	if err != nil && err != bleve.ErrorIndexPathDoesNotExist {
		log.Fatalln(err)
	}

	if err == bleve.ErrorIndexPathDoesNotExist {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(opt.DataDir+"/example.bleve", mapping)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return index
}
