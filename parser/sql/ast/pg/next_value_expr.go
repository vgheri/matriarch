package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type NextValueExpr struct {
	Xpr    ast.Node
	Seqid  Oid
	TypeId Oid
}

func (n *NextValueExpr) Pos() int {
	return 0
}
