package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterSeqStmt struct {
	Sequence    *RangeVar
	Options     *ast.List
	ForIdentity bool
	MissingOk   bool
}

func (n *AlterSeqStmt) Pos() int {
	return 0
}
