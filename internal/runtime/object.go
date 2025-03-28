package runtime

import (
	"fmt"
	"hash/fnv"
)

type ObjType int

const (
	OBJ_FUNCTION ObjType = iota
	OBJ_NATIVE
	OBJ_STRING
)

type Obj struct {
	Type ObjType
	Next *Obj
}

type NativeFn func(argCount int, args []Value) Value

type ObjNative struct {
	Obj
	Function NativeFn
}

func NewNative(function NativeFn) *ObjNative {
	return &ObjNative{
		Function: function,
	}
}

type ObjFunction struct {
	Obj   Obj
	Arity int
	Chunk Chunk
	Name  *ObjString
}

func NewFunction() *ObjFunction {
	function := &ObjFunction{}
	function.Arity = 0
	function.Name = nil
	function.Chunk = *New()
	return function
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
	case *ObjFunction:
		if o.Name == nil {
			fmt.Print("<script>")
		} else {
			fmt.Printf("<fn %s>", o.Name.Chars)
		}
	case *ObjNative:
		fmt.Print("<native fn>")
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
