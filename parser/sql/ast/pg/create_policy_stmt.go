package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type CreatePolicyStmt struct {
	PolicyName *string
	Table      *RangeVar
	CmdName    *string
	Permissive bool
	Roles      *ast.List
	Qual       ast.Node
	WithCheck  ast.Node
}

func (n *CreatePolicyStmt) Pos() int {
	return 0
}
