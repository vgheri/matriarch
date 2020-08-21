package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateEventTrigStmt struct {
	Trigname   *string
	Eventname  *string
	Whenclause *ast.List
	Funcname   *ast.List
}

func (n *CreateEventTrigStmt) Pos() int {
	return 0
}
