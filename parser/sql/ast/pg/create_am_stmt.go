package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateAmStmt struct {
	Amname      *string
	HandlerName *ast.List
	Amtype      byte
}

func (n *CreateAmStmt) Pos() int {
	return 0
}
