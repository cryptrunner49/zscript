package integration

import (
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestFunctionCall(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `func add(a, b):
    return a + b
println(add(5, 3))`
	expectedOutput := "8\n"

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

func TestFunctionWithString(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `func greet(name):
    return "Hello, " + name + "!"
println(greet("World"))`
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
}
