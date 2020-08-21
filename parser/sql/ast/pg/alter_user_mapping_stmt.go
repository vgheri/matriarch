package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterUserMappingStmt struct {
	User       *RoleSpec
	Servername *string
	Options    *ast.List
}

func (n *AlterUserMappingStmt) Pos() int {
	return 0
}
