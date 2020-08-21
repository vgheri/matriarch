package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterDefaultPrivilegesStmt struct {
	Options *ast.List
	Action  *GrantStmt
}

func (n *AlterDefaultPrivilegesStmt) Pos() int {
	return 0
}
