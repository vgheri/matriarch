package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateStatsStmt struct {
	Defnames    *ast.List
	StatTypes   *ast.List
	Exprs       *ast.List
	Relations   *ast.List
	IfNotExists bool
}

func (n *CreateStatsStmt) Pos() int {
	return 0
}
