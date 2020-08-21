package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateExtensionStmt struct {
	Extname     *string
	IfNotExists bool
	Options     *ast.List
}

func (n *CreateExtensionStmt) Pos() int {
	return 0
}
