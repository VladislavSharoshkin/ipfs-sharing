//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

type Contents struct {
	ID        int32 `sql:"primary_key"`
	Name      string
	Cid       string
	ParentID  *int32
	Status    string
	From      string
	Dir       string
	CreatedAt string
}
