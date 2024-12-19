package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/vm"
)

func run() error {
	inputFile := flag.String("in", "", "Path to input binary file")
	verbose := flag.Bool("v", false, "Verbose mode")
	flag.Parse()

	if *inputFile == "" {
		return fmt.Errorf("input file is absent")
	}
	bytecode, err := compiler.ReadFromFile(*inputFile)
	if err != nil {
		return err
	}
	if *verbose {
		fmt.Printf("[COMPILED DATA]\n=========\n%v+\n=========\n", bytecode)
	}
	fvm := vm.NewFVM(bytecode)
	err = fvm.Run()
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
