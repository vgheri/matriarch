package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type Param struct {
	Xpr         ast.Node
	Paramkind   ParamKind
	Paramid     int
	Paramtype   Oid
	Paramtypmod int32
	Paramcollid Oid
	Location    int
}

func (n *Param) Pos() int {
	return n.Location
}
