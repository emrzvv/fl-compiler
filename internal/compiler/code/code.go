package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Instructions []byte

type OpCode byte

const (
	OpConstant OpCode = iota
	OpAdd
	OpCall
	OpReturnValue
	OpConstruct
	OpMatchConstructor
	OpBindVariable
	OpExpandArgs
	OpMatchConstant
	OpMatchFailed
	OpVariable
	OpPrint
)

type Definition struct {
	Name          string
	OperandWidths []int
}

var definitions = map[OpCode]*Definition{
	OpConstant:         {"OpConstant", []int{2}}, // {const_index}
	OpAdd:              {"OpAdd", []int{2}},      // {args_amount}
	OpCall:             {"OpCall", []int{2}},     // {args_amount}
	OpReturnValue:      {"OpReturnValue", []int{}},
	OpConstruct:        {"OpConstruct", []int{2, 2}},        // {constructor_index, arity}
	OpMatchConstructor: {"OpMatchConstructor", []int{2, 2}}, // {constructor_index, jmp_to_if_not_matched}
	OpBindVariable:     {"OpBindVariable", []int{2}},        // {variable_index}
	OpExpandArgs:       {"OpExpandArgs", []int{}},
	OpMatchConstant:    {"OpMatchConstant", []int{2, 2}}, // {const_index, jmp_to_if_not_matched}
	OpMatchFailed:      {"OpMatchFailed", []int{}},
	OpVariable:         {"OpVariable", []int{2}}, // {variable_index}
	OpPrint:            {"OpPrint", []int{}},
}

func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[OpCode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

func Make(op OpCode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		offset += width
	}

	return instruction
}

func (ins Instructions) String() string {
	var out bytes.Buffer
	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}
		operands, read := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))
		i += 1 + read
	}
	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)
	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}
	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}
	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0
	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}
		offset += width
	}
	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
