package pattern

type Pattern interface {
	Type()
}

type ConstructorPattern struct {
}

type VariablePattern struct{}

type ConstPattern struct{}
