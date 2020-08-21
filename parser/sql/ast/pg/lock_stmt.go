package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type LockStmt struct {
	Relations *ast.List
	Mode      int
	Nowait    bool
}

func (n *LockStmt) Pos() int {
	return 0
}
