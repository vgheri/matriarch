package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ScalarArrayOpExpr struct {
	Xpr         ast.Node
	Opno        Oid
	Opfuncid    Oid
	UseOr       bool
	Inputcollid Oid
	Args        *ast.List
	Location    int
}

func (n *ScalarArrayOpExpr) Pos() int {
	return n.Location
}
