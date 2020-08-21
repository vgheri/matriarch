package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateOpFamilyStmt struct {
	Opfamilyname *ast.List
	Amname       *string
}

func (n *CreateOpFamilyStmt) Pos() int {
	return 0
}
