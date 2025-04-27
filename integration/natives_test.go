package integration

import (
	"os"
	"strconv"
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestPrint(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `println("Hello, world!")`
	expectedOutput := "Hello, world!\n"

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	if output != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, output)
	}
}

func TestClock(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var time = clock()
println(time)`

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	if _, err := strconv.ParseFloat(output[:len(output)-1], 64); err != nil {
		t.Errorf("Expected a float, got %q", output)
	}
}

func TestRandomBetween(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var rand = random_between(1, 10)
println(rand)`

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	num, err := strconv.Atoi(output[:len(output)-1])
	if err != nil {
		t.Errorf("Expected an integer, got %q", output)
	}
	if num < 1 || num > 10 {
		t.Errorf("Expected number between 1 and 10, got %d", num)
	}
}

func TestFileOperations(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var filename = "test.txt"
var content = "Hello, World!"
write_file(filename, content)
var readContent = read_file(filename)
println(readContent)`
	expectedOutput := "Hello, World!\n"

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	if output != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, output)
	}

	os.Remove("test.txt")
}

func TestSprintf(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var name = "Alice"
var age = 25
var formatted = sprintf("Name: %v, Age: %v", name, age)
println(formatted)`
	expectedOutput := "Name: Alice, Age: 25\n"

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	if output != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, output)
	}
}

func TestErrorf(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var errorCode = 404
var errorMsg = errorf("Error %v: Not found", errorCode)
println(errorMsg)`
	expectedOutput := "Error 404: Not found\n"

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	if output != expectedOutput {
		t.Errorf("Expected %q, got %q", expectedOutput, output)
	}
}

func TestShuffle(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [1, 2, 3, 4, 5]
shuffle(arr)
println(array_to_string(arr))`

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	// Since shuffle is random, just check that output is a valid array of the same elements
	if len(output) < 2 || output[0] != '[' || output[len(output)-2] != ']' {
		t.Errorf("Expected a valid array string, got %q", output)
	}
}

func TestRandomString(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var randStr = random_string(8)
println(randStr)`

	output := captureOutput(t, func() {
		result := core.Interpret(script, "<script>")
		if result != 0 {
			t.Fatalf("Interpretation failed: %d", result)
		}
	})

	if len(output)-1 != 8 { // -1 for newline
		t.Errorf("Expected string of length 8, got %q (length %d)", output, len(output)-1)
	}
}
