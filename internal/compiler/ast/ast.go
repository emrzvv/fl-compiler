package ast

import (
	"fmt"
	"os"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Node interface {
	String() string
}

type Program struct {
	Pos lexer.Position

	Definitions []*Definition `@@*`
}

func (p *Program) String() string {
	return "tmp"
}

type Definition struct {
	Pos lexer.Position

	TypeDef *TypeDef `@@`
	FunDef  *FunDef  `| @@`
	FunCall *FunCall `| @@`
}

func (d *Definition) String() string { return "tmp" }

type TypeDef struct {
	Pos lexer.Position

	TypeName         *TypeName          `"type" "[" @@`
	TypeGeneral      []*TypeGeneral     `@@* "]" ":"`
	TypeAlternatives []*TypeAlterantive `@@ ("|" @@)* "."`
}

func (td *TypeDef) String() string { return "tmp" }

type TypeName struct {
	Pos lexer.Position

	Name string `@Ident`
}

func (tn *TypeName) String() string { return "tmp" }

type TypeGeneral struct {
	Pos lexer.Position

	Name string `@Ident`
}

func (tg *TypeGeneral) String() string { return "tmp" }

type TypeAlterantive struct {
	Pos lexer.Position

	Constructor *Constructor `@@`
}

func (ta *TypeAlterantive) String() string { return "tmp" }

type Constructor struct {
	Pos lexer.Position

	Name       string                  `@Ident`
	Parameters []*ConstructorParameter `@@*`
}

func (c *Constructor) String() string { return "tmp" }

type ConstructorParameter struct {
	Pos lexer.Position

	TypeName    *TypeName        `"[" @@`
	List        []*TypeParameter `@@* "]"`
	TypeGeneral *TypeGeneral     `| @@`
}

func (cp *ConstructorParameter) String() string { return "tmp" }

type TypeParameter struct {
	Pos lexer.Position

	TypeCommon  *TypeCommon  `@@`
	TypeGeneral *TypeGeneral `| @@`
	TypeBuiltin *TypeBuiltin `| @@`
}

func (tp *TypeParameter) String() string { return "tmp" }

type TypeCommon struct {
	Pos lexer.Position

	TypeName       *TypeName        `"[" @@`
	TypeParameters []*TypeParameter `@@* "]"`
	TypeBuiltin    *TypeBuiltin     `| @@`
}

func (tc *TypeCommon) String() string { return "tmp" }

type TypeBuiltin struct {
	Pos lexer.Position

	Type string `@("Int")`
}

func (tb *TypeBuiltin) String() string { return "tmp" }

type FunDef struct {
	Pos lexer.Position

	Signature *FunSignature `@@ ":"`
	Rules     []*FunRule    `@@ ("|" @@)* "."`
}

func (fd *FunDef) String() string { return "tmp" }

type FunSignature struct {
	Pos lexer.Position

	Name       string        `"fun" "(" @Ident`
	Parameters []*TypeCommon `@@* ")" "->"`
	ReturnType *TypeCommon   `@@`
}

func (fs *FunSignature) String() string { return "tmp" }

type FunRule struct {
	Pos lexer.Position

	Pattern    *Pattern    `@@ "->"`
	Expression *Expression `@@`
}

func (fr *FunRule) String() string { return "tmp" }

type Pattern struct {
	Pos lexer.Position

	FunName   string             `"(" @Ident`
	Arguments []*PatternArgument `@@* ")"`
}

func (p *Pattern) String() string { return "tmp" }

type PatternArgument struct {
	Pos lexer.Position

	Name      TypeName           `"[" @@`
	Arguments []*PatternArgument `@@* "]"`
	Variable  string             `| @Ident`
	Const     *Const             `| @@`
}

func (pa *PatternArgument) String() string { return "tmp" }

type Expression struct {
	Pos lexer.Position

	FunCall         *FunCall         `@@`
	ExprConstructor *ExprConstructor `| @@`
	Const           *Const           `| @@`
	Variable        string           `| @Ident`
}

func (e *Expression) String() string { return "tmp" }

type FunCall struct {
	Pos lexer.Position

	Name      string        `"(" @Ident`
	Arguments []*Expression `@@* ")"`
}

func (fc *FunCall) String() string { return "tmp" }

type Const struct {
	Pos lexer.Position

	Number int `@Int`
}

func (c *Const) String() string { return "tmp" }

type ExprConstructor struct {
	Pos lexer.Position

	Name      TypeName      `"[" @@`
	Arguments []*Expression `@@* "]"`
}

func (ec *ExprConstructor) String() string { return "tmp" }

func GetAST(path string) *Program {
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
	})

	parser := participle.MustBuild[Program](
		participle.Lexer(myLexer),
	)

	r, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	program, err := parser.Parse(path, r)

	if err != nil {
		panic(err)
	}

	return program
}

func ParseFromFile(path string) (*Program, error) {
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
	})

	parser := participle.MustBuild[Program](
		participle.Lexer(myLexer),
	)

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	program, err := parser.Parse(path, r)
	if err != nil {
		return nil, err
	}

	return program, nil
}

type TypeDefKey struct {
	Name  string
	Arity int
}

type ConstructorDefKey struct {
	Name      string
	Arity     int
	Supertype *TypeDefKey
}

type FunctionDefKey struct {
	Name string
}

type VariableDefKey struct {
	FunName string
	VarName string
	Branch  int
}

func CheckSemantics(
	node Node,
	types map[TypeDefKey]interface{},
	constructors map[ConstructorDefKey]interface{},
	functions map[FunctionDefKey]interface{},
	variables map[VariableDefKey]interface{},
	// currentFunction *FunSignature,
	// currentBranch int,
) error {
	switch node := node.(type) {
	case *Program:
		for _, d := range node.Definitions {
			if d.TypeDef != nil {
				def := getTypeDefKey(d.TypeDef)
				_, ok := types[*def]
				if ok {
					return fmt.Errorf(
						"pos %v\ntype %v already declared",
						d.TypeDef.Pos,
						d.TypeDef,
					)
				}
				types[*def] = struct{}{}
				for _, ta := range d.TypeDef.TypeAlternatives {
					cdef := getConstructorDefKey(def, ta.Constructor)
					_, ok := constructors[*cdef]
					if ok {
						return fmt.Errorf(
							"pos %v\nconstructor %v already declared",
							ta.Constructor.Pos,
							ta.Constructor,
						)
					}

					constructors[*cdef] = struct{}{}
				}
			}

		}
	}
	return nil
}

func getTypeDefKey(t *TypeDef) *TypeDefKey {
	return &TypeDefKey{
		Name:  t.TypeName.Name,
		Arity: len(t.TypeGeneral),
	}
}

func getConstructorDefKey(supertype *TypeDefKey, c *Constructor) *ConstructorDefKey {
	return &ConstructorDefKey{
		Name:      c.Name,
		Arity:     len(c.Parameters),
		Supertype: supertype,
	}
}
