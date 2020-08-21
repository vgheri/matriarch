package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type TableSampleClause struct {
	Tsmhandler Oid
	Args       *ast.List
	Repeatable ast.Node
}

func (n *TableSampleClause) Pos() int {
	return 0
}
