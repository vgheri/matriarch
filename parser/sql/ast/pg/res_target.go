package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ResTarget struct {
	Name        *string
	Indirection *ast.List
	Val         ast.Node
	Location    int
}

func (n *ResTarget) Pos() int {
	return n.Location
}
