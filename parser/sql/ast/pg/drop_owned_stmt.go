package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DropOwnedStmt struct {
	Roles    *ast.List
	Behavior DropBehavior
}

func (n *DropOwnedStmt) Pos() int {
	return 0
}
