package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CaseTestExpr struct {
	Xpr       ast.Node
	TypeId    Oid
	TypeMod   int32
	Collation Oid
}

func (n *CaseTestExpr) Pos() int {
	return 0
}
