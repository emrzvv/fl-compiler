package vm

import (
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

const StackSize = 2048 // TODO: to config

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    int
}

func NewVM(bytecode *compiler.Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.OpCode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpAdd:
			amount := code.ReadUint16(vm.instructions[ip+1:])
			ip += 2
			if int(amount) > len(vm.stack) {
				return fmt.Errorf("error when adding stack values")
			}
			var result int64 = 0
			for i := 0; i < int(amount); i++ {
				current := vm.pop()
				currentValue := current.(*object.Integer).Value
				result += currentValue
			}
			vm.push(&object.Integer{Value: result})
		case code.OpConstruct:
			index := code.ReadUint16(vm.instructions[ip+1:])
			arity := code.ReadUint16(vm.instructions[ip+3:])
			ip += 4

			constructor, ok := vm.constants[index].(*object.Constructor)
			if !ok { // TODO: validation?
				return fmt.Errorf("error when exctracting constructor type from constant pull")
			}
			args := make([]object.Object, arity)
			for i := int(arity) - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}

			instance := &object.Instance{
				Constructor: constructor,
				Args:        args,
			}

			vm.push(instance)
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow") // TODO: really overflow?
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}
