package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CommentStmt struct {
	Objtype ObjectType
	Object  ast.Node
	Comment *string
}

func (n *CommentStmt) Pos() int {
	return 0
}
