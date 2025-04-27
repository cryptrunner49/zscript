//go:build cgo
// +build cgo

package main

/*
#cgo pkg-config: readline
#include <stdio.h>
#include <stdlib.h>
#include <readline/readline.h>
#include <readline/history.h>
static int self_insert_wrapper(int count, int key) { return rl_tab_insert(count, key); }
static void bind_tab_key() { rl_bind_key('\t', self_insert_wrapper); }
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"unsafe"

	"github.com/cryptrunner49/zscript/internal/common"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func main() {
	// Bind the Tab key to insert a tab character instead of triggering autocomplete
	C.bind_tab_key()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "--help":
			showUsage()
			os.Exit(0)
		case "-v", "--version":
			showVersion()
			os.Exit(0)
		}
	}

	vm.InitVM(os.Args)
	defer vm.FreeVM()

	if len(os.Args) == 1 {
		fmt.Println("zvm REPL - ZScript Virtual Machine (type Ctrl+D to exit)")
		repl()
	} else {
		runFile(os.Args[1])
	}
}

// showUsage prints detailed help and usage instructions.
func showUsage() {
	usage := `zvm - A ZScript Virtual Machine Interpreter

Usage: zvm [options] [script]

Options:
  -h, --help       Display this help message and exit
  -v, --version    Show version information and exit

Modes:
  - If no script is provided, zvm starts an interactive REPL (Read-Eval-Print Loop)
    where you can type commands and execute them immediately.
  - If a script file is provided, zvm executes the script and exits.

REPL Usage:
  - Type ZScript commands at the ">>> " prompt.
  - To enter a multi-line block (e.g. for control structures), end the first line with ':'
    then continue typing the block on subsequent lines.
  - Terminate the block by entering an empty line (i.e. double Enter).
  - Use arrow keys to navigate command history.
  - Press Ctrl+D or Ctrl+C to exit.

Script Execution:
  - Provide a path to a ZScript script file to execute it.
  - Example: zvm myscript.z

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
	fmt.Printf("zvm version %s\n", common.Version)
}

// repl runs the interactive Read-Eval-Print Loop (REPL) for ZScript
func repl() {
	var blockLines []string // Accumulate multi-line input (block)
	inBlock := false        // Flag indicating if we're in a block

	for {
		// Set prompt to ">>> " for new input or "... " when continuing a multi-line block
		prompt := ">>> "
		if inBlock {
			prompt = "... "
		}
		cPrompt := C.CString(prompt)
		line := C.readline(cPrompt)
		C.free(unsafe.Pointer(cPrompt))

		if line == nil { // EOF (Ctrl+D)
			fmt.Println("\nExiting REPL")
			break
		}

		input := strings.TrimRight(C.GoString(line), "\n")
		C.free(unsafe.Pointer(line))
		trimmed := strings.TrimSpace(input)

		// In normal (non-block) mode, ignore empty lines.
		if !inBlock && trimmed == "" {
			continue
		}

		// An empty line while in a block signals the end of the block
		if inBlock && trimmed == "" {
			source := strings.Join(blockLines, "\n")
			// Add the complete block to history.
			historyEntry := C.CString(source)
			C.add_history(historyEntry)
			C.free(unsafe.Pointer(historyEntry))

			result := vm.Interpret(source, "<repl>")
			switch result {
			case vm.INTERPRET_OK:
				// Execution successful; no further output required.
			case vm.INTERPRET_COMPILE_ERROR:
				fmt.Fprintf(os.Stderr, "Compilation error in REPL\n")
			case vm.INTERPRET_RUNTIME_ERROR:
				fmt.Fprintf(os.Stderr, "Runtime error in REPL\n")
			default:
				fmt.Fprintf(os.Stderr, "Unknown error in REPL: %v\n", result)
			}
			// Reset block state.
			blockLines = []string{}
			inBlock = false
			continue
		}

		// If not in a block and the input line ends with a colon, start a new block.
		if !inBlock && strings.HasSuffix(trimmed, ":") {
			inBlock = true
			blockLines = append(blockLines, input)
			continue
		}

		// If already in a block, accumulate the line.
		if inBlock {
			blockLines = append(blockLines, input)
			continue
		}

		// Otherwise, it's a single-line command.
		// Add the line to history.
		historyEntry := C.CString(trimmed)
		C.add_history(historyEntry)
		C.free(unsafe.Pointer(historyEntry))

		result := vm.Interpret(trimmed, "<repl>")
		switch result {
		case vm.INTERPRET_OK:
			// Execution successful.
		case vm.INTERPRET_COMPILE_ERROR:
			fmt.Fprintf(os.Stderr, "Compilation error in REPL\n")
		case vm.INTERPRET_RUNTIME_ERROR:
			fmt.Fprintf(os.Stderr, "Runtime error in REPL\n")
		default:
			fmt.Fprintf(os.Stderr, "Unknown error in REPL: %v\n", result)
		}
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", path, err)
		os.Exit(74)
	}

	// Trim trailing newlines from the source code and append 'pass;'
	// This ensures that the last indented block is properly dedented,
	// conforming to the language's syntax requirements for indented blocks.
	// Without this, the interpreter may throw an error due to improper indentation.
	sourceStr := strings.TrimRight(string(source), "\n") + "\npass;\n"
	result := vm.Interpret(sourceStr, path)
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
