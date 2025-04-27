package integration

import (
	"testing"

	"github.com/cryptrunner49/zscript/internal/core"
	"github.com/cryptrunner49/zscript/internal/vm"
)

func TestArrayCreationAndAccess(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [1, 2, 3]
println(arr[0])
println(arr[1])
println(arr[2])`
	expectedOutput := "1\n2\n3\n"

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

func TestArrayPushAndPop(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [1, 2, 3]
push(arr, 4)
println(arr[3])
println(pop(arr))
println(len(arr))`
	expectedOutput := "4\n4\n3\n"

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

func TestArrayLength(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [1, 2, 3, 4, 5]
println(len(arr))`
	expectedOutput := "5\n"

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

func TestArraySorting(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [5, 3, 8, 1, 42, 10]
array_sort(arr)
println(arr)`
	expectedOutput := "[1, 3, 5, 8, 10, 42]\n"

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

func TestArraySplit(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [1, 2, "sep", 3, 4, "sep", 5, 6]
var split = array_split(arr, "sep")
println(array_to_string(split[0]))
println(array_to_string(split[1]))
println(array_to_string(split[2]))`
	expectedOutput := "[1, 2]\n[3, 4]\n[5, 6]\n"

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

func TestArrayJoin(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var a1 = [1, 2]
var a2 = [3, 4]
var a3 = [5, 6]
var joined = array_join(a1, a2, a3)
println(array_to_string(joined))`
	expectedOutput := "[1, 2, 3, 4, 5, 6]\n"

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

func TestArraySearch(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = ["cat", "dog", "bird", "dog"]
var linear = array_linear_search(arr, "dog")
var sorted = ["apple", "banana", "cherry", "date"]
var binary = array_binary_search(sorted, "cherry")
println(linear)
println(binary)`
	expectedOutput := "1\n2\n"

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

func TestArraySlices(t *testing.T) {
	vm.InitVM([]string{"zscript"})
	t.Cleanup(vm.FreeVM)

	script := `var arr = [1, 2, 3, 4, 5]
var slice1 = arr[1:3]
var slice2 = arr[2:]
println(array_to_string(slice1))
println(array_to_string(slice2))`
	expectedOutput := "[2, 3]\n[3, 4, 5]\n"

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
