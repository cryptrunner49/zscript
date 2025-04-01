//go:build cgo
// +build cgo

package main

/*
#cgo pkg-config: readline
#include <stdio.h>
#include <stdlib.h>
#include <readline/readline.h>
#include <readline/history.h>
*/
import "C"

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/cryptrunner49/gorex/internal/vm"
)

func main() {
	vm.InitVM()

	if len(os.Args) == 1 {
		repl()
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		fmt.Fprintf(os.Stderr, "Usage: goseedvm [path]\n")
		os.Exit(1)
	}

	vm.FreeVM()
}

func repl() {
	for {
		// Create a prompt string.
		prompt := C.CString("> ")
		// Call GNU readline to get input with history/navigation.
		line := C.readline(prompt)
		// Free the prompt string.
		C.free(unsafe.Pointer(prompt))
		if line == nil {
			// EOF reached (e.g. Ctrl-D)
			fmt.Println()
			break
		}
		// Convert C string to Go string.
		input := C.GoString(line)
		// Free the line allocated by readline.
		C.free(unsafe.Pointer(line))
		// Only add non-empty lines to history.
		if len(input) > 0 {
			C.add_history(C.CString(input))
			vm.Interpret(input)
		}
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open file \"%s\": %v\n", path, err)
		os.Exit(74)
	}
	result := vm.Interpret(string(source))
	if result == vm.INTERPRET_COMPILE_ERROR {
		os.Exit(65)
	}
	if result == vm.INTERPRET_RUNTIME_ERROR {
		os.Exit(70)
	}
}
