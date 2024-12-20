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

func (i *Integer) EqualsTo(other Object) bool {
	otherI, ok := other.(*Integer)
	if !ok {
		return false
	}
	return i.Value == otherI.Value
}

type CompiledFunction struct {
	Instructions code.Instructions
}

func (cf *CompiledFunction) Type() ObjectType {
	return COMPILED_FUNCTION_OBJ
}

func (cf *CompiledFunction) String() string {
	return fmt.Sprintf("CompiledFunction[\n%s\n]", cf.Instructions.String())
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

func (c *Constructor) EqualsTo(other Object) bool {
	otherC, ok := other.(*Constructor)
	if !ok {
		return false
	}
	return c.Name == otherC.Name && c.Arity == otherC.Arity && c.Supertype == otherC.Supertype
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
	return fmt.Sprintf(
		"Instance[\n  Constructor: %s\n  Args: [%s]\n]",
		i.Constructor.Name,
		strings.Join(args, ", "),
	)
}
