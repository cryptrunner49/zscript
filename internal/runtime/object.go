package runtime

import (
	"fmt"
	"hash/fnv"
	"time"
)

// ObjType defines the types of heap-allocated objects in the runtime, used to identify object
// categories like functions, strings, and structs.
type ObjType int

// Enumeration of object types.
const (
	OBJ_UPVALUE        ObjType = iota // Upvalue: a variable captured from an outer scope.
	OBJ_CLOSURE                       // Closure: a function plus its captured environment.
	OBJ_FUNCTION                      // Function: a user-defined function.
	OBJ_NATIVE                        // Native: a built-in (native) function.
	OBJ_STRING                        // String: an immutable string.
	OBJ_STRUCT                        // Struct: a user-defined struct type.
	OBJ_INSTANCE                      // Instance: an instance of a struct.
	OBJ_ARRAY                         // Array: a dynamic array.
	OBJ_ARRAY_ITERATOR                // Array Iterator: iterator for arrays.
	OBJ_MODULE                        // Module: a module containing functions, variables and other modules.
	OBJ_MAP                           // Map: a key-value store.
	OBJ_DATE                          // Date object (year, month, day)
	OBJ_TIME                          // Time object (hour, minute, second)
	OBJ_DATETIME                      // DateTime represents a combined date and time.
)

// Obj is the header for all heap-allocated objects.
type Obj struct {
	Type ObjType // The type of the object.
	Next *Obj    // Linked list pointer for garbage collection.
}

// NativeFn is the function signature for native (built-in) functions.
type NativeFn func(argCount int, args []Value) Value

// ObjNative represents a native (built-in) function object.
type ObjNative struct {
	Obj
	Function NativeFn // Pointer to the native function implementation.
}

// ObjUpvalue represents a closed-over local variable.
type ObjUpvalue struct {
	Obj
	Location *Value      // Points to the variable's slot on the VM stack.
	Closed   Value       // Stores the closed-over value once the variable goes out of scope.
	Next     *ObjUpvalue // Linked list pointer for open upvalues.
}

// ObjClosure represents a function along with its captured upvalues.
type ObjClosure struct {
	Obj
	Function     *ObjFunction  // The function object.
	Upvalues     []*ObjUpvalue // Array of pointers to captured upvalues.
	UpvalueCount int           // Number of upvalues captured.
}

// ObjFunction represents a user-defined function.
type ObjFunction struct {
	Obj          Obj        // Object header.
	Arity        int        // Number of expected arguments.
	UpvalueCount int        // Number of upvalues the function captures.
	Chunk        Chunk      // Bytecode chunk containing the function's code.
	Name         *ObjString // Optional function name.
}

// ObjString represents an immutable string.
type ObjString struct {
	Obj
	Chars string // The actual string characters.
	Hash  uint32 // Cached hash value for quick comparisons.
}

// ObjStruct represents a struct type with named fields and default values.
type ObjStruct struct {
	Obj    Obj
	Name   *ObjString           // The name of the struct.
	Fields map[*ObjString]Value // Map of field names to their default values.
}

// ObjInstance represents an instance of a struct.
type ObjInstance struct {
	Obj       Obj
	Structure *ObjStruct           // The struct type of the instance.
	Fields    map[*ObjString]Value // Instance field values.
}

// strings is a map for interning strings, storing ObjString objects by their hash to reuse
// identical strings and reduce memory usage.
var strings = make(map[uint32]*ObjString)

// NewNative creates a new ObjNative wrapping the given native function.
func NewNative(function NativeFn) *ObjNative {
	return &ObjNative{
		Function: function,
	}
}

// NewUpvalue creates a new upvalue object pointing to a given variable location.
func NewUpvalue(location *Value) *ObjUpvalue {
	return &ObjUpvalue{
		Obj:      Obj{Type: OBJ_UPVALUE},
		Location: location,
		Closed:   Value{Type: VAL_NULL},
		Next:     nil,
	}
}

// NewClosure creates a new closure object for a given function,
// initializing its upvalue array based on the function's UpvalueCount.
func NewClosure(function *ObjFunction) *ObjClosure {
	upvalues := make([]*ObjUpvalue, function.UpvalueCount)
	return &ObjClosure{
		Obj:          Obj{Type: OBJ_CLOSURE},
		Function:     function,
		Upvalues:     upvalues,
		UpvalueCount: function.UpvalueCount,
	}
}

// NewFunction creates a new function object with an empty bytecode chunk.
func NewFunction() *ObjFunction {
	function := &ObjFunction{}
	function.Arity = 0
	function.UpvalueCount = 0
	function.Name = nil
	function.Chunk = *New()
	return function
}

// NewObjString creates (or returns an interned) ObjString for the given string.
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

// CopyString creates or returns an interned ObjString for the given string, reusing an existing
// string if it matches an interned one.
func CopyString(s string) *ObjString {
	return NewObjString(s)
}

// hashString computes a hash value for a string using the FNV-1a algorithm.
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// ObjVal wraps an object into a Value of type VAL_OBJ.
func ObjVal(obj interface{}) Value {
	return Value{Type: VAL_OBJ, Obj: obj}
}

// NewStruct creates a new struct type with the given name and an empty field map.
func NewStruct(name *ObjString) *ObjStruct {
	return &ObjStruct{
		Obj:    Obj{Type: OBJ_STRUCT},
		Name:   name,
		Fields: make(map[*ObjString]Value),
	}
}

// NewInstance creates a new instance of a struct,
// initializing its fields with the default values from the struct definition.
func NewInstance(structure *ObjStruct) *ObjInstance {
	instance := &ObjInstance{
		Obj:       Obj{Type: OBJ_INSTANCE},
		Structure: structure,
		Fields:    make(map[*ObjString]Value),
	}
	// Copy default values for each field.
	for name, value := range structure.Fields {
		instance.Fields[name] = value
	}
	return instance
}

// ObjArray represents an array object that holds a slice of values.
type ObjArray struct {
	Obj
	Elements []Value // The elements of the array.
}

// NewArray creates a new array object with the given elements.
func NewArray(elements []Value) *ObjArray {
	array := &ObjArray{
		Elements: elements,
	}
	array.Type = OBJ_ARRAY
	return array
}

// ObjArrayIterator represents an iterator for arrays.
type ObjArrayIterator struct {
	Obj
	Array *ObjArray // The array being iterated.
	Index int       // Current index in the iteration.
}

// NewArrayIterator creates a new iterator for the given array.
func NewArrayIterator(array *ObjArray) *ObjArrayIterator {
	return &ObjArrayIterator{
		Obj:   Obj{Type: OBJ_ARRAY_ITERATOR},
		Array: array,
		Index: 0,
	}
}

// ObjModule represents a module
type ObjModule struct {
	Obj    Obj
	Name   *ObjString // The name of the module.
	Fields map[*ObjString]Value
}

// NewModule creates a new module
func NewModule(name *ObjString) *ObjModule {
	return &ObjModule{
		Obj:    Obj{Type: OBJ_MODULE},
		Name:   name,
		Fields: make(map[*ObjString]Value),
	}
}

// ObjMap represents a hash map with key-value pairs.
type ObjMap struct {
	Obj
	Entries map[*ObjString]Value // Map of keys (strings) to values.
}

// NewMap creates a new empty hash map object.
func NewMap() *ObjMap {
	return &ObjMap{
		Obj:     Obj{Type: OBJ_MAP},
		Entries: make(map[*ObjString]Value),
	}
}

type ObjDate struct {
	Obj
	Time time.Time // Underlying Go time (time part ignored)
}

type ObjTime struct {
	Obj
	Time time.Time // Underlying Go time (date part ignored)
}

type ObjDateTime struct {
	Obj
	Time time.Time // Full date and time
}

// Helper functions to create instances
func NewDate(year int, month time.Month, day int) *ObjDate {
	return &ObjDate{
		Obj:  Obj{Type: OBJ_DATE},
		Time: time.Date(year, month, day, 0, 0, 0, 0, time.UTC),
	}
}

func NewTime(hour, minute, second int) *ObjTime {
	return &ObjTime{
		Obj:  Obj{Type: OBJ_TIME},
		Time: time.Date(0, 1, 1, hour, minute, second, 0, time.UTC),
	}
}

func NewDateTime(year int, month time.Month, day, hour, minute, second int) *ObjDateTime {
	return &ObjDateTime{
		Obj:  Obj{Type: OBJ_DATETIME},
		Time: time.Date(year, month, day, hour, minute, second, 0, time.UTC),
	}
}

// PrintObject prints a representation of the object to stdout.
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
		fmt.Printf("<struct %s>", o.Name.Chars)
	case *ObjInstance:
		// Print the instance as <(struct structName), field1=value1, field2=value2>, including all
		// field names and their values.
		fmt.Print("<")
		fmt.Printf("(struct %s)", o.Structure.Name.Chars) // Print struct name
		first := true
		for fieldName, fieldValue := range o.Fields {
			if !first {
				fmt.Print(",")
			}
			fmt.Printf(" %s=", fieldName.Chars) // Print field name
			PrintValue(fieldValue)              // Print field value
			first = false
		}
		fmt.Print(">")
	case *ObjArray:
		fmt.Print("[")
		for i, elem := range o.Elements {
			if i > 0 {
				fmt.Print(", ")
			}
			PrintValue(elem)
		}
		fmt.Print("]")
	case *ObjArrayIterator:
		fmt.Printf("<array iterator at %d>", o.Index)
	case *ObjModule:
		fmt.Printf("<mod %s>", o.Name.Chars)
	case *ObjMap:
		fmt.Print("{")
		first := true
		for key, value := range o.Entries {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("%s: ", key.Chars)
			PrintValue(value)
			first = false
		}
		fmt.Print("}")
	case *ObjDate:
		fmt.Printf("<Date %s>", o.Time.Format("2006-01-02"))
	case *ObjTime:
		fmt.Printf("<Time %s>", o.Time.Format("15:04:05"))
	case *ObjDateTime:
		fmt.Printf("<DateTime %s>", o.Time.Format("2006-01-02 15:04:05"))
	default:
		fmt.Print("unknown object")
	}
}
