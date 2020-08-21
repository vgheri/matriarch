package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterTableStmt struct {
	Relation  *RangeVar
	Cmds      *ast.List
	Relkind   ObjectType
	MissingOk bool
}

func (n *AlterTableStmt) Pos() int {
	return 0
}
