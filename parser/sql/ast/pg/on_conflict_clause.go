package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type OnConflictClause struct {
	Action      OnConflictAction
	Infer       *InferClause
	TargetList  *ast.List
	WhereClause ast.Node
	Location    int
}

func (n *OnConflictClause) Pos() int {
	return n.Location
}
