package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
)

func run() error {
	inputFile := flag.String("in", "", "Path to input program file")
	outputFile := flag.String("out", "./out", "Path to output binary file")
	verbose := flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	if *inputFile == "" || *outputFile == "" {
		return fmt.Errorf("input or output file is absent")
	}

	program, err := ast.ParseFromFile(*inputFile)
	if err != nil {
		return err
	}
	compiler := compiler.NewCompiler()
	err = compiler.Compile(program)
	if err != nil {
		return err
	}
	if *verbose {
		fmt.Printf("[COMPILED DATA]\n=========\n%v+\n=========\n", compiler.Bytecode())
	}
	err = compiler.Bytecode().WriteToFile(*outputFile)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}
