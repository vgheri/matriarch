package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ArrayCoerceExpr struct {
	Xpr          ast.Node
	Arg          ast.Node
	Elemfuncid   Oid
	Resulttype   Oid
	Resulttypmod int32
	Resultcollid Oid
	IsExplicit   bool
	Coerceformat CoercionForm
	Location     int
}

func (n *ArrayCoerceExpr) Pos() int {
	return n.Location
}
