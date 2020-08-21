package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type TypeCast struct {
	Arg      ast.Node
	TypeName *TypeName
	Location int
}

func (n *TypeCast) Pos() int {
	return n.Location
}
