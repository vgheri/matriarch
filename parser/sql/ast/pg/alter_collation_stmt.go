package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterCollationStmt struct {
	Collname *ast.List
}

func (n *AlterCollationStmt) Pos() int {
	return 0
}
