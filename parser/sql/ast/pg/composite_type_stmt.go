package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CompositeTypeStmt struct {
	Typevar    *RangeVar
	Coldeflist *ast.List
}

func (n *CompositeTypeStmt) Pos() int {
	return 0
}
