package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/cryptrunner49/zscript/internal/runtime"
	"github.com/cryptrunner49/zscript/internal/vm"
)

// ZScript_Init initializes the ZScript VM with command-line arguments.
//
//export ZScript_Init
func ZScript_Init(argc C.int, argv **C.char) {
	n := int(argc)
	var args []string
	if n > 0 && argv != nil {
		args = make([]string, n)
		slice := (*[1 << 28]*C.char)(unsafe.Pointer(argv))[:n:n]
		for i, s := range slice {
			args[i] = C.GoString(s)
		}
	} else {
		args = []string{}
	}
	vm.InitVM(args)
}

// ZScript_Interpret interprets ZScript source code with a given name.
//
//export ZScript_Interpret
func ZScript_Interpret(csrc, cname *C.char) C.int {
	src := C.GoString(csrc)
	name := C.GoString(cname)
	return C.int(vm.Interpret(src, name))
}

// ZScript_RunFile runs a ZScript script from a file path.
//
//export ZScript_RunFile
func ZScript_RunFile(cpath *C.char) C.int {
	path := C.GoString(cpath)
	source, err := os.ReadFile(path)
	if err != nil {
		return C.int(74) // File I/O error
	}
	// Normalize source and append 'pass;' to ensure a stack value
	sourceStr := strings.TrimRight(string(source), "\n") + "\npass;\n"
	return C.int(vm.Interpret(sourceStr, path))
}

// ZScript_InterpretWithResult interprets ZScript source code and returns the last value as a string.
//
//export ZScript_InterpretWithResult
func ZScript_InterpretWithResult(csrc, cname *C.char, exitCode *C.int) *C.char {
	src := C.GoString(csrc)
	name := C.GoString(cname)
	code := vm.Interpret(src, name)
	*exitCode = C.int(code)
	return valueToCString(vm.GetLastValue())
}

// ZScript_RunFileWithResult runs a ZScript script from a file and returns the last value as a string.
//
//export ZScript_RunFileWithResult
func ZScript_RunFileWithResult(cpath *C.char, exitCode *C.int) *C.char {
	path := C.GoString(cpath)
	source, err := os.ReadFile(path)
	if err != nil {
		*exitCode = C.int(74) // File I/O error
		return valueToCString(runtime.Value{Type: runtime.VAL_NULL})
	}
	// Normalize source and append 'pass;' to ensure a stack value
	sourceStr := strings.TrimRight(string(source), "\n") + "\npass;\n"
	code := vm.Interpret(sourceStr, path)
	*exitCode = C.int(code)
	return valueToCString(vm.GetLastValue())
}

// valueToString converts a runtime.Value to its string representation, mirroring runtime.PrintObject.
func valueToString(val runtime.Value) string {
	switch val.Type {
	case runtime.VAL_NULL:
		return "null"
	case runtime.VAL_BOOL:
		if val.Bool {
			return "true"
		}
		return "false"
	case runtime.VAL_NUMBER:
		return fmt.Sprintf("%g", val.Number)
	case runtime.VAL_OBJ:
		switch obj := val.Obj.(type) {
		case *runtime.ObjString:
			return obj.Chars
		case *runtime.ObjArray:
			elements := make([]string, len(obj.Elements))
			for i, elem := range obj.Elements {
				elements[i] = valueToString(elem)
			}
			return "[" + strings.Join(elements, ", ") + "]"
		case *runtime.ObjMap:
			entries := make([]string, 0, len(obj.Entries))
			for key, value := range obj.Entries {
				entries = append(entries, fmt.Sprintf("%s: %s", key.Chars, valueToString(value)))
			}
			return "{" + strings.Join(entries, ", ") + "}"
		case *runtime.ObjStruct:
			return fmt.Sprintf("<struct %s>", obj.Name.Chars)
		case *runtime.ObjInstance:
			var sb strings.Builder
			sb.WriteString("<(struct ")
			sb.WriteString(obj.Structure.Name.Chars)
			sb.WriteString(")")
			first := true
			for fieldName, fieldValue := range obj.Fields {
				if !first {
					sb.WriteString(",")
				}
				fmt.Fprintf(&sb, " %s=%s", fieldName.Chars, valueToString(fieldValue))
				first = false
			}
			sb.WriteString(">")
			return sb.String()
		case *runtime.ObjFunction:
			if obj.Name == nil {
				return "<script>"
			}
			return fmt.Sprintf("<fn %s>", obj.Name.Chars)
		case *runtime.ObjClosure:
			if obj.Function.Name == nil {
				return "<script>"
			}
			return fmt.Sprintf("<fn %s>", obj.Function.Name.Chars)
		case *runtime.ObjNative:
			return "<native fn>"
		case *runtime.ObjModule:
			return fmt.Sprintf("<mod %s>", obj.Name.Chars)
		case *runtime.ObjDate:
			return fmt.Sprintf("<Date %s>", obj.Time.Format("2006-01-02"))
		case *runtime.ObjTime:
			return fmt.Sprintf("<Time %s>", obj.Time.Format("15:04:05"))
		case *runtime.ObjDateTime:
			return fmt.Sprintf("<DateTime %s>", obj.Time.Format("2006-01-02 15:04:05"))
		case *runtime.ObjArrayIterator:
			return fmt.Sprintf("<array iterator at %d>", obj.Index)
		default:
			return "<unknown object>"
		}
	default:
		return "<unknown>"
	}
}

// valueToCString converts a runtime.Value to a C string, which must be freed by the caller.
func valueToCString(val runtime.Value) *C.char {
	return C.CString(valueToString(val))
}

// ZScript_Free frees the ZScript VM resources.
//
//export ZScript_Free
func ZScript_Free() {
	vm.FreeVM()
}

func main() {}
