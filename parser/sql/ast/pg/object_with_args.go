package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type ObjectWithArgs struct {
	Objname         *ast.List
	Objargs         *ast.List
	ArgsUnspecified bool
}

func (n *ObjectWithArgs) Pos() int {
	return 0
}
