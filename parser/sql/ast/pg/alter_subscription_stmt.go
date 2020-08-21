package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterSubscriptionStmt struct {
	Kind        AlterSubscriptionType
	Subname     *string
	Conninfo    *string
	Publication *ast.List
	Options     *ast.List
}

func (n *AlterSubscriptionStmt) Pos() int {
	return 0
}
