package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateEnumStmt struct {
	TypeName *ast.List
	Vals     *ast.List
}

func (n *CreateEnumStmt) Pos() int {
	return 0
}
