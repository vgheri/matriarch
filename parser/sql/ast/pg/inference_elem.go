package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type InferenceElem struct {
	Xpr          ast.Node
	Expr         ast.Node
	Infercollid  Oid
	Inferopclass Oid
}

func (n *InferenceElem) Pos() int {
	return 0
}
