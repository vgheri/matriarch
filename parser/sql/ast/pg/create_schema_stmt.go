package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateSchemaStmt struct {
	Schemaname  *string
	Authrole    *RoleSpec
	SchemaElts  *ast.List
	IfNotExists bool
}

func (n *CreateSchemaStmt) Pos() int {
	return 0
}
