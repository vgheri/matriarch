package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CopyStmt struct {
	Relation  *RangeVar
	Query     ast.Node
	Attlist   *ast.List
	IsFrom    bool
	IsProgram bool
	Filename  *string
	Options   *ast.List
}

func (n *CopyStmt) Pos() int {
	return 0
}
