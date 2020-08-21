package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterOperatorStmt struct {
	Opername *ObjectWithArgs
	Options  *ast.List
}

func (n *AlterOperatorStmt) Pos() int {
	return 0
}
