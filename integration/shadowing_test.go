package integration

import (
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestShadowingSameScope(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var hue = "Red"
var hue = "Blue"
println(hue)`
	expectedOutput := "Blue\n"

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

func TestShadowingDifferentScopes(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var color = "Yellow"
var outer = "Outer"
func shadow():
    var color = "Green"
    println(color)

shadow()
println(color)
println(outer)`
	expectedOutput := "Green\nYellow\nOuter\n"

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

func TestShadowingWithDifferentTypes(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var value = 100
var value = "Text"
println(value)`
	expectedOutput := "Text\n"

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
