package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateUserMappingStmt struct {
	User        *RoleSpec
	Servername  *string
	IfNotExists bool
	Options     *ast.List
}

func (n *CreateUserMappingStmt) Pos() int {
	return 0
}
