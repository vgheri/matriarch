package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateForeignTableStmt struct {
	Base       *CreateStmt
	Servername *string
	Options    *ast.List
}

func (n *CreateForeignTableStmt) Pos() int {
	return 0
}
