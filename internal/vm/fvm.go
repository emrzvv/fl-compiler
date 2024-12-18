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
	mainFrame := NewFrame(main, 0, []object.Object{})
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
	fmt.Println(fvm.currentFrame().Instructions().String())
	fmt.Println("====================")

	var ip int
	var instructions code.Instructions
	var op code.OpCode

	for fvm.currentFrame().ip < len(fvm.currentFrame().Instructions())-1 {
		fvm.currentFrame().ip++
		ip = fvm.currentFrame().ip
		instructions = fvm.currentFrame().Instructions()
		op = code.OpCode(instructions[ip])
		// fmt.Printf("%d %d\n", ip, op)
		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(instructions[ip+1:])
			fvm.currentFrame().ip += 2
			fmt.Printf("constIndex: %d %d", constIndex, ip)
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
			// fmt.Println("============\nSTACK BEFORE OPCONSTRUCT")
			for i := 0; i < fvm.sp; i++ {
				fmt.Printf("%+v\n", fvm.stack[i])
			}
			// fmt.Printf("VARIABLES\n")
			vars := []string{}
			for _, v := range fvm.variables {
				if v != nil {
					vars = append(vars, v.String())
				}
			}
			// fmt.Printf("\n%s\n=============\n", strings.Join(vars, "\n"))
			constructor, ok := fvm.constants[index].(*object.Constructor)
			if !ok { // TODO: validation?
				return fmt.Errorf("error when exctracting constructor type from constant pull")
			}
			args := make([]object.Object, arity)

			for i := int(arity) - 1; i >= 0; i-- {
				args[i] = fvm.pop()
			}
			aa := []string{}
			for _, a := range args {
				aa = append(aa, a.String())
			}
			// fmt.Printf("ARGS FOR CONSTRUCTOR %d\n", index)
			// fmt.Printf("\n%s", strings.Join(aa, "\n"))
			// fmt.Println("============")
			instance := &object.Instance{
				Constructor: constructor,
				Args:        args,
			}

			fvm.push(instance)
		case code.OpCall:
			argsAmount := int(code.ReadUint16(instructions[ip+1:]))
			fvm.currentFrame().ip += 2
			// fmt.Println("STACK BEFORE OPCALL")
			// for i := 0; i < fvm.sp; i++ {
			// 	fmt.Printf("%+v\n", fvm.stack[i])
			// }
			function, ok := fvm.stack[fvm.sp-1].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("error when trying to call function")
			}
			fvm.pop() // pop function object
			args := make([]object.Object, argsAmount)
			fmt.Println("FVM STACK BEGIN")
			fvm.printStack()
			fmt.Println("FVM STACK END")
			for i := 0; i < argsAmount; i++ {
				args[i] = fvm.pop() // transfer args to frame
			}
			frame := NewFrame(function, argsAmount, args)
			err := frame.transferArgs()
			frame.printArgs()
			fmt.Println("-----------")
			frame.printStack()
			if err != nil {
				return err
			}
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
		case code.OpExpandArgs:
			instance, ok := fvm.currentFrame().pop().(*object.Instance)
			if !ok {
				return fmt.Errorf("error when trying to expand constructor")
			}
			for i := len(instance.Args) - 1; i >= 0; i-- {
				err := fvm.currentFrame().push(instance.Args[i])
				if err != nil {
					return err
				}
			}
		case code.OpMatchConstructor:
			instance, ok := fvm.currentFrame().top().(*object.Instance)
			if !ok {
				return fmt.Errorf("error when trying to match constructor, got %+v", fvm.currentFrame().top())
			}
			patternIdx := code.ReadUint16(instructions[ip+1:])
			constructorPattern := fvm.constants[patternIdx]
			jumpIfFail := code.ReadUint16(instructions[ip+3:])
			fvm.currentFrame().ip += 4
			if instance.Constructor.EqualsTo(constructorPattern) {
				if instance.Constructor.Arity == 0 {
					fvm.currentFrame().pop()
				}
				continue
			}
			fvm.currentFrame().clear()
			fvm.currentFrame().transferArgs()
			fvm.currentFrame().ip = int(jumpIfFail) - 1
		case code.OpMatchConstant:
			constant, ok := fvm.currentFrame().pop().(*object.Integer)
			if !ok {
				return fmt.Errorf("error when trying to match constant")
			}
			constantIdx := code.ReadUint16(instructions[ip+1:])
			constantPattern := fvm.constants[constantIdx]
			jumpIfFail := code.ReadUint16(instructions[ip+3:])
			fvm.currentFrame().ip += 4
			if constant.EqualsTo(constantPattern) {
				continue
			}
			fvm.currentFrame().clear()
			fvm.currentFrame().transferArgs()
			fvm.currentFrame().ip = int(jumpIfFail) - 1
		case code.OpBindVariable:
			idx := code.ReadUint16(instructions[ip+1:])
			fvm.variables[idx] = fvm.currentFrame().pop()
			fvm.currentFrame().ip += 2
			// case code.OpMatch:
			// 	patternIdx := code.ReadUint16(instructions[ip+1:])
			// 	jumpIfFail := code.ReadUint16(instructions[ip+3:])
			// 	fvm.currentFrame().ip += 4

			// 	arg := fvm.pop()
			// 	err := fvm.currentFrame().push(arg)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	// fmt.Println("CURRENT STACK")
			// 	// for i := 0; i < fvm.sp; i++ {
			// 	// 	fmt.Printf("%s\n", fvm.stack[i].String())
			// 	// }
			// 	// fmt.Println("--------------")
			// 	pattern := fvm.patterns[patternIdx]
			// 	// fmt.Printf("\nARG %s\n", arg.String())
			// 	if pattern.Matches(arg, fvm.variables) {
			// 		continue
			// 	}
			// 	for fvm.currentFrame().sp > 0 {
			// 		fvm.push(fvm.currentFrame().pop())
			// 	}
			// 	fvm.currentFrame().ip = int(jumpIfFail) - 1
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

func (fvm *FVM) printStack() {
	fmt.Printf("len: %d, cap: %d, sp: %d\n", len(fvm.stack), cap(fvm.stack), fvm.sp)
	for i := fvm.sp; i >= 0; i-- {
		fmt.Printf("======\n[%d]: %+v\n======\n", i, fvm.stack[i])
	}
}
