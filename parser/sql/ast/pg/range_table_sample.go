package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RangeTableSample struct {
	Relation   ast.Node
	Method     *ast.List
	Args       *ast.List
	Repeatable ast.Node
	Location   int
}

func (n *RangeTableSample) Pos() int {
	return n.Location
}
