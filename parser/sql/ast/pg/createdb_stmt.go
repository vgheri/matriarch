package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreatedbStmt struct {
	Dbname  *string
	Options *ast.List
}

func (n *CreatedbStmt) Pos() int {
	return 0
}
