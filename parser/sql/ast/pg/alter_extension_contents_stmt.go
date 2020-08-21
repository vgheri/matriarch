package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterExtensionContentsStmt struct {
	Extname *string
	Action  int
	Objtype ObjectType
	Object  ast.Node
}

func (n *AlterExtensionContentsStmt) Pos() int {
	return 0
}
