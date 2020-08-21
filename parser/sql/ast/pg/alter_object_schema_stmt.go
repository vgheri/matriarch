package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterObjectSchemaStmt struct {
	ObjectType ObjectType
	Relation   *RangeVar
	Object     ast.Node
	Newschema  *string
	MissingOk  bool
}

func (n *AlterObjectSchemaStmt) Pos() int {
	return 0
}
