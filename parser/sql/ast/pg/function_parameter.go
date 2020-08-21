package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type FunctionParameter struct {
	Name    *string
	ArgType *TypeName
	Mode    FunctionParameterMode
	Defexpr ast.Node
}

func (n *FunctionParameter) Pos() int {
	return 0
}
