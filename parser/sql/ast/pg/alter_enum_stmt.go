package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterEnumStmt struct {
	TypeName           *ast.List
	OldVal             *string
	NewVal             *string
	NewValNeighbor     *string
	NewValIsAfter      bool
	SkipIfNewValExists bool
}

func (n *AlterEnumStmt) Pos() int {
	return 0
}
