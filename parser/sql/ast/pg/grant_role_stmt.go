package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type GrantRoleStmt struct {
	GrantedRoles *ast.List
	GranteeRoles *ast.List
	IsGrant      bool
	AdminOpt     bool
	Grantor      *RoleSpec
	Behavior     DropBehavior
}

func (n *GrantRoleStmt) Pos() int {
	return 0
}
