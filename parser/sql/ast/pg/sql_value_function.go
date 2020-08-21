package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type SQLValueFunction struct {
	Xpr      ast.Node
	Op       SQLValueFunctionOp
	Type     Oid
	Typmod   int32
	Location int
}

func (n *SQLValueFunction) Pos() int {
	return n.Location
}
