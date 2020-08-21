package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type GroupingSet struct {
	Kind     GroupingSetKind
	Content  *ast.List
	Location int
}

func (n *GroupingSet) Pos() int {
	return n.Location
}
