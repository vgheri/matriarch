package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type IndexElem struct {
	Name          *string
	Expr          ast.Node
	Indexcolname  *string
	Collation     *ast.List
	Opclass       *ast.List
	Ordering      SortByDir
	NullsOrdering SortByNulls
}

func (n *IndexElem) Pos() int {
	return 0
}
