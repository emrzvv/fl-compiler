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

	stack []object.Object
	sp    int
}

func NewFrame(fn *object.CompiledFunction, argsAmount int) *Frame {
	return &Frame{fn: fn, ip: -1, ArgsAmount: argsAmount}
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
