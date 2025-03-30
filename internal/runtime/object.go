package runtime

import (
	"fmt"
	"hash/fnv"
)

type ObjType int

const (
	OBJ_UPVALUE ObjType = iota // Added for upvalues
	OBJ_CLOSURE                // Added for closures
	OBJ_FUNCTION
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

type ObjUpvalue struct {
	Obj
	Location *Value
	Closed   Value
	Next     *ObjUpvalue
}

type ObjClosure struct {
	Obj
	Function     *ObjFunction
	Upvalues     []*ObjUpvalue
	UpvalueCount int
}

type ObjFunction struct {
	Obj          Obj
	Arity        int
	UpvalueCount int // Added for closures
	Chunk        Chunk
	Name         *ObjString
}

type ObjString struct {
	Obj
	Chars string
	Hash  uint32
}

var strings = make(map[uint32]*ObjString)

func NewNative(function NativeFn) *ObjNative {
	return &ObjNative{
		Function: function,
	}
}

func NewUpvalue(location *Value) *ObjUpvalue {
	return &ObjUpvalue{
		Obj:      Obj{Type: OBJ_UPVALUE},
		Location: location,
		Closed:   Value{Type: VAL_NULL},
		Next:     nil,
	}
}

func NewClosure(function *ObjFunction) *ObjClosure {
	upvalues := make([]*ObjUpvalue, function.UpvalueCount)
	return &ObjClosure{
		Obj:          Obj{Type: OBJ_CLOSURE},
		Function:     function,
		Upvalues:     upvalues,
		UpvalueCount: function.UpvalueCount,
	}
}

func NewFunction() *ObjFunction {
	function := &ObjFunction{}
	function.Arity = 0
	function.UpvalueCount = 0
	function.Name = nil
	function.Chunk = *New()
	return function
}

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
	case *ObjClosure:
		if o.Function.Name == nil {
			fmt.Print("<script>")
		} else {
			fmt.Printf("<fn %s>", o.Function.Name.Chars)
		}
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
