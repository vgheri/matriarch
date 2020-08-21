package pg

import (
	"github.com/vgheri/matriarch/parser/sql/ast"
)

type RawStmt struct {
	Stmt         ast.Node
	StmtLocation int
	StmtLen      int
}

func (n *RawStmt) Pos() int {
	return 0
}
