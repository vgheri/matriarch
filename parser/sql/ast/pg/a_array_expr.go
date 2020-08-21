package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type A_ArrayExpr struct {
	Elements *ast.List
	Location int
}

func (n *A_ArrayExpr) Pos() int {
	return n.Location
}
