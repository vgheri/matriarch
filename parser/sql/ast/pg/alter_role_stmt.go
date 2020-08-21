package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterRoleStmt struct {
	Role    *RoleSpec
	Options *ast.List
	Action  int
}

func (n *AlterRoleStmt) Pos() int {
	return 0
}
