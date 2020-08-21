package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RangeSubselect struct {
	Lateral  bool
	Subquery ast.Node
	Alias    *Alias
}

func (n *RangeSubselect) Pos() int {
	return 0
}
