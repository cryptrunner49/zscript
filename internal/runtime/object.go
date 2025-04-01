package runtime

import (
	"fmt"
	"hash/fnv"
)

type ObjType int

const (
	OBJ_UPVALUE ObjType = iota
	OBJ_CLOSURE
	OBJ_FUNCTION
	OBJ_NATIVE
	OBJ_STRING
	OBJ_STRUCT
	OBJ_INSTANCE
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
	UpvalueCount int
	Chunk        Chunk
	Name         *ObjString
}

type ObjString struct {
	Obj
	Chars string
	Hash  uint32
}

type ObjStruct struct {
	Obj    Obj
	Name   *ObjString
	Fields map[*ObjString]Value // Map of field names to default values
}

type ObjInstance struct {
	Obj       Obj
	Structure *ObjStruct
	Fields    map[*ObjString]Value
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
	case *ObjStruct:
		fmt.Print(o.Name.Chars)
	default:
		fmt.Print("unknown object")
	}
}

func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func ObjVal(obj interface{}) Value {
	return Value{Type: VAL_OBJ, Obj: obj}
}

func NewStruct(name *ObjString) *ObjStruct {
	return &ObjStruct{
		Obj:    Obj{Type: OBJ_STRUCT},
		Name:   name,
		Fields: make(map[*ObjString]Value),
	}
}

func NewInstance(structure *ObjStruct) *ObjInstance {
	instance := &ObjInstance{
		Obj:       Obj{Type: OBJ_INSTANCE},
		Structure: structure,
		Fields:    make(map[*ObjString]Value),
	}
	// Initialize instance fields with default values from the struct
	for name, value := range structure.Fields {
		instance.Fields[name] = value
	}
	return instance
}
