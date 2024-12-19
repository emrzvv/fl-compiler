package compiler

import (
	"encoding/gob"
	"fmt"
	"testing"

	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

func TestSerialization(t *testing.T) {
	constants := []object.Object{
		&object.Integer{Value: 42},
		&object.Constructor{Name: "Test", Arity: 2, Supertype: "Base"},
		&object.Instance{
			Constructor: &object.Constructor{Name: "Example", Arity: 1, Supertype: "Base"},
			Args: []object.Object{
				&object.Integer{Value: 1},
				&object.Integer{Value: 2},
			},
		},
	}
	gob.Register(&object.Constructor{})
	gob.Register(&object.Instance{})
	gob.Register(&object.Integer{})
	gob.Register(&object.CompiledFunction{})
	bytecode := &Bytecode{
		Instructions: code.Instructions{},
		Constants:    constants,
		VarAmount:    0,
	}
	data, err := bytecode.serializeConstants()
	if err != nil {
		fmt.Printf("Serialization failed: %s\n", err)
		return
	}

	decoded, err := deserializeConstants(data)
	if err != nil {
		fmt.Printf("Deserialization failed: %s\n", err)
		return
	}

	fmt.Println("Decoded objects:")
	for _, obj := range decoded {
		fmt.Printf("Type: %T, Value: %+v\n", obj, obj)
	}
}
