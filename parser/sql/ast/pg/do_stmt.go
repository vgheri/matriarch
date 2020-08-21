package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DoStmt struct {
	Args *ast.List
}

func (n *DoStmt) Pos() int {
	return 0
}
