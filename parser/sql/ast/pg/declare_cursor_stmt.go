package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DeclareCursorStmt struct {
	Portalname *string
	Options    int
	Query      ast.Node
}

func (n *DeclareCursorStmt) Pos() int {
	return 0
}
