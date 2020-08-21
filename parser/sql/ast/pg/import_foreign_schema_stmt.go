package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ImportForeignSchemaStmt struct {
	ServerName   *string
	RemoteSchema *string
	LocalSchema  *string
	ListType     ImportForeignSchemaType
	TableList    *ast.List
	Options      *ast.List
}

func (n *ImportForeignSchemaStmt) Pos() int {
	return 0
}
