package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateOpClassItem struct {
	Itemtype    int
	Name        *ObjectWithArgs
	Number      int
	OrderFamily *ast.List
	ClassArgs   *ast.List
	Storedtype  *TypeName
}

func (n *CreateOpClassItem) Pos() int {
	return 0
}
