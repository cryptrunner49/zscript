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
static int self_insert_wrapper(int count, int key) {
    return rl_tab_insert(count, key);
}

// Bind the tab key to self_insert_wrapper.
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
	usage := `spyvm - A SPYScript Virtual Machine Interpreter

Usage: spyvm [options] [script]

Options:
  -h, --help       Display this help message and exit
  -v, --version    Show version information and exit

Modes:
  - If no script is provided, spyvm starts an interactive REPL (Read-Eval-Print Loop)
    where you can type commands and execute them immediately.
  - If a script file is provided, spyvm executes the script and exits.

REPL Usage:
  - Type SPYScript commands at the ">>> " prompt.
  - To enter a multi-line block (e.g. for control structures), end the first line with ':'
    then continue typing the block on subsequent lines.
  - Terminate the block by entering an empty line (i.e. double Enter).
  - Use arrow keys to navigate command history.
  - Press Ctrl+D or Ctrl+C to exit.

Script Execution:
  - Provide a path to a SPYScript script file to execute it.
  - Example: spyvm myscript.spy

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
	fmt.Printf("spyvm version %s\n", common.Version)
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
		fmt.Println("spyvm REPL - SPYScript Virtual Machine (type Ctrl+D to exit)")
		repl()
	} else {
		runFile(os.Args[1])
	}
}

func repl() {
	var blockLines []string // Accumulate multi-line input (block)
	inBlock := false        // Flag indicating if we're in a block

	for {
		// Use ">>> " when not in a block, otherwise the simple "... " prompt
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

		// If we are in a block and get an empty line, that signals the end of the block.
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
