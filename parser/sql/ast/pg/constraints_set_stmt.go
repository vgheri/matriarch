package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ConstraintsSetStmt struct {
	Constraints *ast.List
	Deferred    bool
}

func (n *ConstraintsSetStmt) Pos() int {
	return 0
}
