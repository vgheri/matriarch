package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type BooleanTest struct {
	Xpr          ast.Node
	Arg          ast.Node
	Booltesttype BoolTestType
	Location     int
}

func (n *BooleanTest) Pos() int {
	return n.Location
}
