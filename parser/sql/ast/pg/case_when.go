package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CaseWhen struct {
	Xpr      ast.Node
	Expr     ast.Node
	Result   ast.Node
	Location int
}

func (n *CaseWhen) Pos() int {
	return n.Location
}
