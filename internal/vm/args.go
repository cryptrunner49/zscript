package vm

import (
	"github.com/cryptrunner49/zscript/internal/runtime"
)

// defineArgs creates a global variable "args" containing an array of command-line arguments,
// where each argument is converted to an ObjString and stored as a runtime Value.
func defineArgs(args []string) {
	elements := make([]runtime.Value, len(args))
	for i, arg := range args {
		elements[i] = runtime.ObjVal(runtime.NewObjString(arg))
	}

	// Define the "args" global as an array.
	argsName := runtime.NewObjString("args")
	vm.globals[argsName] = runtime.ObjVal(runtime.NewArray(elements))
}
