package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type NamedArgExpr struct {
	Xpr       ast.Node
	Arg       ast.Node
	Name      *string
	Argnumber int
	Location  int
}

func (n *NamedArgExpr) Pos() int {
	return n.Location
}
