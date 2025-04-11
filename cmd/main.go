//go:build cgo
// +build cgo

package main

/*
#cgo pkg-config: readline
#include <stdio.h>
#include <stdlib.h>
#include <readline/readline.h>
#include <readline/history.h>

// Helper wrapper that inserts a literal tab character.
// Using rl_tab_insert as an alternative.
static int self_insert_wrapper(int count, int key) {
    return rl_tab_insert(count, key);
}

// Bind the tab key to self_insert_wrapper.
// This function performs the binding completely in C so that pointer issues are avoided.
static void bind_tab_key() {
    rl_bind_key('\t', self_insert_wrapper);
}
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/cryptrunner49/spy/internal/common"
	"github.com/cryptrunner49/spy/internal/vm"
)

// showUsage prints detailed help and usage instructions.
func showUsage() {
	usage := `goseedvm - A Seed Virtual Machine Interpreter

Usage: goseedvm [options] [script]

Options:
  -h, --help       Display this help message and exit
  -v, --version    Show version information and exit

Modes:
  - If no script is provided, goseedvm starts an interactive REPL (Read-Eval-Print Loop)
    where you can type commands and execute them immediately.
  - If a script file is provided, goseedvm executes the script and exits.

REPL Usage:
  - Type Seed VM commands at the ">>> " prompt
  - For multi-line constructs (like if-statements or loops), continue typing on the 
    next line — you'll see the "... " prompt until all opening braces '{' are matched
  - Use arrow keys to navigate command history
  - Press Ctrl+D or Ctrl+C to exit

Script Execution:
  - Provide a path to a Seed VM script file to execute it
  - Example: goseedvm myscript.seed

Exit Codes:
  0   Successful execution
  65  Compilation error
  70  Runtime error
  74  File I/O error
`
	fmt.Print(usage)
}

// showVersion prints version information.
func showVersion() {
	fmt.Printf("goseedvm version %s\n", common.Version)
}

func main() {
	// Bind the tab key so that it inserts a literal tab instead of invoking completion.
	C.bind_tab_key()

	// Handle command-line arguments
	if len(os.Args) > 1 {
		switch arg := os.Args[1]; arg {
		case "-h", "--help":
			showUsage()
			os.Exit(0)
		case "-v", "--version":
			showVersion()
			os.Exit(0)
		}
	}

	// Initialize VM with arguments
	vm.InitVM(os.Args)
	defer vm.FreeVM()

	// Determine mode: REPL or script execution
	if len(os.Args) == 1 {
		fmt.Println("goseedvm REPL - Seed Virtual Machine (type Ctrl+D to exit)")
		repl()
	} else {
		runFile(os.Args[1])
	}
}

// countBlocks counts the net number of open blocks (unmatched '{' minus '}')
func countBlocks(input string) int {
	count := 0
	for _, char := range input {
		if char == '{' {
			count++
		} else if char == '}' {
			count--
		}
	}
	return count
}

func repl() {
	var buffer strings.Builder // Accumulate multi-line input
	blockDepth := 0            // Track open blocks

	for {
		// Use ">>> " when not in a block, otherwise the simple "... " prompt
		prompt := ">>> "
		if blockDepth > 0 {
			prompt = "... "
		}
		cPrompt := C.CString(prompt)
		line := C.readline(cPrompt)
		C.free(unsafe.Pointer(cPrompt))

		if line == nil { // EOF (Ctrl+D)
			fmt.Println("\nExiting REPL")
			break
		}

		input := strings.TrimSpace(C.GoString(line))
		C.free(unsafe.Pointer(line))

		if len(input) == 0 && blockDepth == 0 {
			continue // Skip empty lines unless in a block
		}

		// Add input to the buffer with a newline when necessary.
		if buffer.Len() > 0 {
			buffer.WriteString("\n")
		}
		buffer.WriteString(input)

		// Update block depth based on the input.
		blockDepth += countBlocks(input)

		if blockDepth < 0 {
			fmt.Fprintf(os.Stderr, "REPL error: Unmatched closing brace '}'\n")
			buffer.Reset()
			blockDepth = 0
			continue
		}

		// Once all blocks are closed, interpret the accumulated input.
		if blockDepth == 0 {
			source := buffer.String()

			// Add input to history only when the block is complete.
			historyEntry := C.CString(source)
			C.add_history(historyEntry)
			C.free(unsafe.Pointer(historyEntry))

			result := vm.Interpret(source, "<repl>")
			switch result {
			case vm.INTERPRET_OK:
				// Successful execution, no output needed.
			case vm.INTERPRET_COMPILE_ERROR:
				fmt.Fprintf(os.Stderr, "Compilation error in REPL\n")
			case vm.INTERPRET_RUNTIME_ERROR:
				fmt.Fprintf(os.Stderr, "Runtime error in REPL\n")
			default:
				fmt.Fprintf(os.Stderr, "Unknown error in REPL: %v\n", result)
			}

			// Reset the buffer after interpreting.
			buffer.Reset()
		}
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", path, err)
		os.Exit(74)
	}

	result := vm.Interpret(string(source), path)
	switch result {
	case vm.INTERPRET_OK:
		// Successful execution, exit silently.
	case vm.INTERPRET_COMPILE_ERROR:
		fmt.Fprintf(os.Stderr, "Compilation error in '%s'\n", path)
		os.Exit(65)
	case vm.INTERPRET_RUNTIME_ERROR:
		fmt.Fprintf(os.Stderr, "Runtime error in '%s'\n", path)
		os.Exit(70)
	default:
		fmt.Fprintf(os.Stderr, "Unknown error: %v\n", result)
		os.Exit(1)
	}
}
