package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type SecLabelStmt struct {
	Objtype  ObjectType
	Object   ast.Node
	Provider *string
	Label    *string
}

func (n *SecLabelStmt) Pos() int {
	return 0
}
