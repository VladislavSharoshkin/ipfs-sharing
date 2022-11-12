package models

type NamedCid struct {
	Name string
	Cid  string
}

func NewNamedCid(name string, cid string) NamedCid {
	return NamedCid{name, cid}
}
