package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type BoolExpr struct {
	Xpr      ast.Node
	Boolop   BoolExprType
	Args     *ast.List
	Location int
}

func (n *BoolExpr) Pos() int {
	return n.Location
}
