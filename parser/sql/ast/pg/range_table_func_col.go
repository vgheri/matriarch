package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RangeTableFuncCol struct {
	Colname       *string
	TypeName      *TypeName
	ForOrdinality bool
	IsNotNull     bool
	Colexpr       ast.Node
	Coldefexpr    ast.Node
	Location      int
}

func (n *RangeTableFuncCol) Pos() int {
	return n.Location
}
