package compiler

import (
	"encoding/binary"
	"fmt"

	"github.com/emrzvv/fl-compiler/internal/compiler/ast"
	"github.com/emrzvv/fl-compiler/internal/compiler/code"
	"github.com/emrzvv/fl-compiler/internal/types/object"
	"github.com/emrzvv/fl-compiler/internal/types/pattern"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
	// superToSub          map[string][]string
	// subToSuper          map[string]string
	// constructorArity    map[string]int
	constructorsMapping map[string]int
	patterns            []pattern.Pattern
	patmatJumps         []int
	matches             [][]int
}

func NewCompiler() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
		// superToSub:       make(map[string][]string),
		// subToSuper:       make(map[string]string),
		// constructorArity: make(map[string]int),
		constructorsMapping: make(map[string]int),
		patterns:            []pattern.Pattern{},
		patmatJumps:         []int{},
		matches:             [][]int{},
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
			if d.FunDef != nil {
				err := c.Compile(d.FunDef)
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
	case *ast.FunDef:
		c.patmatJumps = make([]int, 0)
		c.matches = make([][]int, 0)
		// begin := len(c.instructions) - 1
		for _, rule := range node.Rules {
			err := c.Compile(rule)
			if err != nil {
				return err
			}
		}
		end := c.emit(code.OpMatchFailed)
		c.patmatJumps = append(c.patmatJumps, end)
		c.patmatJumps = c.patmatJumps[1:]
		c.setPatmatJumpingPoints()

	case *ast.FunRule:
		patternDef := node.Pattern
		expr := node.Expression
		matchesPositions := []int{}

		for _, patternArg := range patternDef.Arguments {
			pattern, err := c.collectPattern(patternArg)
			if err != nil {
				return err
			}
			patternIndex := c.addPattern(pattern)
			matchPos := c.emit(code.OpMatch, patternIndex, 0)
			matchesPositions = append(matchesPositions, matchPos)
		}
		c.patmatJumps = append(c.patmatJumps, matchesPositions[0])
		c.matches = append(c.matches, matchesPositions)
		c.Compile(expr)
		c.emit(code.OpReturnValue)

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

func (c *Compiler) setPatmatJumpingPoints() error {
	for i, jumpTo := range c.patmatJumps {
		for _, matchIdx := range c.matches[i] {
			offset := 3
			binary.BigEndian.PutUint16(c.instructions[matchIdx+offset:], uint16(jumpTo))
		}
	}
	return nil // TODO: ???
}

func (c *Compiler) collectPattern(p *ast.PatternArgument) (pattern.Pattern, error) {
	if p.Variable != "" {
		return &pattern.VariablePattern{
			Name: p.Variable,
		}, nil
	}
	if p.Const != nil {
		return &pattern.ConstPattern{
			Const: &object.Integer{ // TODO: not only integer possible const
				Value: int64(p.Const.Number),
			},
		}, nil
	}
	if p.Name.Name != "" {
		args := []pattern.Pattern{}
		for _, arg := range p.Arguments {
			argPattern, err := c.collectPattern(arg)
			if err != nil {
				return &pattern.ConstructorPattern{}, err
			}
			args = append(args, argPattern)
		}
		constrIndex, ok := c.constructorsMapping[p.Name.Name]
		if !ok {
			return nil, fmt.Errorf("could not find constructor %s on index %d", p.Name.Name, constrIndex)
		}
		return &pattern.ConstructorPattern{
			Constructor: c.constants[constrIndex].(*object.Constructor),
			Args:        args,
		}, nil
	}
	return nil, fmt.Errorf("could not construct pattern: %+v", p)
}

func (c *Compiler) addPattern(pat pattern.Pattern) int {
	c.patterns = append(c.patterns, pat)
	return len(c.patterns) - 1
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
