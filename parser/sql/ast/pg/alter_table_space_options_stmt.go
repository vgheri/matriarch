package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterTableSpaceOptionsStmt struct {
	Tablespacename *string
	Options        *ast.List
	IsReset        bool
}

func (n *AlterTableSpaceOptionsStmt) Pos() int {
	return 0
}
