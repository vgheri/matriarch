package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RelabelType struct {
	Xpr           ast.Node
	Arg           ast.Node
	Resulttype    Oid
	Resulttypmod  int32
	Resultcollid  Oid
	Relabelformat CoercionForm
	Location      int
}

func (n *RelabelType) Pos() int {
	return n.Location
}
