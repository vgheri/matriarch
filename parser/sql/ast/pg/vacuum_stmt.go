package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type VacuumStmt struct {
	Options  int
	Relation *RangeVar
	VaCols   *ast.List
}

func (n *VacuumStmt) Pos() int {
	return 0
}
