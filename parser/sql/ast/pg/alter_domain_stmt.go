package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterDomainStmt struct {
	Subtype   byte
	TypeName  *ast.List
	Name      *string
	Def       ast.Node
	Behavior  DropBehavior
	MissingOk bool
}

func (n *AlterDomainStmt) Pos() int {
	return 0
}
