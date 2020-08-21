package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CaseExpr struct {
	Xpr        ast.Node
	Casetype   Oid
	Casecollid Oid
	Arg        ast.Node
	Args       *ast.List
	Defresult  ast.Node
	Location   int
}

func (n *CaseExpr) Pos() int {
	return n.Location
}
