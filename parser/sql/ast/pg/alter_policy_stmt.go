package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type AlterPolicyStmt struct {
	PolicyName *string
	Table      *RangeVar
	Roles      *ast.List
	Qual       ast.Node
	WithCheck  ast.Node
}

func (n *AlterPolicyStmt) Pos() int {
	return 0
}
