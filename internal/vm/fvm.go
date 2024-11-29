package vm

import (
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
	"github.com/emrzvv/fl-compiler/internal/types/pattern"
)

const MaxFrames = 1024
const StackSize = 2048

type FVM struct {
	constants []object.Object
	patterns  []pattern.Pattern
	variables []object.Object

	frames      []*Frame
	framesIndex int

	stack []object.Object
	sp    int
}

func (fvm *FVM) currentFrame() *Frame {
	return fvm.frames[fvm.framesIndex-1]
}

func (fvm *FVM) pushFrame(f *Frame) {
	fvm.frames[fvm.framesIndex] = f
	fvm.framesIndex++
}

func (fvm *FVM) popFrame() *Frame {
	fvm.framesIndex--
	return fvm.frames[fvm.framesIndex]
}

func NewFVM(bytecode *compiler.Bytecode) *FVM {
	main := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(main, 0)
	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &FVM{
		constants:   bytecode.Constants,
		patterns:    bytecode.Patterns,
		variables:   make([]object.Object, bytecode.VarAmount+1), // TODO: what len?
		frames:      frames,
		framesIndex: 1,
		stack:       make([]object.Object, StackSize),
		sp:          0,
	}
}

func (fvm *FVM) Run() error {
	// fmt.Println(fvm.currentFrame().Instructions().String())
	// fmt.Println("====================")

	var ip int
	var instructions code.Instructions
	var op code.OpCode

	for fvm.currentFrame().ip < len(fvm.currentFrame().Instructions())-1 {
		fvm.currentFrame().ip++
		ip = fvm.currentFrame().ip
		instructions = fvm.currentFrame().Instructions()
		op = code.OpCode(instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(instructions[ip+1:])
			fvm.currentFrame().ip += 2
			err := fvm.push(fvm.constants[constIndex])
			if err != nil {
				return err
			}
		case code.OpVariable:
			variableIndex := code.ReadUint16(instructions[ip+1:])
			fvm.currentFrame().ip += 2
			err := fvm.push(fvm.variables[variableIndex])
			if err != nil {
				return err
			}
		case code.OpAdd:
			amount := code.ReadUint16(instructions[ip+1:])
			fvm.currentFrame().ip += 2
			if int(amount) > len(fvm.stack) {
				return fmt.Errorf("error when adding stack values")
			}
			var result int64 = 0
			for i := 0; i < int(amount); i++ {
				current := fvm.pop()
				currentValue := current.(*object.Integer).Value
				result += currentValue
			}
			fvm.push(&object.Integer{Value: result})
		case code.OpConstruct:
			index := code.ReadUint16(instructions[ip+1:])
			arity := code.ReadUint16(instructions[ip+3:])
			fvm.currentFrame().ip += 4

			constructor, ok := fvm.constants[index].(*object.Constructor)
			if !ok { // TODO: validation?
				return fmt.Errorf("error when exctracting constructor type from constant pull")
			}
			args := make([]object.Object, arity)
			for i := int(arity) - 1; i >= 0; i-- {
				args[i] = fvm.pop()
			}

			instance := &object.Instance{
				Constructor: constructor,
				Args:        args,
			}

			fvm.push(instance)
		case code.OpCall:
			argsAmount := code.ReadUint16(instructions[ip+1:])
			fvm.currentFrame().ip += 2

			function, ok := fvm.stack[fvm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("error when trying to call function")
			}
			frame := NewFrame(function, int(argsAmount))
			fvm.pop() // pop function object
			fvm.pushFrame(frame)
		case code.OpReturnValue:
			// value := fvm.pop()

			fvm.popFrame()

			// err := fvm.push(value)
			// if err != nil {
			// 	return err
			// }
		case code.OpMatchFailed:
			return fmt.Errorf("error when trying to match")
		case code.OpMatch:
			patternIdx := code.ReadUint16(instructions[ip+1:])
			jumpIfFail := code.ReadUint16(instructions[ip+3:])
			fvm.currentFrame().ip += 4

			arg := fvm.pop()
			err := fvm.currentFrame().push(arg)
			if err != nil {
				return err
			}

			pattern := fvm.patterns[patternIdx]
			if pattern.Matches(arg, fvm.variables) {
				continue
			}
			for fvm.currentFrame().sp > 0 {
				fvm.push(fvm.currentFrame().pop())
			}
			fvm.currentFrame().ip = int(jumpIfFail)
		}
	}
	return nil
}

func (fvm *FVM) StackTop() object.Object {
	if fvm.sp == 0 {
		return nil
	}
	return fvm.stack[fvm.sp-1]
}

func (fvm *FVM) push(o object.Object) error {
	if fvm.sp >= StackSize {
		return fmt.Errorf("stack overflow") // TODO: really overflow?
	}

	fvm.stack[fvm.sp] = o
	fvm.sp++

	return nil
}

func (fvm *FVM) pop() object.Object {
	o := fvm.stack[fvm.sp-1]
	fvm.sp--
	return o
}
