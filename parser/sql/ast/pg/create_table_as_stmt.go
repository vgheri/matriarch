package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateTableAsStmt struct {
	Query        ast.Node
	Into         *IntoClause
	Relkind      ObjectType
	IsSelectInto bool
	IfNotExists  bool
}

func (n *CreateTableAsStmt) Pos() int {
	return 0
}
