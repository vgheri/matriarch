package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type A_Indices struct {
	IsSlice bool
	Lidx    ast.Node
	Uidx    ast.Node
}

func (n *A_Indices) Pos() int {
	return 0
}
