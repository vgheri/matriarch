package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterExtensionStmt struct {
	Extname *string
	Options *ast.List
}

func (n *AlterExtensionStmt) Pos() int {
	return 0
}
