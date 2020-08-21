package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type SetToDefault struct {
	Xpr       ast.Node
	TypeId    Oid
	TypeMod   int32
	Collation Oid
	Location  int
}

func (n *SetToDefault) Pos() int {
	return n.Location
}
