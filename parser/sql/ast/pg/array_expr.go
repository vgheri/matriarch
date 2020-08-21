package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ArrayExpr struct {
	Xpr           ast.Node
	ArrayTypeid   Oid
	ArrayCollid   Oid
	ElementTypeid Oid
	Elements      *ast.List
	Multidims     bool
	Location      int
}

func (n *ArrayExpr) Pos() int {
	return n.Location
}
