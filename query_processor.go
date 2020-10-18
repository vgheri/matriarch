package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"

	pg_query "github.com/lfittl/pg_query_go"
	// nodes "github.com/lfittl/pg_query_go/nodes"
	"github.com/vgheri/matriarch/parser/engine"
	"github.com/vgheri/matriarch/parser/sql/ast"
	"github.com/vgheri/matriarch/parser/sql/ast/pg"
	"github.com/vgheri/matriarch/proxy"
)

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
		res, _ := pg_query.ParseToJSON(q.String)
		fmt.Println(res)
		stmts, err := engine.NewParser().Parse(strings.NewReader(q.String))
		if err != nil {
			return fmt.Errorf("cannot parse frontend Query message: %w", err)
		}
		for _, stmt := range stmts {
			switch s := stmt.Raw.Stmt.(type) {
			case *pg.InsertStmt:
				if err = processInsertStmt(s, q, mock, cluster, vschema); err != nil {
					return err
				}
			case *pg.DeleteStmt:
				if err = processDeleteStmt(s, q, mock, cluster, vschema); err != nil {
					return err
				}
			case *pg.UpdateStmt:
				if err = processUpdateStmt(s, q, mock, cluster, vschema); err != nil {
					return err
				}
			case *pg.SelectStmt:
				if err = processSelectStmt(s, q, mock, cluster, vschema); err != nil {
					return err
				}
			default:
				return fmt.Errorf("Unknown statement %s", q.String)
			}
		}
	}
	return nil
}

func appendToConcatenate(concat, val string) string {
	if concat == "" {
		return val
	}
	return fmt.Sprintf("%s&%s", concat, val)
}

// Limitations: primary vindex columns must be present in the list of values to insert
func processInsertStmt(s *pg.InsertStmt, q QueryMessage, mock *proxy.PGMock, cluster *Cluster, vschema *Vschema) error {
	relation := *s.Relation.Relname
	var columns []string
	for _, item := range s.Cols.Items {
		t := item.(*pg.ResTarget)
		columns = append(columns, *t.Name)
	}
	// build list of indexes of insert stmt columns that match the vschema table primary vindex columns.
	// e.g. insert into orders(id, user_id, total_amount, order_date) -> primary vindex for table orders is `id`,
	// so the result will be [0].
	// Iterate first on table vindex columns, and  then on insert stmt columns
	table := vschema.GetTable(relation)
	if table == nil {
		return fmt.Errorf("cannot process message, table %s is not part of the vschema", relation)
	}
	var indexes []int
	primaryIndexColumns := table.GetPrimaryVIndex().Columns
	for _, pc := range primaryIndexColumns {
		for i, c := range columns {
			if pc == c {
				indexes = append(indexes, i)
			}
		}
	}
	if len(indexes) != len(primaryIndexColumns) {
		return fmt.Errorf("cannot insert row without all primary vindex columns being present in the insert statement")
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
						return fmt.Errorf("unknown type in InsertStmt->SelectStmt->ValuesList.Items %#v", tt)
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
					if pgErr, ok := err.(*pgconn.PgError); ok {
						mock.SendErrorMessage(pgErr)
						return nil
					}
					return err
				}
				return mock.FinaliseExecuteSequence("INSERT", results)
			default:
				return fmt.Errorf("unknown type in InsertStmt->SelectStmt->ValuesList %#v", t)
			}
		}
	default:
		return fmt.Errorf("unknown type in InsertStmt->SelectStmt %#v", ss)
	}
	return fmt.Errorf("unknown error processing InsertStmt")
}

// Limitations: all primary vindex columns must be present and must be the only fields part of the where clause list of columns
// Column names must be used as left expression (i.e. order_id = '123342')
// Expressions allowed in the where clause: =
// where clause columns are linked by a AND boolean expression (i.e. column_1 = '123342 AND column_2 = 'abcd')
// Algo:
// 1. extract all columns and their values involved in the where clause
// 2. extract the boolean expression linking all clauses. If not "AND", return error
// 2. extract the expression:
// 3. if expr is
//    3.1 =, build the concatenate, select the shard and issue the delete command
//    3.2 in, for on each value and treat each iteration as a = expression -> not yet supported
func processDeleteStmt(s *pg.DeleteStmt, q QueryMessage, mock *proxy.PGMock, cluster *Cluster, vschema *Vschema) error {
	relation := *s.Relation.Relname
	var whereClauseColumns []string
	var whereClauseValues []string
	switch ss := s.WhereClause.(type) {
	case *pg.A_Expr:
		for _, expr := range ss.Name.Items {
			switch exprElem := expr.(type) {
			case *pg.String:
				if exprElem.Str != "=" {
					return fmt.Errorf("only equal expression is allowed in DELETE statements")
				}
			}
		}
		switch lexpr := ss.Lexpr.(type) {
		case *pg.ColumnRef:
			for _, column := range lexpr.Fields.Items {
				switch columnElem := column.(type) {
				case *pg.String:
					whereClauseColumns = append(whereClauseColumns, columnElem.Str)
				default:
					return fmt.Errorf("left expression of where clause must be a column name")
				}
			}
		}
		switch rexpr := ss.Rexpr.(type) {
		case *pg.A_Const:
			switch rexprConst := rexpr.Val.(type) {
			case *pg.String:
				whereClauseValues = append(whereClauseValues, rexprConst.Str)
			case *pg.Integer:
				whereClauseValues = append(whereClauseValues, fmt.Sprintf("%d", rexprConst.Ival))
			default:
				return fmt.Errorf("unknown constant type in where clause. String and Integer only")
			}
		}
	case *pg.BoolExpr:
		// 0 = AND, 1 = OR, 2 NOT. See https://doxygen.postgresql.org/primnodes_8h.html#a27f637bf3e2c33cc8e48661a8864c7af
		if ss.Boolop > 0 {
			return fmt.Errorf("Only AND expression is allowed as a boolean operator in DELETE statements")
		}
		for _, argItem := range ss.Args.Items {
			switch arg := argItem.(type) {
			case *pg.A_Expr:
				for _, expr := range arg.Name.Items {
					switch exprElem := expr.(type) {
					case *pg.String:
						if exprElem.Str != "=" {
							return fmt.Errorf("only equal expression is allowed in DELETE statements")
						}
					}
				}
				switch lexpr := arg.Lexpr.(type) {
				case *pg.ColumnRef:
					for _, column := range lexpr.Fields.Items {
						switch columnElem := column.(type) {
						case *pg.String:
							whereClauseColumns = append(whereClauseColumns, columnElem.Str)
						default:
							return fmt.Errorf("left expression of where clause must be a column name")
						}
					}
				}
				switch rexpr := arg.Rexpr.(type) {
				case *pg.A_Const:
					switch rexprConst := rexpr.Val.(type) {
					case *pg.String:
						whereClauseValues = append(whereClauseValues, rexprConst.Str)
					case *pg.Integer:
						whereClauseValues = append(whereClauseValues, fmt.Sprintf("%d", rexprConst.Ival))
					default:
						return fmt.Errorf("unknown constant type in where clause. String and Integer only")
					}
				}
			}
		}
	default:
		return fmt.Errorf("expecting a list of columns in where clause, found unknown expression")
	}
	// build list of indexes of delete stmt columns that match the vschema table primary vindex columns.
	// e.g. delete from orders(id, user_id, total_amount, order_date) -> primary vindex for table orders is `id`,
	// so the result will be [0].
	// Iterate first on table vindex columns, and then on insert stmt columns
	table := vschema.GetTable(relation)
	if table == nil {
		return fmt.Errorf("cannot process delete statement, table %s is not part of the vschema", relation)
	}
	var indexes []int
	for _, pc := range table.GetPrimaryVIndex().Columns {
		for i, c := range whereClauseColumns {
			if pc == c {
				indexes = append(indexes, i)
				continue
			}
		}
	}
	if len(indexes) != len(whereClauseColumns) {
		return fmt.Errorf("cannot execute delete statement without all primary vindex columns being present in the where clause")
	}
	fmt.Printf("DEBUG: Relation: %s, columns: %s\n", relation, whereClauseColumns)
	var concat string
	for _, val := range indexes {
		whereClauseValue := whereClauseValues[val]
		concat = appendToConcatenate(concat, whereClauseValue)
	}
	target, err := cluster.GetShardForKeyspaceId(concat)
	if err != nil {
		return fmt.Errorf("cannot select destination shard for delete statement: %w", err)
	}
	fmt.Printf("DEBUG: Shard selected: %s\n", target.Name)
	res := target.Conn.Exec(context.Background(), q.String)
	defer res.Close()
	results, err := res.ReadAll()
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			mock.SendErrorMessage(pgErr)
			return nil
		}
		return err
	}
	return mock.FinaliseExecuteSequence("DELETE", results)
}

// Limitations: all primary vindex columns must be present and must be the only fields part of the where clause list of columns
// Column names must be used as left expression (i.e. order_id = '123342')
// Expressions allowed in the where clause: =
// where clause columns are linked by a AND boolean expression (i.e. column_1 = '123342 AND column_2 = 'abcd')
// Algo:
// 1. extract all columns and their values involved in the where clause
// 2. extract the boolean expression linking all clauses. If not "AND", return error
// 2. extract the expression:
// 3. if expr is
//    3.1 =, build the concatenate, select the shard and issue the delete command
//    3.2 in, for on each value and treat each iteration as a = expression -> not yet supported
func processUpdateStmt(s *pg.UpdateStmt, q QueryMessage, mock *proxy.PGMock, cluster *Cluster, vschema *Vschema) error {
	relation := *s.Relation.Relname
	var whereClauseColumns []string
	var whereClauseValues []string
	switch ss := s.WhereClause.(type) {
	case *pg.A_Expr:
		for _, expr := range ss.Name.Items {
			switch exprElem := expr.(type) {
			case *pg.String:
				if exprElem.Str != "=" {
					return fmt.Errorf("only equal expression is allowed in UPDATE statements")
				}
			}
		}
		switch lexpr := ss.Lexpr.(type) {
		case *pg.ColumnRef:
			for _, column := range lexpr.Fields.Items {
				switch columnElem := column.(type) {
				case *pg.String:
					whereClauseColumns = append(whereClauseColumns, columnElem.Str)
				default:
					return fmt.Errorf("left expression of where clause must be a column name")
				}
			}
		}
		switch rexpr := ss.Rexpr.(type) {
		case *pg.A_Const:
			switch rexprConst := rexpr.Val.(type) {
			case *pg.String:
				whereClauseValues = append(whereClauseValues, rexprConst.Str)
			case *pg.Integer:
				whereClauseValues = append(whereClauseValues, fmt.Sprintf("%d", rexprConst.Ival))
			default:
				return fmt.Errorf("unknown constant type in where clause. String and Integer only")
			}
		}
	case *pg.BoolExpr:
		// 0 = AND, 1 = OR, 2 NOT. See https://doxygen.postgresql.org/primnodes_8h.html#a27f637bf3e2c33cc8e48661a8864c7af
		if ss.Boolop > 0 {
			return fmt.Errorf("Only AND expression is allowed as a boolean operator in UPDATE statements")
		}
		for _, argItem := range ss.Args.Items {
			switch arg := argItem.(type) {
			case *pg.A_Expr:
				for _, expr := range arg.Name.Items {
					switch exprElem := expr.(type) {
					case *pg.String:
						if exprElem.Str != "=" {
							return fmt.Errorf("only equal expression is allowed in UPDATE statements")
						}
					}
				}
				switch lexpr := arg.Lexpr.(type) {
				case *pg.ColumnRef:
					for _, column := range lexpr.Fields.Items {
						switch columnElem := column.(type) {
						case *pg.String:
							whereClauseColumns = append(whereClauseColumns, columnElem.Str)
						default:
							return fmt.Errorf("left expression of where clause must be a column name")
						}
					}
				}
				switch rexpr := arg.Rexpr.(type) {
				case *pg.A_Const:
					switch rexprConst := rexpr.Val.(type) {
					case *pg.String:
						whereClauseValues = append(whereClauseValues, rexprConst.Str)
					case *pg.Integer:
						whereClauseValues = append(whereClauseValues, fmt.Sprintf("%d", rexprConst.Ival))
					default:
						return fmt.Errorf("unknown constant type in where clause. String and Integer only")
					}
				}
			}
		}
	default:
		return fmt.Errorf("expecting a list of columns in where clause, found unknown expression")
	}
	// build list of indexes of update stmt columns that match the vschema table primary vindex columns.
	// e.g. update orders set amount = 500 where id = 'abcd' -> primary vindex for table orders is `id`,
	// so the result will be [0].
	// Iterate first on table vindex columns, and then on update stmt where clause columns
	table := vschema.GetTable(relation)
	if table == nil {
		return fmt.Errorf("cannot process UPDATE statement, table %s is not part of the vschema", relation)
	}
	var indexes []int
	for _, pc := range table.GetPrimaryVIndex().Columns {
		for i, c := range whereClauseColumns {
			if pc == c {
				indexes = append(indexes, i)
				continue
			}
		}
	}
	if len(indexes) != len(whereClauseColumns) {
		return fmt.Errorf("cannot execute UPDATE statement without all primary vindex columns being present in the where clause")
	}
	fmt.Printf("DEBUG: Relation: %s, columns: %s\n", relation, whereClauseColumns)
	var concat string
	for _, val := range indexes {
		whereClauseValue := whereClauseValues[val]
		concat = appendToConcatenate(concat, whereClauseValue)
	}
	target, err := cluster.GetShardForKeyspaceId(concat)
	if err != nil {
		return fmt.Errorf("cannot select destination shard for UPDATE statement: %w", err)
	}
	fmt.Printf("DEBUG: Shard selected: %s\n", target.Name)
	res := target.Conn.Exec(context.Background(), q.String)
	defer res.Close()
	results, err := res.ReadAll()
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			mock.SendErrorMessage(pgErr)
			return nil
		}
		return err
	}
	return mock.FinaliseExecuteSequence("UPDATE", results)
}

func walkJoinExpressionTree(node ast.Node, relations *[]string) error {
	switch fc := node.(type) {
	case *pg.RangeVar:
		*relations = append(*relations, *fc.Relname)
	case *pg.JoinExpr:
		switch jlarg := fc.Larg.(type) {
		case *pg.RangeVar:
			*relations = append(*relations, *jlarg.Relname)
		case *pg.JoinExpr:
			walkJoinExpressionTree(jlarg, relations)
		default:
			return fmt.Errorf("cannot parse FROM clause for SELECT JOIN statement")
		}
		switch jrarg := fc.Rarg.(type) {
		case *pg.RangeVar:
			*relations = append(*relations, *jrarg.Relname)
		default:
			return fmt.Errorf("cannot parse FROM clause for SELECT JOIN statement")
		}
	}
	return nil
}

// Limitations: all primary vindex columns must be present and must be the only fields part of the where clause list of columns
// Column names must be used as left expression (i.e. order_id = '123342')
// Expressions allowed in the where clause: =
// where clause columns are linked by a AND boolean expression (i.e. column_1 = '123342 AND column_2 = 'abcd')
// Algo:
// 1. extract the tables involved in the select operation
// 2. extract all columns and their values involved in the where clause
// 3. extract the boolean expression linking all clauses. If not "AND", return error
// 4. extract the expression:
// 5. if expr is
//    5.1 =, build the concatenate, select the shard and issue the delete command
//    5.2 in, for on each value and treat each iteration as a = expression -> not yet supported
func processSelectStmt(s *pg.SelectStmt, q QueryMessage, mock *proxy.PGMock, cluster *Cluster, vschema *Vschema) error {
	var relations []string
	// for _, fromClause := range s.FromClause.Items {
	// 	switch fc := fromClause.(type) {
	// 	case *pg.RangeVar:
	// 		relations = append(relations, *fc.Relname)
	// 	case *pg.JoinExpr:
	// 		switch jlarg := fc.Larg.(type) {
	// 		case *pg.RangeVar:
	// 			relations = append(relations, *jlarg.Relname)
	// 		default:
	// 			return fmt.Errorf("cannot parse FROM clause for SELECT JOIN statement")
	// 		}
	// 		switch jrarg := fc.Rarg.(type) {
	// 		case *pg.RangeVar:
	// 			relations = append(relations, *jrarg.Relname)
	// 		default:
	// 			return fmt.Errorf("cannot parse FROM clause for SELECT JOIN statement")
	// 		}
	// 	}
	// }
	for _, fromClause := range s.FromClause.Items {
		walkJoinExpressionTree(fromClause, &relations)
	}
	var whereClauseColumns = make(map[string][]string)
	var whereClauseValues = make(map[string][]string)
	switch ss := s.WhereClause.(type) {
	case *pg.A_Expr:
		var tableName string
		for _, expr := range ss.Name.Items {
			switch exprElem := expr.(type) {
			case *pg.String:
				if exprElem.Str != "=" {
					return fmt.Errorf("only equal expression is allowed in DELETE statements")
				}
			}
		}
		switch lexpr := ss.Lexpr.(type) {
		case *pg.ColumnRef:
			var fields []string
			for _, column := range lexpr.Fields.Items {
				switch columnElem := column.(type) {
				case *pg.String:
					fields = append(fields, columnElem.Str)
				default:
					return fmt.Errorf("left expression of where clause must be contain a column name")
				}
			}
			switch len(fields) {
			case 1:
				tableName = relations[0]
				whereClauseColumns[tableName] = append(whereClauseColumns[tableName], fields[0])
			case 2:
				tableName = fields[0]
				if relations[0] != tableName {
					return fmt.Errorf("the first where clause should be related to the first table in the FROM list")
				}
				whereClauseColumns[tableName] = append(whereClauseColumns[tableName], fields[1])
			default:
				return fmt.Errorf("to specify clauses, the form table.column_name is only supported")
			}
		}
		switch rexpr := ss.Rexpr.(type) {
		case *pg.A_Const:
			switch rexprConst := rexpr.Val.(type) {
			case *pg.String:
				whereClauseValues[tableName] = append(whereClauseValues[tableName], rexprConst.Str)
			case *pg.Integer:
				whereClauseValues[tableName] = append(whereClauseValues[tableName], fmt.Sprintf("%d", rexprConst.Ival))
			default:
				return fmt.Errorf("unknown constant type in where clause. String and Integer only")
			}
		}
	case *pg.BoolExpr:
		// 0 = AND, 1 = OR, 2 NOT. See https://doxygen.postgresql.org/primnodes_8h.html#a27f637bf3e2c33cc8e48661a8864c7af
		if ss.Boolop > 0 {
			return fmt.Errorf("Only AND expression is allowed as a boolean operator in the WHERE clause")
		}
		var tableName string
		for i, argItem := range ss.Args.Items {
			switch arg := argItem.(type) {
			case *pg.A_Expr:
				switch lexpr := arg.Lexpr.(type) {
				case *pg.ColumnRef:
					var fields []string
					for _, column := range lexpr.Fields.Items {
						switch columnElem := column.(type) {
						case *pg.String:
							fields = append(fields, columnElem.Str)
						default:
							return fmt.Errorf("left expression of where clause must contain a column name")
						}
					}
					switch len(fields) {
					case 1:
						tableName = relations[0]
						whereClauseColumns[tableName] = append(whereClauseColumns[tableName], fields[0])
					case 2:
						tableName = fields[0]
						if i == 0 && relations[0] != tableName {
							return fmt.Errorf("the first where clause should be related to the first table in the FROM list")
						}
						whereClauseColumns[tableName] = append(whereClauseColumns[tableName], fields[1])
					default:
						return fmt.Errorf("to specify clauses, the form table.column_name is only supported")
					}
				}
				for _, expr := range arg.Name.Items {
					switch exprElem := expr.(type) {
					case *pg.String:
						if tableName == relations[0] && exprElem.Str != "=" {
							return fmt.Errorf("only equal expression is allowed in the WHERE clause")
						}
					}
				}
				switch rexpr := arg.Rexpr.(type) {
				case *pg.A_Const:
					switch rexprConst := rexpr.Val.(type) {
					case *pg.String:
						whereClauseValues[tableName] = append(whereClauseValues[tableName], rexprConst.Str)
					case *pg.Integer:
						whereClauseValues[tableName] = append(whereClauseValues[tableName], fmt.Sprintf("%d", rexprConst.Ival))
					default:
						return fmt.Errorf("unknown constant type in where clause. String and Integer only")
					}
				}
			}
		}
	default:
		return fmt.Errorf("expecting a list of columns in where clause, found unknown expression")
	}
	// build list of indexes of select stmt where clause columns that match the vschema table primary vindex columns.
	// e.g. select * from orders where id = 'abcd' -> primary vindex for table orders is `id`,
	// so the result will be [0].
	// Iterate first on table vindex columns, and then on select stmt columns
	table := vschema.GetTable(relations[0])
	if table == nil {
		return fmt.Errorf("cannot process select statement, table %s is not part of the vschema", relations[0])
	}
	var indexes []int
	for _, pc := range table.GetPrimaryVIndex().Columns {
		for i, c := range whereClauseColumns[relations[0]] {
			if pc == c {
				indexes = append(indexes, i)
				continue
			}
		}
	}
	if len(indexes) != len(whereClauseColumns[relations[0]]) {
		return fmt.Errorf("cannot execute select statement without all primary vindex columns being present in the where clause")
	}
	var concat string
	for _, val := range indexes {
		whereClauseValue := whereClauseValues[relations[0]][val]
		concat = appendToConcatenate(concat, whereClauseValue)
	}
	target, err := cluster.GetShardForKeyspaceId(concat)
	if err != nil {
		return fmt.Errorf("cannot select destination shard for delete statement: %w", err)
	}
	fmt.Printf("DEBUG: Shard selected: %s\n", target.Name)
	res := target.Conn.Exec(context.Background(), q.String)
	defer res.Close()
	results, err := res.ReadAll()
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			mock.SendErrorMessage(pgErr)
			return nil
		}
		return err
	}
	return mock.FinaliseExecuteSequence("SELECT", results)
}
