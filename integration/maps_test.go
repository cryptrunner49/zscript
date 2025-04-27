package integration

import (
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestMapBasicOperations(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var map = { "name": "Alice", "age": 30 }
println(map["name"])
map["age"] = 31
println(map["age"])`
	expectedOutput := "Alice\n31\n"

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

func TestMapFunctions(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var m = {"a": 1, "b": 2}
map_remove(m, "a")
println(map_contains_key(m, "b"))
println(map_size(m))`
	expectedOutput := "true\n1\n"

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

func TestMapAddition(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var a = {"x": 1, "y": 2}
var b = {"y": 3, "z": 4}
var c = a + b
println(c["x"])
println(c["y"])
println(c["z"])`
	expectedOutput := "1\n3\n4\n"

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
