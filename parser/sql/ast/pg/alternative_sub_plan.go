package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlternativeSubPlan struct {
	Xpr      ast.Node
	Subplans *ast.List
}

func (n *AlternativeSubPlan) Pos() int {
	return 0
}
