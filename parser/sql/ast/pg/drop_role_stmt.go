package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DropRoleStmt struct {
	Roles     *ast.List
	MissingOk bool
}

func (n *DropRoleStmt) Pos() int {
	return 0
}
