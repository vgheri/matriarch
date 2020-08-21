package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterDatabaseStmt struct {
	Dbname  *string
	Options *ast.List
}

func (n *AlterDatabaseStmt) Pos() int {
	return 0
}
