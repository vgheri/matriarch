package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CoerceToDomain struct {
	Xpr            ast.Node
	Arg            ast.Node
	Resulttype     Oid
	Resulttypmod   int32
	Resultcollid   Oid
	Coercionformat CoercionForm
	Location       int
}

func (n *CoerceToDomain) Pos() int {
	return n.Location
}
