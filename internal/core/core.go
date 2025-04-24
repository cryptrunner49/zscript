package core

import (
	"fmt"
	"os"

	"github.com/cryptrunner49/zscript/internal/vm"
)

// Interpret runs the source code in the VM and returns an exit code:
// 0 = OK, 1 = compile error, 2 = runtime error.
func Interpret(source, name string) int {
	result := vm.Interpret(source, name)
	switch result {
	case vm.INTERPRET_OK:
		return 0
	case vm.INTERPRET_COMPILE_ERROR:
		return 1
	case vm.INTERPRET_RUNTIME_ERROR:
		return 2
	default:
		fmt.Fprintf(os.Stderr, "Unknown error in '%s'\n", name)
		return 1
	}
}

// RunFile loads and runs a file, returning an exit code:
// 0 = OK, 1 = compile error, 2 = runtime error, -1 = I/O error.
func RunFile(path string) int {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", path, err)
		return -1
	}
	return Interpret(string(source), path)
}
