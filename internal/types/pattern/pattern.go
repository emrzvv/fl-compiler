package pattern

import (
	"fmt"
	"strings"

	"github.com/emrzvv/fl-compiler/internal/types/object"
)

type PatternType string

const (
	CONSTRUCTOR_PAT = "CONSTRUCTOR"
	VARIABLE_PAT    = "VARIABLE"
	CONST_PAT       = "CONST"
)

type Pattern interface {
	Type() PatternType
	String() string
	Matches(obj object.Object, variables []object.Object) bool
}

type ConstructorPattern struct {
	Constructor *object.Constructor
	Args        []Pattern
	Index       int
}

func (cp *ConstructorPattern) Type() PatternType {
	return CONSTRUCTOR_PAT
}

func (cp *ConstructorPattern) String() string {
	args := []string{}
	for _, arg := range cp.Args {
		args = append(args, arg.String())
	}
	return fmt.Sprintf("Constructor Pattern %s(%s)", cp.Constructor.Name, strings.Join(args, ", "))
}

func (cp *ConstructorPattern) Matches(obj object.Object, variables []object.Object) bool {
	switch obj := obj.(type) {
	case *object.Instance:
		if cp.Constructor.Name == obj.Constructor.Name &&
			cp.Constructor.Arity == obj.Constructor.Arity &&
			cp.Constructor.Supertype == obj.Constructor.Supertype {

			for i, _ := range cp.Args {
				if !cp.Args[i].Matches(obj.Args[i], variables) {
					return false
				}
			}
			return true
		}
	default:
		return false
	}

	return false
}

type VariablePattern struct {
	Name      string
	FunName   string
	RuleIndex int
	Index     int
}

func (vp *VariablePattern) Type() PatternType {
	return VARIABLE_PAT
}

func (vp *VariablePattern) String() string {
	return fmt.Sprintf("%s %s %d", vp.Name, vp.FunName, vp.RuleIndex)
}

func (vp *VariablePattern) Matches(obj object.Object, variables []object.Object) bool {
	// fmt.Printf("\n%s MATCHES? %s\n", vp.Name, obj.String())
	variables[vp.Index] = obj
	return true
}

type ConstPattern struct {
	Const *object.Integer
	Index int
}

func (cp *ConstPattern) Type() PatternType {
	return CONST_PAT
}

func (cp *ConstPattern) String() string {
	return (*cp.Const).String()
}

func (cp *ConstPattern) Matches(obj object.Object, variables []object.Object) bool {
	result, ok := obj.(*object.Integer)
	return ok && result.Value == cp.Const.Value
}
