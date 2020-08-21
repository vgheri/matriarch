package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type MultiAssignRef struct {
	Source   ast.Node
	Colno    int
	Ncolumns int
}

func (n *MultiAssignRef) Pos() int {
	return 0
}
