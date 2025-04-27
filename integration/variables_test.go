package integration

import (
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestVariableTypes(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var num = 42
println(num)
var message = "Hello"
println(message)
var isTrue = true
println(isTrue)
var nothing = null
println(nothing)
var num2 = 42
println(num2)
var message2 = "Hello"
println(message2)
var isTrue2 = true
println(isTrue2)
var nothing2 = null
println(nothing2)`
	expectedOutput := "42\nHello\ntrue\nnull\n42\nHello\ntrue\nnull\n"

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

func TestNegativeNumbers(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var negative = -10
println(negative)`
	expectedOutput := "-10\n"

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
