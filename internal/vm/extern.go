//go:build cgo
// +build cgo

package vm

/*
#cgo pkg-config: libffi
#cgo LDFLAGS: -ldl
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>
#include <string.h>
#include <complex.h>
#include <inttypes.h>
#include <dlfcn.h>
#include <ffi.h>

// ReturnValue is a union that stores the return value of a C function call
typedef union {
    void* ptr;
    int8_t i8;
    uint8_t u8;
    int16_t i16;
    uint16_t u16;
    int32_t i32;
    uint32_t u32;
    int64_t i64;
    uint64_t u64;
    float f;
    double d;
    float _Complex fc;
    double _Complex dc;
    _Bool b;
    char c;
    unsigned char uc;
    signed char sc;
    intptr_t iptr;
    uintptr_t uptr;
    intmax_t imax;
    uintmax_t umax;
    size_t size;
} ReturnValue;

// TypeCode is an enumeration of type codes representing C data types, used to map argument and
// return types for libffi function calls.
typedef enum {
    TYPE_VOID = 0,
    TYPE_INT8,
    TYPE_UINT8,
    TYPE_INT16,
    TYPE_UINT16,
    TYPE_INT32,
    TYPE_UINT32,
    TYPE_INT64,
    TYPE_UINT64,
    TYPE_FLOAT,
    TYPE_DOUBLE,
    TYPE_FLOAT_COMPLEX,
    TYPE_DOUBLE_COMPLEX,
    TYPE_BOOL,
    TYPE_CHAR,
    TYPE_UCHAR,
    TYPE_SCHAR,
    TYPE_INTPTR,
    TYPE_UINTPTR,
    TYPE_INTMAX,
    TYPE_UINTMAX,
    TYPE_SIZE,
    TYPE_PTR
} TypeCode;

// Structure to hold an argument
typedef struct {
    TypeCode argType;
    union {
        void* ptr;
        int8_t i8;
        uint8_t u8;
        int16_t i16;
        uint16_t u16;
        int32_t i32;
        uint32_t u32;
        int64_t i64;
        uint64_t u64;
        float f;
        double d;
        float _Complex fc;
        double _Complex dc;
        _Bool b;
        char c;
        unsigned char uc;
        signed char sc;
        intptr_t iptr;
        uintptr_t uptr;
        intmax_t imax;
        uintmax_t umax;
        size_t size;
    } value;
} Argument;

// get_ffi_type maps a TypeCode to the corresponding libffi type, enabling proper type handling
// for function arguments and return values.
ffi_type* get_ffi_type(TypeCode type) {
    switch (type) {
        case TYPE_VOID: return &ffi_type_void;
        case TYPE_INT8: return &ffi_type_sint8;
        case TYPE_UINT8: return &ffi_type_uint8;
        case TYPE_INT16: return &ffi_type_sint16;
        case TYPE_UINT16: return &ffi_type_uint16;
        case TYPE_INT32: return &ffi_type_sint32;
        case TYPE_UINT32: return &ffi_type_uint32;
        case TYPE_INT64: return &ffi_type_sint64;
        case TYPE_UINT64: return &ffi_type_uint64;
        case TYPE_FLOAT: return &ffi_type_float;
        case TYPE_DOUBLE: return &ffi_type_double;
        case TYPE_BOOL: return &ffi_type_uint8; // _Bool typically maps to uint8
        case TYPE_CHAR: return &ffi_type_sint8;
        case TYPE_UCHAR: return &ffi_type_uint8;
        case TYPE_SCHAR: return &ffi_type_sint8;
        case TYPE_INTPTR: return sizeof(intptr_t) == 8 ? &ffi_type_sint64 : &ffi_type_sint32;
        case TYPE_UINTPTR: return sizeof(uintptr_t) == 8 ? &ffi_type_uint64 : &ffi_type_uint32;
        case TYPE_INTMAX: return &ffi_type_sint64;
        case TYPE_UINTMAX: return &ffi_type_uint64;
        case TYPE_SIZE: return &ffi_type_uint64;
        case TYPE_PTR: return &ffi_type_pointer;
        // Map complex types to their real component types as placeholders, since libffi does not
		// directly support complex numbers.
        case TYPE_FLOAT_COMPLEX: return &ffi_type_float;
        case TYPE_DOUBLE_COMPLEX: return &ffi_type_double;
        default: return &ffi_type_pointer; // Fallback to pointer
    }
}

// Generic function to call any C function using libffi
ReturnValue call_function(void* func, TypeCode return_type, int arg_count, Argument* args) {
    ffi_cif cif;
    ffi_type* rtype = get_ffi_type(return_type);
    ffi_type** atypes = NULL;
    void** avalues = NULL;
    ReturnValue ret = {0};

    if (arg_count > 0) {
        atypes = (ffi_type**)malloc(arg_count * sizeof(ffi_type*));
        avalues = (void**)malloc(arg_count * sizeof(void*));
        if (!atypes || !avalues) {
            fprintf(stderr, "Memory allocation failed\n");
            free(atypes);
            free(avalues);
            return ret;
        }

        for (int i = 0; i < arg_count; i++) {
            atypes[i] = get_ffi_type(args[i].argType);
            switch (args[i].argType) {
                case TYPE_INT8: avalues[i] = &args[i].value.i8; break;
                case TYPE_UINT8: avalues[i] = &args[i].value.u8; break;
                case TYPE_INT16: avalues[i] = &args[i].value.i16; break;
                case TYPE_UINT16: avalues[i] = &args[i].value.u16; break;
                case TYPE_INT32: avalues[i] = &args[i].value.i32; break;
                case TYPE_UINT32: avalues[i] = &args[i].value.u32; break;
                case TYPE_INT64: avalues[i] = &args[i].value.i64; break;
                case TYPE_UINT64: avalues[i] = &args[i].value.u64; break;
                case TYPE_FLOAT: avalues[i] = &args[i].value.f; break;
                case TYPE_DOUBLE: avalues[i] = &args[i].value.d; break;
                case TYPE_BOOL: avalues[i] = &args[i].value.b; break;
                case TYPE_CHAR: avalues[i] = &args[i].value.c; break;
                case TYPE_UCHAR: avalues[i] = &args[i].value.uc; break;
                case TYPE_SCHAR: avalues[i] = &args[i].value.sc; break;
                case TYPE_INTPTR: avalues[i] = &args[i].value.iptr; break;
                case TYPE_UINTPTR: avalues[i] = &args[i].value.uptr; break;
                case TYPE_INTMAX: avalues[i] = &args[i].value.imax; break;
                case TYPE_UINTMAX: avalues[i] = &args[i].value.umax; break;
                case TYPE_SIZE: avalues[i] = &args[i].value.size; break;
                case TYPE_PTR: avalues[i] = &args[i].value.ptr; break;
                case TYPE_FLOAT_COMPLEX: avalues[i] = &args[i].value.fc; break;
                case TYPE_DOUBLE_COMPLEX: avalues[i] = &args[i].value.dc; break;
                default: avalues[i] = &args[i].value.ptr; break;
            }
        }
    }

    if (ffi_prep_cif(&cif, FFI_DEFAULT_ABI, arg_count, rtype, atypes) == FFI_OK) {
        ffi_call(&cif, func, &ret, avalues);
    } else {
        fprintf(stderr, "Failed to prepare CIF for function call\n");
    }

    free(atypes);
    free(avalues);
    return ret;
}
*/
import "C"

import (
	"strings"
	"unsafe"

	"github.com/cryptrunner49/zscript/internal/runtime"
)

// typeToCode maps ZScript type names to C TypeCode values, enabling conversion of script
// types to C types for foreign function interface (FFI) calls.
var typeToCode = map[string]C.TypeCode{
	"void":            C.TYPE_VOID,
	"int8_t":          C.TYPE_INT8,
	"uint8_t":         C.TYPE_UINT8,
	"int16_t":         C.TYPE_INT16,
	"uint16_t":        C.TYPE_UINT16,
	"int32_t":         C.TYPE_INT32,
	"uint32_t":        C.TYPE_UINT32,
	"int64_t":         C.TYPE_INT64,
	"uint64_t":        C.TYPE_UINT64,
	"float":           C.TYPE_FLOAT,
	"double":          C.TYPE_DOUBLE,
	"float _Complex":  C.TYPE_FLOAT_COMPLEX,
	"double _Complex": C.TYPE_DOUBLE_COMPLEX,
	"bool":            C.TYPE_BOOL,
	"char":            C.TYPE_CHAR,
	"unsigned char":   C.TYPE_UCHAR,
	"signed char":     C.TYPE_SCHAR,
	"intptr_t":        C.TYPE_INTPTR,
	"uintptr_t":       C.TYPE_UINTPTR,
	"intmax_t":        C.TYPE_INTMAX,
	"uintmax_t":       C.TYPE_UINTMAX,
	"size_t":          C.TYPE_SIZE,
	"void*":           C.TYPE_PTR,
	"char*":           C.TYPE_PTR,
	"int*":            C.TYPE_PTR,
}

// createNativeFunc creates an ObjNative wrapper for a C function, converting TulipScript
// arguments and return values to C types using libffi, and handling type validation and errors.
func createNativeFunc(funcName string, cFunc unsafe.Pointer, returnType string, paramTypes []string) *runtime.ObjNative {
	return &runtime.ObjNative{
		Function: func(argCount int, args []runtime.Value) runtime.Value {
			if argCount != len(paramTypes) {
				runtimeError("Function '%s' expects %d arguments but got %d.", funcName, len(paramTypes), argCount)
				return runtime.Value{Type: runtime.VAL_NULL}
			}

			// Map parameter types to TypeCode
			cParamTypes := make([]C.TypeCode, len(paramTypes))
			for i, pt := range paramTypes {
				code, ok := typeToCode[pt]
				if !ok {
					runtimeError("Unsupported parameter type '%s' for '%s'.", pt, funcName)
					return runtime.Value{Type: runtime.VAL_NULL}
				}
				cParamTypes[i] = code
			}

			// Map return type to TypeCode
			cReturnType, ok := typeToCode[returnType]
			if !ok {
				runtimeError("Unsupported return type '%s' for '%s'.", returnType, funcName)
				return runtime.Value{Type: runtime.VAL_NULL}
			}

			// Convert ZScript arguments to C arguments
			cArgs := make([]C.Argument, argCount)
			var cStrings []unsafe.Pointer // Track allocated C strings
			for i, pt := range paramTypes {
				cArgs[i].argType = cParamTypes[i]
				switch pt {
				case "int8_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int8)(unsafe.Pointer(&cArgs[i].value[0])) = int8(args[i].Number)
				case "uint8_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint8)(unsafe.Pointer(&cArgs[i].value[0])) = uint8(args[i].Number)
				case "int16_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int16)(unsafe.Pointer(&cArgs[i].value[0])) = int16(args[i].Number)
				case "uint16_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint16)(unsafe.Pointer(&cArgs[i].value[0])) = uint16(args[i].Number)
				case "int32_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int32)(unsafe.Pointer(&cArgs[i].value[0])) = int32(args[i].Number)
				case "uint32_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint32)(unsafe.Pointer(&cArgs[i].value[0])) = uint32(args[i].Number)
				case "int64_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int64)(unsafe.Pointer(&cArgs[i].value[0])) = int64(args[i].Number)
				case "uint64_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint64)(unsafe.Pointer(&cArgs[i].value[0])) = uint64(args[i].Number)
				case "float":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*float32)(unsafe.Pointer(&cArgs[i].value[0])) = float32(args[i].Number)
				case "double":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*float64)(unsafe.Pointer(&cArgs[i].value[0])) = args[i].Number
				case "float _Complex":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number (for real part).", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*float32)(unsafe.Pointer(&cArgs[i].value[0])) = float32(args[i].Number) // Real part only
				case "double _Complex":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number (for real part).", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*float64)(unsafe.Pointer(&cArgs[i].value[0])) = args[i].Number // Real part only
				case "bool":
					if args[i].Type != runtime.VAL_BOOL {
						runtimeError("Argument %d of '%s' must be a boolean.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*bool)(unsafe.Pointer(&cArgs[i].value[0])) = args[i].Bool
				case "char":
					if args[i].Type != runtime.VAL_OBJ {
						runtimeError("Argument %d of '%s' must be a string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					objString, ok := args[i].Obj.(*runtime.ObjString)
					if !ok {
						runtimeError("Argument %d of '%s' must be a string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					s := objString.Chars
					if len(s) != 1 {
						runtimeError("Argument %d of '%s' must be a single-character string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int8)(unsafe.Pointer(&cArgs[i].value[0])) = int8(s[0])
				case "unsigned char":
					if args[i].Type != runtime.VAL_OBJ {
						runtimeError("Argument %d of '%s' must be a string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					objString, ok := args[i].Obj.(*runtime.ObjString)
					if !ok {
						runtimeError("Argument %d of '%s' must be a string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					s := objString.Chars
					if len(s) != 1 {
						runtimeError("Argument %d of '%s' must be a single-character string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint8)(unsafe.Pointer(&cArgs[i].value[0])) = uint8(s[0])
				case "signed char":
					if args[i].Type != runtime.VAL_OBJ {
						runtimeError("Argument %d of '%s' must be a string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					objString, ok := args[i].Obj.(*runtime.ObjString)
					if !ok {
						runtimeError("Argument %d of '%s' must be a string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					s := objString.Chars
					if len(s) != 1 {
						runtimeError("Argument %d of '%s' must be a single-character string.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int8)(unsafe.Pointer(&cArgs[i].value[0])) = int8(s[0])
				case "intptr_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int)(unsafe.Pointer(&cArgs[i].value[0])) = int(args[i].Number)
				case "uintptr_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint)(unsafe.Pointer(&cArgs[i].value[0])) = uint(args[i].Number)
				case "intmax_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*int64)(unsafe.Pointer(&cArgs[i].value[0])) = int64(args[i].Number)
				case "uintmax_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint64)(unsafe.Pointer(&cArgs[i].value[0])) = uint64(args[i].Number)
				case "size_t":
					if args[i].Type != runtime.VAL_NUMBER {
						runtimeError("Argument %d of '%s' must be a number.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
					*(*uint64)(unsafe.Pointer(&cArgs[i].value[0])) = uint64(args[i].Number)
				case "char*":
					if args[i].Type == runtime.VAL_NULL {
						*(*unsafe.Pointer)(unsafe.Pointer(&cArgs[i].value[0])) = nil
					} else if args[i].Type == runtime.VAL_OBJ {
						objString, ok := args[i].Obj.(*runtime.ObjString)
						if !ok {
							runtimeError("Argument %d of '%s' must be null or a string for 'char*'.", i+1, funcName)
							return runtime.Value{Type: runtime.VAL_NULL}
						}
						cStr := C.CString(objString.Chars)
						*(*unsafe.Pointer)(unsafe.Pointer(&cArgs[i].value[0])) = unsafe.Pointer(cStr)
						cStrings = append(cStrings, unsafe.Pointer(cStr))
					} else {
						runtimeError("Argument %d of '%s' must be null or a string for 'char*'.", i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
				default:
					if strings.HasSuffix(pt, "*") {
						if args[i].Type != runtime.VAL_NULL {
							runtimeError("Argument %d of '%s' must be null for pointer type '%s'.", i+1, funcName, pt)
							return runtime.Value{Type: runtime.VAL_NULL}
						}
						*(*unsafe.Pointer)(unsafe.Pointer(&cArgs[i].value[0])) = nil
					} else {
						runtimeError("Unsupported parameter type '%s' for argument %d of '%s'.", pt, i+1, funcName)
						return runtime.Value{Type: runtime.VAL_NULL}
					}
				}
			}

			// Call the C function using libffi
			var ret C.ReturnValue
			if argCount > 0 {
				ret = C.call_function(cFunc, cReturnType, C.int(argCount), &cArgs[0])
			} else {
				ret = C.call_function(cFunc, cReturnType, 0, nil)
			}

			// Free allocated C strings
			for _, cStr := range cStrings {
				C.free(cStr)
			}

			// Convert return value back to ZScript
			switch cReturnType {
			case C.TYPE_VOID:
				return runtime.Value{Type: runtime.VAL_NULL}
			case C.TYPE_INT8:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*int8)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_UINT8:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint8)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_INT16:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*int16)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_UINT16:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint16)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_INT32:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*int32)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_UINT32:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint32)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_INT64:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*int64)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_UINT64:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint64)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_FLOAT:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*float32)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_DOUBLE:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: *(*float64)(unsafe.Pointer(&ret[0]))}
			case C.TYPE_FLOAT_COMPLEX:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*float32)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_DOUBLE_COMPLEX:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: *(*float64)(unsafe.Pointer(&ret[0]))}
			case C.TYPE_BOOL:
				return runtime.Value{Type: runtime.VAL_BOOL, Bool: *(*bool)(unsafe.Pointer(&ret[0]))}
			case C.TYPE_CHAR:
				return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(string(rune(*(*int8)(unsafe.Pointer(&ret[0])))))}
			case C.TYPE_UCHAR:
				return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(string(rune(*(*uint8)(unsafe.Pointer(&ret[0])))))}
			case C.TYPE_SCHAR:
				return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(string(rune(*(*int8)(unsafe.Pointer(&ret[0])))))}
			case C.TYPE_INTPTR:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*int)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_UINTPTR:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_INTMAX:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*int64)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_UINTMAX:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint64)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_SIZE:
				return runtime.Value{Type: runtime.VAL_NUMBER, Number: float64(*(*uint64)(unsafe.Pointer(&ret[0])))}
			case C.TYPE_PTR:
				ptr := *(*unsafe.Pointer)(unsafe.Pointer(&ret[0]))
				if ptr == nil {
					return runtime.Value{Type: runtime.VAL_NULL}
				}
				if returnType == "char*" {
					return runtime.Value{Type: runtime.VAL_OBJ, Obj: runtime.NewObjString(C.GoString((*C.char)(ptr)))}
				}
				// For non-char* pointers, return a null value since we don’t have an opaque type
				runtimeError("Non-char* pointer return type '%s' not fully supported; returning null.", returnType)
				return runtime.Value{Type: runtime.VAL_NULL}
			default:
				runtimeError("Unexpected return type code %d for '%s'.", cReturnType, funcName)
				return runtime.Value{Type: runtime.VAL_NULL}
			}
		},
	}
}
