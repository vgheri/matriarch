package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type FromExpr struct {
	Fromlist *ast.List
	Quals    ast.Node
}

func (n *FromExpr) Pos() int {
	return 0
}
