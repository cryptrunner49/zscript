package object

import (
	"fmt"
	"hash/fnv"
)

type ObjType int

const (
	OBJ_STRING ObjType = iota
)

type Obj struct {
	Type ObjType
	Next *Obj
}

type ObjString struct {
	Obj
	Chars string
	Hash  uint32
}

var strings = make(map[uint32]*ObjString)

func NewObjString(s string) *ObjString {
	hash := hashString(s)
	if interned, exists := strings[hash]; exists {
		return interned
	}
	objString := &ObjString{
		Obj:   Obj{Type: OBJ_STRING},
		Chars: s,
		Hash:  hash,
	}
	strings[hash] = objString
	return objString
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

func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
