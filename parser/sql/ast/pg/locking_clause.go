package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type LockingClause struct {
	LockedRels *ast.List
	Strength   LockClauseStrength
	WaitPolicy LockWaitPolicy
}

func (n *LockingClause) Pos() int {
	return 0
}
