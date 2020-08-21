package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ReassignOwnedStmt struct {
	Roles   *ast.List
	Newrole *RoleSpec
}

func (n *ReassignOwnedStmt) Pos() int {
	return 0
}
