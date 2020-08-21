package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ColumnRef struct {
	Fields   *ast.List
	Location int
}

func (n *ColumnRef) Pos() int {
	return n.Location
}
