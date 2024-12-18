package compiler

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {

	tests := []compilerTestCase{
		{
			input:             "(+ 1 2)",
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd, 2),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestTypeDefinitions(t *testing.T) {
	predefinedConstants := map[string][]interface{}{
		"list_nil": {
			object.Constructor{
				Name:      "Nil",
				Arity:     0,
				Supertype: "List",
			},
		},
		"list_full": {
			object.Constructor{
				Name:      "Cons",
				Arity:     2,
				Supertype: "List",
			},
			object.Constructor{
				Name:      "Nil",
				Arity:     0,
				Supertype: "List",
			},
		},
	}
	tests := []compilerTestCase{
		{
			input:                "type [List x]: Nil .",
			expectedConstants:    predefinedConstants["list_nil"],
			expectedInstructions: []code.Instructions{},
		},
		{
			input:                "type [List x]: Cons x [List x] | Nil .",
			expectedConstants:    predefinedConstants["list_full"],
			expectedInstructions: []code.Instructions{},
		},
	}
	runCompilerTests(t, tests)
}

func TestPatterns(t *testing.T) {
	predef := map[string][]interface{}{
		"list_nil": {
			&object.Constructor{
				Name:      "Nil",
				Arity:     0,
				Supertype: "List",
			},
		},
		"list_full": {
			&object.Constructor{
				Name:      "Cons",
				Arity:     2,
				Supertype: "List",
			},
			&object.Constructor{
				Name:      "Nil",
				Arity:     0,
				Supertype: "List",
			},
		},
		"int_int_simple": {
			&object.CompiledFunction{
				Instructions: concatInstructions([]code.Instructions{
					code.Make(code.OpBindVariable, 0),
					code.Make(code.OpBindVariable, 1),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpMatchFailed),
				}),
			},
		},
		"list_1_0": {
			&object.CompiledFunction{
				Instructions: concatInstructions([]code.Instructions{
					code.Make(code.OpMatchConstructor, 0, 16),
					code.Make(code.OpExpandArgs),
					code.Make(code.OpBindVariable, 0),
					code.Make(code.OpBindVariable, 1),
					code.Make(code.OpConstant, 3),
					code.Make(code.OpReturnValue),
					code.Make(code.OpMatchConstructor, 1, 25),
					code.Make(code.OpConstant, 4),
					code.Make(code.OpReturnValue),
					code.Make(code.OpMatchFailed),
				}),
			},
		},
	}
	tests := []compilerTestCase{
		{
			input: `fun (test Int Int) -> Int : 
			(test x y) -> 0 .`,
			expectedConstants:    []interface{}{predef["int_int_simple"], 0},
			expectedInstructions: []code.Instructions{},
		},
		{
			input: `type [List x]: Cons x [List x] | Nil .
			fun (sum [List Int]) -> Int :
			(sum [Cons x xs]) -> 1 |
			(sum [Nil]) -> 0 .`,
			expectedConstants:    append(predef["list_full"], []interface{}{predef["list_1_0"], 1, 0}...), // 0 1 2 3
			expectedInstructions: []code.Instructions{},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunCall(t *testing.T) {
	predef := map[string][]interface{}{
		"list_nil": {
			&object.Constructor{
				Name:      "Nil",
				Arity:     0,
				Supertype: "List",
			},
		},
		"list_full": {
			&object.Constructor{
				Name:      "Cons",
				Arity:     2,
				Supertype: "List",
			},
			&object.Constructor{
				Name:      "Nil",
				Arity:     0,
				Supertype: "List",
			},
		},
		"int_int_simple": {
			&object.CompiledFunction{
				Instructions: concatInstructions([]code.Instructions{
					code.Make(code.OpBindVariable, 0),
					code.Make(code.OpBindVariable, 1),
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpMatchFailed),
				}),
			},
		},
		"list_1_0": {
			&object.CompiledFunction{
				Instructions: concatInstructions([]code.Instructions{
					code.Make(code.OpMatchConstructor, 0, 16),
					code.Make(code.OpExpandArgs),
					code.Make(code.OpBindVariable, 0),
					code.Make(code.OpBindVariable, 1),
					code.Make(code.OpConstant, 3),
					code.Make(code.OpReturnValue),
					code.Make(code.OpMatchConstructor, 1, 25),
					code.Make(code.OpConstant, 4),
					code.Make(code.OpReturnValue),
					code.Make(code.OpMatchFailed),
				}),
			},
		},
	}

	tests := []compilerTestCase{
		{
			input: `fun (test Int Int) -> Int :
			(test x y) -> 0 .

			(test 2 3)`,
			expectedConstants: []interface{}{predef["int_int_simple"], 0, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpCall, 2),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()
	var myLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Keyword", Pattern: `\b(type|fun)\b`},
		{Name: "Operator", Pattern: `->|\||:`},
		{Name: "Ident", Pattern: `[a-zA-Z\+][a-zA-Z0-9_]*`},
		// {Name: "TypeName", Pattern: `[a-zA-Z][a-zA-Z0-9_]*`},
		// {Name: "TypeGeneral", Pattern: `[a-zA-Z][a-zA-Z0-9_]*`},
		// {Name: "FunName", Pattern: `[a-zA-Z\+\-\*\/][a-zA-Z0-9_]*`},
		// {Name: "VarName", Pattern: `[a-zA-Z][a-zA-Z0-9_]`},
		{Name: "Int", Pattern: `[0-9]+`}, // TODO: remove leading zeroes
		{Name: "Punct", Pattern: `[\[\]\(\)\.]`},
		{Name: "whitespace", Pattern: `[ \t\n\r]+`},
	}) // TODO: to config

	parser := participle.MustBuild[ast.Program]( // TODO: custom ast for tests
		participle.Lexer(myLexer),
	)

	for _, tt := range tests {
		program, err := parser.ParseString("tests", tt.input)
		if err != nil {
			t.Fatalf("parse error: %s", err)
		}
		compiler := NewCompiler()
		err = compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}

	}
}

func concatInstructions(instructions []code.Instructions) code.Instructions {
	result := code.Instructions{}
	for _, instr := range instructions {
		result = append(result, instr...)
	}

	return result
}

func testInstructions(
	expectedInstructions []code.Instructions,
	actualInstructions code.Instructions) error {

	expectedConcated := concatInstructions(expectedInstructions)
	if len(actualInstructions) != len(expectedConcated) {
		return fmt.Errorf("instructions length mismatch\nexpected %q\ngot %q",
			expectedConcated, actualInstructions)
	}

	for i, instr := range expectedConcated {
		if instr != actualInstructions[i] {
			return fmt.Errorf("instruction mismatch at %d.\nexpected %q\ngot %q",
				i, expectedConcated, actualInstructions)
		}
	}

	return nil
}

func testConstants(
	t *testing.T,
	expected []interface{},
	actual []object.Object) error {

	if len(expected) != len(actual) {
		return fmt.Errorf("constants size mismatch. expected %d, got %d", len(expected), len(actual))
	}
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case object.Constructor:
			err := testConstructorObject(object.Constructor(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testConstructorObject failed: %s", i, err)
			}
		case object.CompiledFunction:
			err := testCompiledFunctionObject(object.CompiledFunction(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testCompiledFunctionObject failed: %s", i, err)
			}
		}
	}

	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not integer. got %T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. expected %d, got %d", expected, result.Value)
	}

	return nil
}

func testConstructorObject(expected object.Constructor, actual object.Object) error {
	result, ok := actual.(*object.Constructor)
	if !ok {
		return fmt.Errorf("object is not constructor. got %T\n(%+v)", actual, actual)
	}

	if result.Name != expected.Name ||
		result.Arity != expected.Arity ||
		result.Supertype != expected.Supertype {
		return fmt.Errorf("constructor has wrong fields. expected %+v\ngot %+v", expected, result)
	}

	return nil
}

func testCompiledFunctionObject(expected object.CompiledFunction, actual object.Object) error {
	result, ok := actual.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("object is not a c-function. got %T\n(%+v)", actual, actual)
	}

	if len(expected.Instructions) != len(result.Instructions) {
		return fmt.Errorf("functions have wrong instructions length. expected %d, got %d", len(expected.Instructions), len(result.Instructions))
	}

	for i, _ := range expected.Instructions {
		if expected.Instructions[i] != result.Instructions[i] {
			return fmt.Errorf("functions have different instructions. \nexpected %+v\ngot %+v", expected.Instructions, result.Instructions)
		}
	}

	return nil
}
