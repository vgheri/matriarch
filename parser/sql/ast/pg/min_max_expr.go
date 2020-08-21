package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type MinMaxExpr struct {
	Xpr          ast.Node
	Minmaxtype   Oid
	Minmaxcollid Oid
	Inputcollid  Oid
	Op           MinMaxOp
	Args         *ast.List
	Location     int
}

func (n *MinMaxExpr) Pos() int {
	return n.Location
}
