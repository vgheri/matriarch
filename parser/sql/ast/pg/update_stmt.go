package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type UpdateStmt struct {
	Relation      *RangeVar
	TargetList    *ast.List
	WhereClause   ast.Node
	FromClause    *ast.List
	ReturningList *ast.List
	WithClause    *WithClause
}

func (n *UpdateStmt) Pos() int {
	return 0
}
