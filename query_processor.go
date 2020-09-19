package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"

	// pg_query "github.com/lfittl/pg_query_go"
	// nodes "github.com/lfittl/pg_query_go/nodes"
	"github.com/vgheri/matriarch/parser/engine"
	"github.com/vgheri/matriarch/parser/sql/ast"
	"github.com/vgheri/matriarch/parser/sql/ast/pg"
	"github.com/vgheri/matriarch/proxy"
)

// This should be further split into 2 distinct functions: parse and execute, where
// parse returns classic go errors, and execute returns sql errors
func Process(msg pgproto3.FrontendMessage, mock *proxy.PGMock, cluster *Cluster, vschema *Vschema) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("cannot marshal message into JSON: %w", err)
	}
	switch msg.(type) {
	case *pgproto3.Terminate:
		return mock.Close()
	case *pgproto3.Query:
		var q QueryMessage
		err = json.NewDecoder(strings.NewReader(string(buf))).Decode(&q)
		if err != nil {
			return fmt.Errorf("cannot decode frontend Query message into QueryMessage struct: %w", err)
		}
		stmts, err := engine.NewParser().Parse(strings.NewReader(q.String))
		if err != nil {
			return fmt.Errorf("cannot parse frontend Query message: %w", err)
		}
		for _, stmt := range stmts {
			switch s := stmt.Raw.Stmt.(type) {
			case *pg.InsertStmt:
				relation := *s.Relation.Relname
				var columns []string
				for _, item := range s.Cols.Items {
					t := item.(*pg.ResTarget)
					columns = append(columns, *t.Name)
				}
				// TODO build list of indexes of insert stmt columns that match the vschema table primary vindex columns.
				// e.g. insert into orders(id, user_id, total_amount, order_date) -> primary vindex for table orders is `id`,
				// so the result will be [0].
				// Iterate first on table vindex columns, and  then on insert stmt columns
				table := vschema.GetTable(relation)
				if table == nil {
					return fmt.Errorf("cannot process message, table %s is not part of the vschema", relation)
				}
				var indexes []int
				for _, pc := range table.GetPrimaryVIndex().Columns {
					for i, c := range columns {
						if pc == c {
							indexes = append(indexes, i)
						}
					}
				}
				fmt.Printf("DEBUG: Relation: %s, columns: %s\n", relation, columns)
				switch ss := s.SelectStmt.(type) {
				case *pg.SelectStmt:
					for _, v := range ss.ValuesLists.Items {
						switch t := v.(type) {
						case *ast.List:
							var concat string
							for _, val := range indexes {
								node := t.Items[val]
								switch tt := node.(type) {
								case *pg.A_Const:
									switch vt := tt.Val.(type) {
									case *pg.Integer:
										concat = appendToConcatenate(concat, fmt.Sprintf("%d", vt.Ival))
									case *pg.String:
										concat = appendToConcatenate(concat, vt.Str)
									case *pg.Null:
										return errors.New("cannot insert row with null value for column part of a primary VIndex")
									default:
										return errors.New("cannot insert row with unknown value for column part of a primary VIndex")
									}
								default:
									return fmt.Errorf("unknown type in InsertStmt->SelectStmt->ValuesList.Items %#v\n", tt)
								}
							}
							target, err := cluster.GetShardForKeyspaceId(concat)
							if err != nil {
								return fmt.Errorf("cannot select destination shard for insert statement: %w", err)
							}
							fmt.Printf("DEBUG: Shard selected: %s\n", target.Name)
							res := target.Conn.Exec(context.Background(), q.String)
							defer res.Close()
							results, err := res.ReadAll()
							if err != nil {
								fmt.Printf("cannot read insert statement result: %v", err)
								if pgErr, ok := err.(*pgconn.PgError); ok {
									mock.SendErrorMessage(pgErr)
									return nil
								}
								return err
							}
							mock.FinaliseInsertSequence(results)
							return nil
						default:
							fmt.Printf("Unknown type in InsertStmt->SelectStmt->ValuesList %#v\n", t)
						}
					}
				default:
					fmt.Printf("Unknown type in InsertStmt->SelectStmt %#v\n", ss)
				}
			}
		}
	}
	return nil
}

func appendToConcatenate(concat, val string) string {
	if concat == "" {
		return val
	}
	return fmt.Sprintf("%s&s", concat, val)
}
