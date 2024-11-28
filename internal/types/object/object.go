package object

import (
	"fmt"
	"strings"

	"github.com/emrzvv/fl-compiler/internal/compiler/code"
)

type ObjectType string

const (
	INTEGER_OBJ           = "INTEGER"
	COMPILED_FUNCTION_OBJ = "COMPILED_FUNCTION"
	CONSTRUCTOR_OBJ       = "CONSTRUCTOR"
	INSTANCE_OBJ          = "INSTANCE"
)

type Object interface {
	Type() ObjectType
	String() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_OBJ
}

func (i *Integer) String() string {
	return fmt.Sprintf("%d", i.Value)
}

type CompiledFunction struct {
	Instructions code.Instructions
}

func (cf *CompiledFunction) Type() ObjectType {
	return COMPILED_FUNCTION_OBJ
}

func (cf *CompiledFunction) String() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

type Constructor struct {
	Name      string
	Arity     int64
	Supertype string
}

func (c *Constructor) Type() ObjectType {
	return CONSTRUCTOR_OBJ
}

func (c *Constructor) String() string {
	return fmt.Sprintf(`Constructor[
name=%s
arity=%d
supertype=%s
	]`, c.Name, c.Arity, c.Supertype)
}

type Instance struct {
	Constructor *Constructor
	Args        []Object
}

func (i *Instance) Type() ObjectType { return "INSTANCE" }
func (i *Instance) String() string {
	args := []string{}
	for _, arg := range i.Args {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("Constructor Instance %s(%s)", i.Constructor.Name, strings.Join(args, ", "))
}
