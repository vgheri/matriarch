package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DefElem struct {
	Defnamespace *string
	Defname      *string
	Arg          ast.Node
	Defaction    DefElemAction
	Location     int
}

func (n *DefElem) Pos() int {
	return n.Location
}
