package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CollateExpr struct {
	Xpr      ast.Node
	Arg      ast.Node
	CollOid  Oid
	Location int
}

func (n *CollateExpr) Pos() int {
	return n.Location
}
