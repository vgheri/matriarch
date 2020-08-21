package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type A_Indirection struct {
	Arg         ast.Node
	Indirection *ast.List
}

func (n *A_Indirection) Pos() int {
	return 0
}
