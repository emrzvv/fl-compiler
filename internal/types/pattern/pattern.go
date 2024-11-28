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
}

type ConstructorPattern struct {
	Constructor *object.Constructor
	Args        []Pattern
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

type VariablePattern struct {
	Name string
}

func (vp *VariablePattern) Type() PatternType {
	return VARIABLE_PAT
}

func (vp *VariablePattern) String() string {
	return vp.Name
}

type ConstPattern struct {
	Const *object.Integer
}

func (cp *ConstPattern) Type() PatternType {
	return CONST_PAT
}

func (cp *ConstPattern) String() string {
	return (*cp.Const).String()
}
