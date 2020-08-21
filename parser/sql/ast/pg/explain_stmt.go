package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ExplainStmt struct {
	Query   ast.Node
	Options *ast.List
}

func (n *ExplainStmt) Pos() int {
	return 0
}
