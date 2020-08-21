package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterForeignServerStmt struct {
	Servername *string
	Version    *string
	Options    *ast.List
	HasVersion bool
}

func (n *AlterForeignServerStmt) Pos() int {
	return 0
}
