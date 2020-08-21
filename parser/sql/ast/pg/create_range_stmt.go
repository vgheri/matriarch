package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateRangeStmt struct {
	TypeName *ast.List
	Params   *ast.List
}

func (n *CreateRangeStmt) Pos() int {
	return 0
}
