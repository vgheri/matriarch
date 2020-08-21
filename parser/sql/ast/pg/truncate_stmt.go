package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type TruncateStmt struct {
	Relations   *ast.List
	RestartSeqs bool
	Behavior    DropBehavior
}

func (n *TruncateStmt) Pos() int {
	return 0
}
