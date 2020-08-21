package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type NullTest struct {
	Xpr          ast.Node
	Arg          ast.Node
	Nulltesttype NullTestType
	Argisrow     bool
	Location     int
}

func (n *NullTest) Pos() int {
	return n.Location
}
