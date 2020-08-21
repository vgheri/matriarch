package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type Alias struct {
	Aliasname *string
	Colnames  *ast.List
}

func (n *Alias) Pos() int {
	return 0
}
