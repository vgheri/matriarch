package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ViewStmt struct {
	View            *RangeVar
	Aliases         *ast.List
	Query           ast.Node
	Replace         bool
	Options         *ast.List
	WithCheckOption ViewCheckOption
}

func (n *ViewStmt) Pos() int {
	return 0
}
