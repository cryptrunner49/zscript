//go:build cgo
// +build cgo

package main

// #cgo CFLAGS: -I${SRCDIR}/../../../bin
// #cgo LDFLAGS: -L${SRCDIR}/../../../bin -lzscript
// #include <stdlib.h>
// #include "libzscript.h"
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

func main() {
	// Convert Go command-line arguments to C-style argc/argv
	argc := C.int(len(os.Args))
	argv := make([]*C.char, argc)
	for i, arg := range os.Args {
		argv[i] = C.CString(arg)
		defer C.free(unsafe.Pointer(argv[i]))
	}

	// Initialize ZScript
	C.ZScript_Init(argc, &argv[0])

	// Run Seed script
	if len(os.Args) > 1 {
		// Run script from file
		path := C.CString(os.Args[1])
		defer C.free(unsafe.Pointer(path))
		C.ZScript_RunFile(path)
	} else {
		// Run inline script
		source := C.CString("1 + 2;")
		name := C.CString("<test>")
		defer C.free(unsafe.Pointer(source))
		defer C.free(unsafe.Pointer(name))
		var exitCode C.int
		result := C.ZScript_InterpretWithResult(source, name, &exitCode)
		if exitCode == 0 {
			fmt.Printf("Last value: %s\n", C.GoString(result))
		} else {
			fmt.Printf("Execution failed with code %d\n", exitCode)
		}
		C.free(unsafe.Pointer(result))
	}

	// Free ZScript
	C.ZScript_Free()
}
