package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/cryptrunner49/gorex/internal/vm"
)

func main() {
	vm.InitVM()

	if len(os.Args) == 1 {
		repl()
	} else if len(os.Args) == 2 {
		runFile(os.Args[1])
	} else {
		fmt.Fprintf(os.Stderr, "Usage: crex [path]\n")
		os.Exit(1)
	}

	vm.FreeVM()
}

func repl() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			break
		}
		vm.Interpret(line)
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
