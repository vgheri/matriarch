package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type IntoClause struct {
	Rel            *RangeVar
	ColNames       *ast.List
	Options        *ast.List
	OnCommit       OnCommitAction
	TableSpaceName *string
	ViewQuery      ast.Node
	SkipData       bool
}

func (n *IntoClause) Pos() int {
	return 0
}
