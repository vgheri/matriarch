package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreateDomainStmt struct {
	Domainname  *ast.List
	TypeName    *TypeName
	CollClause  *CollateClause
	Constraints *ast.List
}

func (n *CreateDomainStmt) Pos() int {
	return 0
}
