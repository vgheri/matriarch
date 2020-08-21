package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type WithClause struct {
	Ctes      *ast.List
	Recursive bool
	Location  int
}

func (n *WithClause) Pos() int {
	return n.Location
}
