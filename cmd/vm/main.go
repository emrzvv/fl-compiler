package main

import (
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/vm"
)

func main() {
	bytecode, err := compiler.ReadFromFile("./bin/out")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%v+", bytecode)
	fvm := vm.NewFVM(bytecode)
	fvm.Run()
}
