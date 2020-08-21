package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterPublicationStmt struct {
	Pubname      *string
	Options      *ast.List
	Tables       *ast.List
	ForAllTables bool
	TableAction  DefElemAction
}

func (n *AlterPublicationStmt) Pos() int {
	return 0
}
