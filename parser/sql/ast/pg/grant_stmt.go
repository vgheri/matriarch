package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type GrantStmt struct {
	IsGrant     bool
	Targtype    GrantTargetType
	Objtype     GrantObjectType
	Objects     *ast.List
	Privileges  *ast.List
	Grantees    *ast.List
	GrantOption bool
	Behavior    DropBehavior
}

func (n *GrantStmt) Pos() int {
	return 0
}
