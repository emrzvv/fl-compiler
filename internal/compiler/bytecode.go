package compiler

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"os"

	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	VarAmount    int
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
		VarAmount:    c.varAmount,
	}
}

func (b *Bytecode) WriteToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = binary.Write(file, binary.BigEndian, uint32(len(b.Instructions)))
	if err != nil {
		return err
	}
	_, err = file.Write(b.Instructions)
	if err != nil {
		return err
	}
	gob.Register(&object.Constructor{})
	gob.Register(&object.Instance{})
	gob.Register(&object.Integer{})
	gob.Register(&object.CompiledFunction{})

	constantsData, err := b.serializeConstants()
	if err != nil {
		return err
	}
	err = binary.Write(file, binary.BigEndian, uint32(len(constantsData)))
	if err != nil {
		return err
	}
	_, err = file.Write(constantsData)
	binary.Write(file, binary.BigEndian, uint32(b.VarAmount))
	return err
}

// func (b *Bytecode) serializeConstants() ([]byte, error) {
// 	var buffer bytes.Buffer

// 	encoder := gob.NewEncoder(&buffer)

// 	for _, constant := range b.Constants {
// 		if err := encoder.Encode(constant); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return buffer.Bytes(), nil

// }

func (b *Bytecode) serializeConstants() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)

	for _, constant := range b.Constants {
		switch constant := constant.(type) {
		case *object.Integer:
			if err := encoder.Encode(object.INTEGER_OBJ); err != nil {
				return nil, err
			}
			if err := encoder.Encode(constant); err != nil {
				return nil, err
			}
		case *object.CompiledFunction:
			if err := encoder.Encode(object.COMPILED_FUNCTION_OBJ); err != nil {
				return nil, err
			}
			if err := encoder.Encode(constant); err != nil {
				return nil, err
			}
		case *object.Constructor:
			if err := encoder.Encode(object.CONSTRUCTOR_OBJ); err != nil {
				return nil, err
			}
			if err := encoder.Encode(constant); err != nil {
				return nil, err
			}
		case *object.Instance:
			if err := encoder.Encode(object.INSTANCE_OBJ); err != nil {
				return nil, err
			}
			if err := encoder.Encode(constant); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported constant type: %T", constant)
		}
	}

	return buffer.Bytes(), nil
}

func ReadFromFile(filename string) (*Bytecode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var instructionCount uint32
	err = binary.Read(file, binary.BigEndian, &instructionCount)
	if err != nil {
		return nil, err
	}

	instructions := make([]byte, instructionCount)
	_, err = file.Read(instructions)
	if err != nil {
		return nil, err
	}

	var constantsSize uint32
	err = binary.Read(file, binary.BigEndian, &constantsSize)
	if err != nil {
		return nil, err
	}

	constantsData := make([]byte, constantsSize)
	_, err = file.Read(constantsData)
	if err != nil {
		return nil, err
	}
	gob.Register(&object.Integer{})
	gob.Register(&object.CompiledFunction{})
	gob.Register(&object.Constructor{})
	gob.Register(&object.Instance{})

	constants, err := deserializeConstants(constantsData)
	if err != nil {
		return nil, err
	}
	var varAmount uint32
	err = binary.Read(file, binary.BigEndian, &varAmount)
	if err != nil {
		return nil, err
	}
	return &Bytecode{
		Instructions: code.Instructions(instructions),
		Constants:    constants,
		VarAmount:    int(varAmount),
	}, nil
}

// func deserializeConstants(data []byte) ([]object.Object, error) {
// 	buffer := bytes.NewBuffer(data)

// 	decoder := gob.NewDecoder(buffer)
// 	var constants []object.Object
// 	for buffer.Len() > 0 {
// 		var obj object.Object
// 		if err := decoder.Decode(&obj); err != nil {
// 			return nil, err
// 		}
// 		constants = append(constants, obj)
// 	}
// 	return constants, nil

// }

func deserializeConstants(data []byte) ([]object.Object, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)

	var constants []object.Object
	for buffer.Len() > 0 {
		var objType object.ObjectType
		if err := decoder.Decode(&objType); err != nil {
			return nil, err
		}

		switch objType {
		case object.INTEGER_OBJ:
			var integer object.Integer
			if err := decoder.Decode(&integer); err != nil {
				return nil, err
			}
			constants = append(constants, &integer)
		case object.COMPILED_FUNCTION_OBJ:
			var compiledFunction object.CompiledFunction
			if err := decoder.Decode(&compiledFunction); err != nil {
				return nil, err
			}
			constants = append(constants, &compiledFunction)
		case object.CONSTRUCTOR_OBJ:
			var constructor object.Constructor
			if err := decoder.Decode(&constructor); err != nil {
				return nil, err
			}
			constants = append(constants, &constructor)
		case object.INSTANCE_OBJ:
			var instance object.Instance
			if err := decoder.Decode(&instance); err != nil {
				return nil, err
			}
			constants = append(constants, &instance)
		default:
			return nil, fmt.Errorf("unknown object type %s", objType)
		}
	}
	return constants, nil
}
