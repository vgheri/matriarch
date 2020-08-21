package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type InsertStmt struct {
	Relation         *RangeVar
	Cols             *ast.List
	SelectStmt       ast.Node
	OnConflictClause *OnConflictClause
	ReturningList    *ast.List
	WithClause       *WithClause
	Override         OverridingKind
}

func (n *InsertStmt) Pos() int {
	return 0
}
