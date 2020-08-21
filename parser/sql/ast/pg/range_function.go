package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RangeFunction struct {
	Lateral    bool
	Ordinality bool
	IsRowsfrom bool
	Functions  *ast.List
	Alias      *Alias
	Coldeflist *ast.List
}

func (n *RangeFunction) Pos() int {
	return 0
}
