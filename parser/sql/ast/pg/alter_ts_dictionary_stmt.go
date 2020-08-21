package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterTSDictionaryStmt struct {
	Dictname *ast.List
	Options  *ast.List
}

func (n *AlterTSDictionaryStmt) Pos() int {
	return 0
}
