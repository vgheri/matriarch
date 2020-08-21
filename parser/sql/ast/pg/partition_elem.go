package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type PartitionElem struct {
	Name      *string
	Expr      ast.Node
	Collation *ast.List
	Opclass   *ast.List
	Location  int
}

func (n *PartitionElem) Pos() int {
	return n.Location
}
