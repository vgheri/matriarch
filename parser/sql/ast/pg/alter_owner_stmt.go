package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterOwnerStmt struct {
	ObjectType ObjectType
	Relation   *RangeVar
	Object     ast.Node
	Newowner   *RoleSpec
}

func (n *AlterOwnerStmt) Pos() int {
	return 0
}
