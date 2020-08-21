package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateTableSpaceStmt struct {
	Tablespacename *string
	Owner          *RoleSpec
	Location       *string
	Options        *ast.List
}

func (n *CreateTableSpaceStmt) Pos() int {
	return 0
}
