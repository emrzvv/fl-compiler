package vm

import (
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type Frame struct {
	fn         *object.CompiledFunction
	ip         int
	ArgsAmount int
}

func NewFrame(fn *object.CompiledFunction, argsAmount int) *Frame {
	return &Frame{fn: fn, ip: -1, ArgsAmount: argsAmount}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
