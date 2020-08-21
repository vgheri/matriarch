package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type DefineStmt struct {
	Kind        ObjectType
	Oldstyle    bool
	Defnames    *ast.List
	Args        *ast.List
	Definition  *ast.List
	IfNotExists bool
}

func (n *DefineStmt) Pos() int {
	return 0
}
