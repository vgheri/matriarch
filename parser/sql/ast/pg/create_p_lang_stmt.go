package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreatePLangStmt struct {
	Replace     bool
	Plname      *string
	Plhandler   *ast.List
	Plinline    *ast.List
	Plvalidator *ast.List
	Pltrusted   bool
}

func (n *CreatePLangStmt) Pos() int {
	return 0
}
