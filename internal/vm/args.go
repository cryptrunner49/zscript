package vm

import (
	"fmt"

	"github.com/cryptrunner49/goseedvm/internal/runtime"
)

// Define the "args" global variable as an instance of a struct with fields
func defineArgs(args []string) {
	// Define the Args struct.
	objStruct := runtime.NewStruct(runtime.NewObjString("Args"))

	// Create an instance of the Args struct.
	objInstance := runtime.NewInstance(objStruct)

	// Populate the instance's fields with argument data.
	objInstance.Fields[runtime.NewObjString("length")] = runtime.Value{
		Type:   runtime.VAL_NUMBER,
		Number: float64(len(args)),
	}
	for i, arg := range args {
		// Use valid identifier keys: "_0", "_1", etc.
		key := runtime.NewObjString(fmt.Sprintf("_%d", i))
		objInstance.Fields[key] = runtime.Value{
			Type: runtime.VAL_OBJ,
			Obj:  runtime.NewObjString(arg),
		}
	}

	// Define the "args" global as the instance.
	argsName := runtime.NewObjString("args")
	vm.globals[argsName] = runtime.Value{
		Type: runtime.VAL_OBJ,
		Obj:  objInstance,
	}
}
