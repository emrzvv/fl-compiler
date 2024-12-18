package vm

import (
	"fmt"
	"strings"
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

		vm := NewFVM(comp.Bytecode())
		err = vm.Run()
		t.Logf("VARIABLES")
		vars := []string{}
		for _, v := range vm.variables {
			if v != nil {
				vars = append(vars, v.String())
			}
		}
		t.Logf("\n%s", strings.Join(vars, "\n"))
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}
		stackElem := vm.StackTop()
		t.Logf("Stack Top Element\n%+v", stackElem)
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
	case *object.Integer:
		err := testIntegerObject(t, expected.Value, actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case int:
		err := testIntegerObject(t, int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case *object.Instance:
		err := testInstanceObject(t, *expected, actual)
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
	t.Logf("TESTING INSTANCE OBJECT %+v \n%+v", expected, actual)
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
	Cons := &object.Constructor{
		Name:      "Cons",
		Arity:     2,
		Supertype: "List",
	}
	Nil := &object.Constructor{
		Name:      "Nil",
		Arity:     0,
		Supertype: "List",
	}
	Pair := &object.Constructor{
		Name:      "Pair",
		Arity:     2,
		Supertype: "Pair",
	}
	B := &object.Constructor{
		Name:      "B",
		Arity:     0,
		Supertype: "Letter",
	}
	C := &object.Constructor{
		Name:      "C",
		Arity:     0,
		Supertype: "Letter",
	}
	tests := []vmTestCase{
		{
			`fun (test) -> Int:
			(test) -> 0 .

			(test)`,
			0,
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			fun (sum [List Int]) -> Int :
			(sum [Cons x xs]) -> 1 |
			(sum [Nil]) -> 0 .

			(sum [Cons 1 [Cons 2 [Nil]]])`,
			1,
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			fun (sum [List Int]) -> Int :
			(sum [Cons x xs]) -> (+ x (sum xs)) |
			(sum [Nil]) -> 0 .

			(sum [Nil])`,
			0,
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			type [Pair x y]: Pair x y .

			fun (zip [List x] [List y]) -> [List [Pair x y]] :
			(zip [Cons x xs] [Cons y ys]) -> [Cons [Pair x y] (zip xs ys)] |
			(zip xs ys) -> [Nil] .

			(zip [Cons 1 [Cons 2 [Nil]]] [Cons 3 [Cons 4 [Nil]]])
			`,
			&object.Instance{
				Constructor: Cons,
				Args: []object.Object{
					&object.Instance{
						Constructor: Pair,
						Args: []object.Object{
							&object.Integer{Value: 1},
							&object.Integer{Value: 3},
						},
					},
					&object.Instance{
						Constructor: Cons,
						Args: []object.Object{
							&object.Instance{
								Constructor: Pair,
								Args: []object.Object{
									&object.Integer{Value: 2},
									&object.Integer{Value: 4},
								},
							},
							&object.Instance{
								Constructor: Nil,
								Args:        []object.Object{},
							},
						},
					},
				},
			},
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			type [Pair x y]: Pair x y .

			fun (append [List x] [List x]) -> [List x] :
				(append [Cons x xs] ys) -> [Cons x (append xs ys)] |
				(append [Nil] ys) -> ys .

			(append [Nil] [Cons 1 [Cons 2 [Nil]]])`,
			&object.Instance{
				Constructor: Cons,
				Args: []object.Object{
					&object.Integer{
						Value: 1,
					},
					&object.Instance{
						Constructor: Cons,
						Args: []object.Object{
							&object.Integer{
								Value: 2,
							},
							&object.Instance{
								Constructor: Nil,
								Args:        []object.Object{},
							},
						},
					},
				},
			},
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			type [Pair x y]: Pair x y .

			fun (flatten [List [List x]]) -> [List x]:
				(flatten [Cons [Cons x xs] xss]) -> [Cons x (flatten [Cons xs xss])] |
				(flatten [Cons [Nil] xss]) -> (flatten xss) |
				(flatten [Nil]) -> [Nil] .

			(flatten [Cons [Cons 1 [Cons 2 [Nil]]] [Cons [Cons 3 [Cons 4 [Nil]]] [Nil]]])
			`, // [[1, 2], [3, 4]]
			&object.Instance{
				Constructor: Cons,
				Args: []object.Object{
					&object.Integer{Value: 1},
					&object.Instance{
						Constructor: Cons,
						Args: []object.Object{
							&object.Integer{Value: 2},
							&object.Instance{
								Constructor: Cons,
								Args: []object.Object{
									&object.Integer{Value: 3},
									&object.Instance{
										Constructor: Cons,
										Args: []object.Object{
											&object.Integer{Value: 4},
											&object.Instance{
												Constructor: Nil,
												Args:        []object.Object{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			type [Letter]: A | B | C | D .

			fun (fab [List Letter]) -> [List Letter] :
				(fab [Cons [A] xs]) -> [Cons [B] (fab xs)] |
				(fab [Cons x xs]) -> [Cons x (fab xs)] |
				(fab [Nil]) -> [Nil] .

			(fab [Cons [A] [Cons [B] [Cons [A] [Cons [A] [Nil]]]]])
			`,
			&object.Instance{
				Constructor: Cons,
				Args: []object.Object{
					&object.Instance{Constructor: B, Args: []object.Object{}},
					&object.Instance{
						Constructor: Cons,
						Args: []object.Object{
							&object.Instance{Constructor: B, Args: []object.Object{}},
							&object.Instance{
								Constructor: Cons,
								Args: []object.Object{
									&object.Instance{Constructor: B, Args: []object.Object{}},
									&object.Instance{
										Constructor: Cons,
										Args: []object.Object{
											&object.Instance{Constructor: B, Args: []object.Object{}},
											&object.Instance{Constructor: Nil, Args: []object.Object{}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			`type [List x]: Cons x [List x] | Nil .
			type [Letter]: A | B | C | D .

			fun (fab [List Letter]) -> [List Letter] :
				(fab [Cons [A] xs]) -> [Cons [B] (fab xs)] |
				(fab [Cons x xs]) -> [Cons x (fab xs)] |
				(fab [Nil]) -> [Nil] .

			fun (fbc [List Letter]) -> [List Letter] :
				(fbc [Cons [B] xs]) -> [Cons [C] (fbc xs)] |
				(fbc [Cons x xs]) -> [Cons x (fbc xs)] |
				(fbc [Nil]) -> [Nil] .

			fun (fabc [List Letter]) -> [List Letter] :
				(fabc xs) -> (fbc (fab xs)) .

			(fabc [Cons [A] [Cons [B] [Cons [A] [Cons [C] [Nil]]]]])
			`,
			&object.Instance{
				Constructor: Cons,
				Args: []object.Object{
					&object.Instance{Constructor: C, Args: []object.Object{}},
					&object.Instance{
						Constructor: Cons,
						Args: []object.Object{
							&object.Instance{Constructor: C, Args: []object.Object{}},
							&object.Instance{
								Constructor: Cons,
								Args: []object.Object{
									&object.Instance{Constructor: C, Args: []object.Object{}},
									&object.Instance{
										Constructor: Cons,
										Args: []object.Object{
											&object.Instance{Constructor: C, Args: []object.Object{}},
											&object.Instance{Constructor: Nil, Args: []object.Object{}},
										},
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
