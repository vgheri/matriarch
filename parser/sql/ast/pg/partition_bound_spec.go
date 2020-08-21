package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type PartitionBoundSpec struct {
	Strategy    byte
	Listdatums  *ast.List
	Lowerdatums *ast.List
	Upperdatums *ast.List
	Location    int
}

func (n *PartitionBoundSpec) Pos() int {
	return n.Location
}
