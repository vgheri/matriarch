package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CoalesceExpr struct {
	Xpr            ast.Node
	Coalescetype   Oid
	Coalescecollid Oid
	Args           *ast.List
	Location       int
}

func (n *CoalesceExpr) Pos() int {
	return n.Location
}
