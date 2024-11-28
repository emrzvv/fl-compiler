package compiler

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
	"github.com/emrzvv/fl-compiler/internal/types/pattern"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
	expectedPatterns     []interface{}
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
			expectedPatterns: []interface{}{},
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
			expectedPatterns:     []interface{}{},
		},
		{
			input:                "type [List x]: Cons x [List x] | Nil .",
			expectedConstants:    predefinedConstants["list_full"],
			expectedInstructions: []code.Instructions{},
			expectedPatterns:     []interface{}{},
		},
		{
			input: `type [List x]: Cons x [List x] | Nil .
			[Nil]`,
			expectedConstants: predefinedConstants["list_full"],
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstruct, 1, 0),
			},
			expectedPatterns: []interface{}{},
		},
		{
			input: `type [List x]: Cons x [List x] | Nil .
			[Cons 1 [Cons 2 [Cons 3 [Nil]]]]`,
			expectedConstants: append(predefinedConstants["list_full"], []interface{}{1, 2, 3}...),
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstruct, 1, 0),
				code.Make(code.OpConstruct, 0, 2),
				code.Make(code.OpConstruct, 0, 2),
				code.Make(code.OpConstruct, 0, 2),
			},
			expectedPatterns: []interface{}{},
			// constants: [Cons, Nil, 1, 2, 3]
		},
	}
	runCompilerTests(t, tests)
}

func TestPatterns(t *testing.T) {
	predefinedConstants := map[string][]interface{}{
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
	}
	tests := []compilerTestCase{
		{
			input: `fun (test Int Int) -> Int : 
			(test x y) -> 0 .`,
			expectedConstants: []interface{}{0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpMatch, 0, 14),
				code.Make(code.OpMatch, 1, 14),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpReturnValue),
				code.Make(code.OpMatchFailed),
			},
			expectedPatterns: []interface{}{
				&pattern.VariablePattern{
					Name: "x",
				},
				&pattern.VariablePattern{
					Name: "y",
				},
			},
		},
		{
			input: `type [List x]: Cons x [List x] | Nil .
			fun (sum [List Int]) -> Int :
			(sum [Cons x xs]) -> 1 |
			(sum [Nil]) -> 0 .`,
			expectedConstants: append(predefinedConstants["list_full"], []interface{}{1, 0}...), // 0 1 2 3
			expectedInstructions: []code.Instructions{
				code.Make(code.OpMatch, 0, 9),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpReturnValue),
				code.Make(code.OpMatch, 1, 18),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpReturnValue),
				code.Make(code.OpMatchFailed),
			},
			expectedPatterns: []interface{}{
				&pattern.ConstructorPattern{
					Constructor: predefinedConstants["list_full"][0].(*object.Constructor),
					Args: []pattern.Pattern{
						&pattern.VariablePattern{
							Name: "x",
						},
						&pattern.VariablePattern{
							Name: "xs",
						},
					},
				},
				&pattern.ConstructorPattern{
					Constructor: predefinedConstants["list_full"][1].(*object.Constructor),
					Args:        []pattern.Pattern{},
				},
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

func testPatterns(
	t *testing.T,
	expected []interface{},
	actual []pattern.Pattern,
) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("patterns size mismatch. expected %d, got %d", len(expected), len(actual))
	}

	for i, expectedPattern := range expected {
		actualPattern := actual[i]

		switch e := expectedPattern.(type) {
		case *pattern.ConstructorPattern:
			ap, ok := actualPattern.(*pattern.ConstructorPattern)
			if !ok {
				return fmt.Errorf("pattern mismatch at index %d: expected ConstructorPattern, got %T", i, actualPattern)
			}
			if e.Constructor.Name != ap.Constructor.Name {
				return fmt.Errorf("constructor mismatch at index %d: expected %s, got %s", i, e.Constructor.Name, ap.Constructor.Name)
			}
			if len(e.Args) != len(ap.Args) {
				return fmt.Errorf("arguments size mismatch at index %d: expected %d, got %d", i, len(e.Args), len(ap.Args))
			}
			// TODO: check constructor arg values (recursively test)
		case *pattern.VariablePattern:
			ap, ok := actualPattern.(*pattern.VariablePattern)
			if !ok {
				return fmt.Errorf("pattern mismatch at index %d: expected VariablePattern, got %T", i, actualPattern)
			}
			if e.Name != ap.Name {
				return fmt.Errorf("variable name mismatch at index %d: expected %s, got %s", i, e.Name, ap.Name)
			}
		case *pattern.ConstPattern:
			ap, ok := actualPattern.(*pattern.ConstPattern)
			if !ok {
				return fmt.Errorf("pattern mismatch at index %d: expected ConstPattern, got %T", i, actualPattern)
			}
			if e.Const.Value != ap.Const.Value {
				return fmt.Errorf("constant value mismatch at index %d: expected %d, got %d", i, e.Const.Value, ap.Const.Value)
			}
		default:
			return fmt.Errorf("unexpected pattern type at index %d: %T", i, expectedPattern)
		}
	}

	return nil
}
