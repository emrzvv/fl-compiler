package main

import (
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
)

func main() {
	program := ast.GetAST("./samples/input6")
	compiler := compiler.NewCompiler()
	compiler.Compile(program)
	fmt.Printf("%v+", compiler.Bytecode())
	compiler.Bytecode().WriteToFile("./bin/out")
}
