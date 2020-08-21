package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type FieldSelect struct {
	Xpr          ast.Node
	Arg          ast.Node
	Fieldnum     AttrNumber
	Resulttype   Oid
	Resulttypmod int32
	Resultcollid Oid
}

func (n *FieldSelect) Pos() int {
	return 0
}
