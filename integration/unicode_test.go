package integration

import (
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestUnicodeVariables(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var 挨拶 = "こんにちは"
println(挨拶)`
	expectedOutput := "こんにちは\n"

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

func TestEmojiStructs(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `struct 🐱
var kitty = 🐱{}
kitty.name = "Whiskers"
println(kitty.name)`
	expectedOutput := "Whiskers\n"

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
