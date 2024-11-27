package compiler

import (
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
	// superToSub          map[string][]string
	// subToSuper          map[string]string
	// constructorArity    map[string]int
	constructorsMapping map[string]int
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		// superToSub:       make(map[string][]string),
		// subToSuper:       make(map[string]string),
		// constructorArity: make(map[string]int),
		constructorsMapping: make(map[string]int),
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, d := range node.Definitions {
			if d.FunCall != nil {
				err := c.Compile(d.FunCall)
				if err != nil {
					return err
				}
			}
			if d.TypeDef != nil {
				err := c.Compile(d.TypeDef)
				if err != nil {
					return err
				}
			}
			// TODO: everything below - to remove
			if d.ExprConstructor != nil {
				err := c.Compile(d.ExprConstructor)
				if err != nil {
					return err
				}
			}
		}
	case *ast.TypeDef:
		// subtypes := []string{}
		for _, alt := range node.TypeAlternatives {
			constructorName := alt.Constructor.Name
			constructorArity := len(alt.Constructor.Parameters)
			supertype := node.TypeName.Name

			constructorObj := &object.Constructor{
				Name:      constructorName,
				Arity:     int64(constructorArity),
				Supertype: supertype,
			}

			index := c.addConstant(constructorObj)
			c.constructorsMapping[constructorName] = index
			// subtypes = append(subtypes, constructorName)
			// c.subToSuper[constructorName] = supertype
			// c.constructorArity[constructorName] = constructorArity
		}
		// c.superToSub[node.TypeName.Name] = subtypes
	case *ast.ExprConstructor:
		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}
		name := node.Name.Name
		index := c.constructorsMapping[name]
		c.emit(code.OpConstruct, index, len(node.Arguments))
	case *ast.FunCall:
		switch node.Name {
		case "+":
			for _, expr := range *node.Arguments {
				if expr.Const != nil {
					err := c.Compile(expr.Const)
					if err != nil {
						return err
					}
				}
				if expr.FunCall != nil {
					err := c.Compile(expr.FunCall)
					if err != nil {
						return err
					}
				}
			}
			c.emit(code.OpAdd, len(*node.Arguments))
		default:
			fmt.Println("TODO compile funcall")
		}
	case *ast.Expression:
		if node.Const != nil {
			err := c.Compile(node.Const)
			if err != nil {
				return err
			}
		}
		if node.ExprConstructor != nil {
			err := c.Compile(node.ExprConstructor)
			if err != nil {
				return err
			}
		}
	case *ast.Const:
		number := &object.Integer{Value: int64(node.Number)}
		c.emit(code.OpConstant, c.addConstant(number))
	}
	return nil
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.OpCode, operands ...int) int {
	instr := code.Make(op, operands...)
	pos := c.addInstruction(instr)
	return pos
}

func (c *Compiler) addInstruction(instruction []byte) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, instruction...)
	return pos
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
