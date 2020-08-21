package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterFdwStmt struct {
	Fdwname     *string
	FuncOptions *ast.List
	Options     *ast.List
}

func (n *AlterFdwStmt) Pos() int {
	return 0
}
