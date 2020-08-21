package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RangeTblFunction struct {
	Funcexpr          ast.Node
	Funccolcount      int
	Funccolnames      *ast.List
	Funccoltypes      *ast.List
	Funccoltypmods    *ast.List
	Funccolcollations *ast.List
	Funcparams        []uint32
}

func (n *RangeTblFunction) Pos() int {
	return 0
}
