package vm

import (
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type Frame struct {
	fn         *object.CompiledFunction
	ip         int
	ArgsAmount int
	args       []object.Object
	stack      []object.Object
	sp         int
}

func NewFrame(fn *object.CompiledFunction, argsAmount int, args []object.Object) *Frame {
	return &Frame{
		fn:         fn,
		ip:         -1,
		ArgsAmount: argsAmount,
		args:       args,
		stack:      make([]object.Object, StackSize),
		sp:         0,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}

func (f *Frame) push(o object.Object) error {
	if f.sp >= StackSize {
		return fmt.Errorf("stack overflow") // TODO: really overflow?
	}

	f.stack[f.sp] = o
	f.sp++

	return nil
}

func (f *Frame) pop() object.Object {
	o := f.stack[f.sp-1]
	f.sp--
	return o
}

func (f *Frame) top() object.Object {
	return f.stack[f.sp-1]
}

func (f *Frame) clear() {
	f.stack = make([]object.Object, StackSize)
}

func (f *Frame) transferArgs() error {
	for i := f.ArgsAmount - 1; i >= 0; i-- {
		err := f.push(f.args[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) printArgs() {
	for i := 0; i < f.ArgsAmount; i++ {
		fmt.Printf("======\n[%d]: %+v\n======\n", i, f.args[i])
	}
}

func (f *Frame) printStack() {
	fmt.Printf("len: %d, cap: %d, sp: %d\n", len(f.stack), cap(f.stack), f.sp)
	for i := f.sp; i >= 0; i-- {
		fmt.Printf("======\n[%d]: %+v\n======\n", i, f.stack[i])
	}
}
