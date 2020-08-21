package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ExecuteStmt struct {
	Name   *string
	Params *ast.List
}

func (n *ExecuteStmt) Pos() int {
	return 0
}
