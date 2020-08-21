package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type PartitionSpec struct {
	Strategy   *string
	PartParams *ast.List
	Location   int
}

func (n *PartitionSpec) Pos() int {
	return n.Location
}
