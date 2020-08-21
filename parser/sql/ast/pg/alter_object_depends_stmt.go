package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterObjectDependsStmt struct {
	ObjectType ObjectType
	Relation   *RangeVar
	Object     ast.Node
	Extname    ast.Node
}

func (n *AlterObjectDependsStmt) Pos() int {
	return 0
}
