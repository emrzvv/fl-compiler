package vm

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/emrzvv/fl-compiler/internal/compiler"
	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.NewCompiler()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		consts := []string{}
		for i, c := range comp.Bytecode().Constants {
			consts = append(consts, fmt.Sprintf("%d-%s", i, c.String()))
		}
		t.Logf("\n%+q", consts)
		t.Logf("\n%s", comp.Bytecode().Instructions.String())
		// vm := NewVM(comp.Bytecode())
		vm := NewFVM(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}
		stackElem := vm.StackTop()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(
	t *testing.T,
	expected interface{},
	actual object.Object,
) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(t, int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case object.Instance:
		err := testInstanceObject(t, expected, actual)
		if err != nil {
			t.Errorf("testInstanceObject failed: %s", err)
		}
	}
}

func parse(input string) *ast.Program {

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

	parser := participle.MustBuild[ast.Program](
		participle.Lexer(myLexer),
	)
	program, _ := parser.ParseString("tests", input)
	return program
}

func testIntegerObject(t *testing.T, expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not integer. got %T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. expected %d, got %d", expected, result.Value)
	}

	return nil
}

func testInstanceObject(
	t *testing.T,
	expected object.Instance,
	actual object.Object) error {
	result, ok := actual.(*object.Instance)
	if !ok {
		return fmt.Errorf("object is not instance. got %T (%+v)", actual, actual)
	}

	if len(result.Args) != len(expected.Args) {
		return fmt.Errorf("instance has wrong args length. expected %d, got %d", len(expected.Args), len(result.Args))
	}

	err := testInstanceConstructor(expected.Constructor, result.Constructor)
	if err != nil {
		return fmt.Errorf("testInstanceConstructor failed: %s", err)
	}

	for i, _ := range result.Args {
		testExpectedObject(t, expected.Args[i], result.Args[i])
	}

	return nil
}

func testInstanceConstructor(
	expected *object.Constructor,
	actual *object.Constructor) error {

	if actual.Name != expected.Name ||
		actual.Arity != expected.Arity ||
		actual.Supertype != expected.Supertype {
		return fmt.Errorf("constructor has wrong fields. expected %+v\ngot %+v", expected, actual)
	}

	return nil
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"(+ 1 2)", 3},
		{"(+ (+ 1 2) 3)", 6},
		{"(+ (+ 1 (+ 2 (+ 3 4) 5) 6) 7)", 28},
	}

	runVmTests(t, tests)
}

func TestExprConstructor(t *testing.T) {
	tests := []vmTestCase{
		{
			`type [List x]: Cons x [List x] | Nil .
			[Nil]`,
			object.Instance{
				Constructor: &object.Constructor{
					Name:      "Nil",
					Arity:     0,
					Supertype: "List",
				},
				Args: []object.Object{},
			},
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			[Cons 1 [Cons 2 [Cons 3 [Nil]]]]`,
			object.Instance{
				Constructor: &object.Constructor{
					Name:      "Cons",
					Arity:     2,
					Supertype: "List",
				},
				Args: []object.Object{
					&object.Integer{
						Value: 1,
					},
					&object.Instance{
						Constructor: &object.Constructor{
							Name:      "Cons",
							Arity:     2,
							Supertype: "List",
						},
						Args: []object.Object{
							&object.Integer{
								Value: 2,
							},
							&object.Instance{
								Constructor: &object.Constructor{
									Name:      "Cons",
									Arity:     2,
									Supertype: "List",
								},
								Args: []object.Object{
									&object.Integer{
										Value: 3,
									},
									&object.Instance{
										Constructor: &object.Constructor{
											Name:      "Nil",
											Arity:     0,
											Supertype: "List",
										},
										Args: []object.Object{},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runVmTests(t, tests)
}

func TestFunCalls(t *testing.T) {
	tests := []vmTestCase{
		{
			`fun (test) -> Int : 
			(test) -> 0 .
			
			(test)`,
			0,
		},
	}
	runVmTests(t, tests)
}
