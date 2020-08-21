package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DeleteStmt struct {
	Relation      *RangeVar
	UsingClause   *ast.List
	WhereClause   ast.Node
	ReturningList *ast.List
	WithClause    *WithClause
}

func (n *DeleteStmt) Pos() int {
	return 0
}
