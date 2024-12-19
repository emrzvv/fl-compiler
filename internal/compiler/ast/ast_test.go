package ast

import (
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func parse(input string) *Program {

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

	parser := participle.MustBuild[Program](
		participle.Lexer(myLexer),
	)
	program, _ := parser.ParseString("tests", input)
	return program
}

func TestCheckSemantics(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple type & constr def",
			input:       `type [List x]: Cons x [List x] | Nil .`,
			expectError: false,
		},
		{
			name: "simple type & constr def",
			input: `
			type [List x]: Cons x [List x] | Nil .
			type [List x]: Abacaba .
			`,
			expectError: true,
		},
		{
			name: "simple type & constr def",
			input: `
			type [List x]: Cons x [List x] | Nil | Cons x [List x] .
			`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(tt.input)
			types := make(map[TypeDefKey]interface{})
			constructors := make(map[ConstructorDefKey]interface{})
			functions := make(map[FunctionDefKey]interface{})
			variables := make(map[VariableDefKey]interface{})
			err := CheckSemantics(
				program,
				types,
				constructors,
				functions,
				variables,
			)
			if (err != nil) != tt.expectError {
				t.Errorf("CheckSemantics() error = %v, expectErr %v", err, tt.expectError)
			}
		})
	}

}
