package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DropStmt struct {
	Objects    *ast.List
	RemoveType ObjectType
	Behavior   DropBehavior
	MissingOk  bool
	Concurrent bool
}

func (n *DropStmt) Pos() int {
	return 0
}
