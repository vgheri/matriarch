package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateSeqStmt struct {
	Sequence    *RangeVar
	Options     *ast.List
	OwnerId     Oid
	ForIdentity bool
	IfNotExists bool
}

func (n *CreateSeqStmt) Pos() int {
	return 0
}
