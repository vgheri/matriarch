package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateRoleStmt struct {
	StmtType RoleStmtType
	Role     *string
	Options  *ast.List
}

func (n *CreateRoleStmt) Pos() int {
	return 0
}
