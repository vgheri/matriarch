package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateSubscriptionStmt struct {
	Subname     *string
	Conninfo    *string
	Publication *ast.List
	Options     *ast.List
}

func (n *CreateSubscriptionStmt) Pos() int {
	return 0
}
