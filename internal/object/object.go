package object

import (
	"fmt"
)

type ObjType int

const (
	OBJ_STRING ObjType = iota
)

type Obj struct {
	Type ObjType
}

type ObjString struct {
	Obj
	Chars string
}

func NewObjString(s string) *ObjString {
	return &ObjString{
		Obj:   Obj{Type: OBJ_STRING},
		Chars: s,
	}
}

func TakeString(s string) *ObjString {
	return NewObjString(s)
}

func CopyString(s string) *ObjString {
	return NewObjString(s)
}

func PrintObject(obj interface{}) {
	switch o := obj.(type) {
	case *ObjString:
		fmt.Print(o.Chars)
	default:
		fmt.Print("unknown object")
	}
}
