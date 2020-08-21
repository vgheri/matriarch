package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type TransactionStmt struct {
	Kind    TransactionStmtKind
	Options *ast.List
	Gid     *string
}

func (n *TransactionStmt) Pos() int {
	return 0
}
