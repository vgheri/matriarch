package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RenameStmt struct {
	RenameType   ObjectType
	RelationType ObjectType
	Relation     *RangeVar
	Object       ast.Node
	Subname      *string
	Newname      *string
	Behavior     DropBehavior
	MissingOk    bool
}

func (n *RenameStmt) Pos() int {
	return 0
}
