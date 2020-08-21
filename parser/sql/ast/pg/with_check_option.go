package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type WithCheckOption struct {
	Kind     WCOKind
	Relname  *string
	Polname  *string
	Qual     ast.Node
	Cascaded bool
}

func (n *WithCheckOption) Pos() int {
	return 0
}
