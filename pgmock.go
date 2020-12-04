package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgmock"
	"github.com/jackc/pgproto3/v2"
	pg_query "github.com/lfittl/pg_query_go"
	"github.com/vgheri/matriarch/parser/engine"
	"github.com/vgheri/matriarch/parser/sql/ast"
	"github.com/vgheri/matriarch/parser/sql/ast/pg"
)

type PGMock struct {
	backend          *pgproto3.Backend
	frontendConn     net.Conn
	connectionClosed bool
	logger           log.Logger
}

func NewMock(frontendConn net.Conn, logger log.Logger) *PGMock {
	backend := pgproto3.NewBackend(pgproto3.NewChunkReader(frontendConn), frontendConn)

	mock := &PGMock{
		backend:      backend,
		frontendConn: frontendConn,
		logger:       logger,
	}
	return mock
}

// psql -h localhost -p 15432 -U vgheri -w
// F {"Type":"StartupMessage","ProtocolVersion":196608,"Parameters":{"application_name":"Postico 1.5.13","client_encoding":"UNICODE","database":"vgheri","options":"-c extra_float_digits=3","user":"vgheri"}}
// B {}
// B {"Type":"ParameterStatus","Name":"application_name","Value":"Postico 1.5.13"}
// B {"Type":"ParameterStatus","Name":"client_encoding","Value":"UNICODE"}
// B {"Type":"ParameterStatus","Name":"DateStyle","Value":"ISO, MDY"}
// B {"Type":"ParameterStatus","Name":"integer_datetimes","Value":"on"}
// B {"Type":"ParameterStatus","Name":"IntervalStyle","Value":"postgres"}
// B {"Type":"ParameterStatus","Name":"is_superuser","Value":"on"}
// B {"Type":"ParameterStatus","Name":"server_encoding","Value":"UTF8"}
// B {"Type":"ParameterStatus","Name":"server_version","Value":"12.3"}
// B {"Type":"ParameterStatus","Name":"session_authorization","Value":"vgheri"}
// B {"Type":"ParameterStatus","Name":"standard_conforming_strings","Value":"on"}
// B {"Type":"ParameterStatus","Name":"TimeZone","Value":"Europe/Paris"}
// B {"Type":"BackendKeyData","ProcessID":17399,"SecretKey":1755195487}
// B {"Type":"ReadyForQuery","TxStatus":"I"}
// POSTICO does a SELECT VERSION() immediately after receiving the first OK, so the  server needs to respond with the following commands
// F {"Type":"Query","String":"SELECT VERSION()"}
// B {"Type":"RowDescription","Fields":[{"Name":"version","TableOID":0,"TableAttributeNumber":0,"DataTypeOID":25,"DataTypeSize":-1,"TypeModifier":-1,"Format":0}]}
// B {"Type":"DataRow","Values":[{"text":"PostgreSQL 12.3 on x86_64-apple-darwin16.7.0, compiled by Apple LLVM version 8.1.0 (clang-802.0.42), 64-bit"}]}
// B {"Type":"CommandComplete","CommandTag":"SELECT 1"}
// B {"Type":"ReadyForQuery","TxStatus":"I"}
///
func (m *PGMock) AcceptUnauthenticatedConnRequestSteps() error {
	buf, err := json.Marshal(pgproto3.AuthenticationOk{})
	if err != nil {
		return err
	}
	m.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
	steps := []pgmock.Step{
		pgmock.SendMessage(&pgproto3.AuthenticationOk{}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "client_encoding", Value: "UNICODE"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "DateStyle", Value: "ISO, MDY"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "integer_datetimes", Value: "on"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "IntervalStyle", Value: "postgres"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "is_superuser", Value: "on"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "server_encoding", Value: "UTF8"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "server_version", Value: "12.3"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "session_authorization", Value: "vgheri"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "standard_conforming_strings", Value: "on"}),
		pgmock.SendMessage(&pgproto3.ParameterStatus{Name: "TimeZone", Value: "Europe/Paris"}),
		pgmock.SendMessage(&pgproto3.BackendKeyData{ProcessID: 17399, SecretKey: 1755195487}),
		pgmock.SendMessage(&pgproto3.ReadyForQuery{TxStatus: 'I'}),
	}
	script := pgmock.Script{Steps: steps}
	err = script.Run(m.backend)
	if err != nil {
		return err
	}
	return nil
}

func (m *PGMock) ReadClientConn() error {
	startupMessage, err := m.backend.ReceiveStartupMessage()
	if err != nil {
		return err
	}

	buf, err := json.Marshal(startupMessage)
	if err != nil {
		return err
	}
	m.logger.Log("msg", fmt.Sprintf("F %s", string(buf)))

	if _, ok := startupMessage.(*pgproto3.SSLRequest); ok {
		_, err = m.frontendConn.Write([]byte("N"))
		if err != nil {
			return fmt.Errorf("error sending deny SSL request: %w", err)
		}
		m.ReadClientConn()
	}
	return nil
}

func (m *PGMock) HandleConnectionPhase() error {
	err := m.ReadClientConn()
	if err != nil {
		return fmt.Errorf("error reading client connection %w", err)
	}
	return m.AcceptUnauthenticatedConnRequestSteps()
}

// TODO Use this
func (m *PGMock) HandleStartup() error {
	startupMessage, err := m.backend.ReceiveStartupMessage()
	if err != nil {
		return fmt.Errorf("error receiving startup message: %w", err)
	}

	switch startupMessage.(type) {
	case *pgproto3.StartupMessage:
		buf := (&pgproto3.AuthenticationOk{}).Encode(nil)
		buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(buf)
		_, err = m.frontendConn.Write(buf)
		if err != nil {
			return fmt.Errorf("error sending ready for query: %w", err)
		}
	case *pgproto3.SSLRequest:
		_, err = m.frontendConn.Write([]byte("N"))
		if err != nil {
			return fmt.Errorf("error sending deny SSL request: %w", err)
		}
		return m.HandleStartup()
	default:
		return fmt.Errorf("unknown startup message: %#v", startupMessage)
	}

	return nil
}

func (m *PGMock) Receive() (pgproto3.FrontendMessage, error) {
	msg, err := m.backend.Receive()
	if err != nil {
		return nil, fmt.Errorf("cannot receive client message: %w", err)
	}
	buf, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal message into JSON: %w", err)
	}
	m.logger.Log("msg", fmt.Sprintf("F %s", string(buf)))
	return msg, nil
}

// SendError sends an error message to the SQL client.
// The function can manage either a generic error or a Postgres specific one
func (m *PGMock) SendError(err error) error {
	var sendErr error
	if pgErr, ok := err.(*pgconn.PgError); ok {
		if sendErr = m.SendPGSQLErrorMessage(pgErr); sendErr != nil {
			return sendErr
		}
	} else {
		if sendErr = m.SendMatriarchErrorMessage(err); sendErr != nil {
			return sendErr
		}
	}
	// Send ReadyForQuery
	readForQueryMsg := &pgproto3.ReadyForQuery{
		TxStatus: 'I',
	}
	// DEBUG
	buf, err := json.Marshal(readForQueryMsg)
	if err != nil {
		return err
	}
	m.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
	if err := m.backend.Send(readForQueryMsg); err != nil {
		return fmt.Errorf("cannot send ReadForQuery message to client: %w", err)
	}
	return nil
}

func (m *PGMock) SendPGSQLErrorMessage(err *pgconn.PgError) error {
	msg := &pgproto3.ErrorResponse{
		Severity:         err.Severity,
		Code:             err.Code,
		Message:          err.Message,
		Detail:           err.Detail,
		Hint:             err.Hint,
		Position:         err.Position,
		InternalPosition: err.InternalPosition,
		InternalQuery:    err.InternalQuery,
		Where:            err.Where,
		SchemaName:       err.SchemaName,
		TableName:        err.TableName,
		ColumnName:       err.ColumnName,
		DataTypeName:     err.DataTypeName,
		ConstraintName:   err.ConstraintName,
		File:             err.File,
		Line:             err.Line,
		Routine:          err.Routine,
	}
	return m.backend.Send(msg)
}

func (m *PGMock) SendMatriarchErrorMessage(err error) error {
	msg := &pgproto3.ErrorResponse{
		Severity: "ERROR",
		Code:     "XX000",
		Message:  err.Error(),
	}
	return m.backend.Send(msg)
}

func (m *PGMock) FinaliseExecuteSequence(command string, results []*pgconn.Result) error {
	for _, result := range results {
		// Send RowDescription and then DataRow messages
		if len(result.FieldDescriptions) > 0 {
			rowDescriptionMsg := &pgproto3.RowDescription{
				Fields: result.FieldDescriptions,
			}
			// DEBUG
			buf, err := json.Marshal(rowDescriptionMsg)
			if err != nil {
				return err
			}
			m.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
			if err := m.backend.Send(rowDescriptionMsg); err != nil {
				return fmt.Errorf("cannot send RowDescription message to client: %w", err)
			}
		}
		for _, row := range result.Rows {
			dataRowMsg := &pgproto3.DataRow{
				Values: row,
			}
			// DEBUG
			buf, err := json.Marshal(dataRowMsg)
			if err != nil {
				return err
			}
			m.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
			if err := m.backend.Send(dataRowMsg); err != nil {
				return fmt.Errorf("cannot send DataRow message to client: %w", err)
			}
		}
		// Send command complete
		var cmdCompleteMsg pgproto3.CommandComplete
		switch command {
		case "INSERT":
			cmdCompleteMsg.CommandTag = []byte(fmt.Sprintf("%s 0 %d", command, result.CommandTag.RowsAffected()))
		case "DELETE", "UPDATE":
			cmdCompleteMsg.CommandTag = []byte(fmt.Sprintf("%s %d", command, result.CommandTag.RowsAffected()))
		default:
			cmdCompleteMsg.CommandTag = []byte(fmt.Sprintf("%s %d", command, result.CommandTag.RowsAffected()))
		}
		// DEBUG
		buf, err := json.Marshal(&cmdCompleteMsg)
		if err != nil {
			return err
		}
		m.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
		if err := m.backend.Send(&cmdCompleteMsg); err != nil {
			return fmt.Errorf("cannot send CommandComplete message to client: %w", err)
		}
	}
	// Send ReadyForQuery
	readForQueryMsg := &pgproto3.ReadyForQuery{
		TxStatus: 'I',
	}
	// DEBUG
	buf, err := json.Marshal(readForQueryMsg)
	if err != nil {
		return err
	}
	m.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
	if err := m.backend.Send(readForQueryMsg); err != nil {
		return fmt.Errorf("cannot send ReadForQuery message to client: %w", err)
	}
	return nil
}

func (p *PGMock) Close() error {
	closeNotificationMsg := &pgproto3.ErrorResponse{
		Severity:         "FATAL",
		Code:             "57P01",
		Message:          "terminating connection due to administrator command",
		Detail:           "",
		Hint:             "",
		Position:         0,
		InternalPosition: 0,
		InternalQuery:    "",
		Where:            "",
		SchemaName:       "",
		TableName:        "",
		ColumnName:       "",
		DataTypeName:     "",
		ConstraintName:   "",
		File:             "",
		Line:             0,
		Routine:          "ProcessInterrupts",
	}
	buf, err := json.Marshal(closeNotificationMsg)
	if err != nil {
		return err
	}
	p.logger.Log("msg", fmt.Sprintf("B %s", string(buf)))
	if err := p.backend.Send(closeNotificationMsg); err != nil {
		return fmt.Errorf("cannot send CloseComplete message to client: %w", err)
	}
	p.connectionClosed = true
	return p.frontendConn.Close()
}

func (p *PGMock) IsClosed() bool {
	return p.connectionClosed
}

// Process executes the SQL contained in the message by choosing the appropriate shard
func (mock *PGMock) Process(msg pgproto3.FrontendMessage, cluster *Cluster, vschema *Vschema) error {
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
		mock.logger.Log("msg", res)
		stmts, err := engine.NewParser().Parse(strings.NewReader(q.String))
		if err != nil {
			return fmt.Errorf("cannot parse frontend Query message: %w", err)
		}
		for _, stmt := range stmts {
			switch s := stmt.Raw.Stmt.(type) {
			case *pg.InsertStmt:
				if err = mock.processInsertStmt(s, q, cluster, vschema); err != nil {
					return err
				}
			case *pg.DeleteStmt:
				if err = mock.processDeleteStmt(s, q, cluster, vschema); err != nil {
					return err
				}
			case *pg.UpdateStmt:
				if err = mock.processUpdateStmt(s, q, cluster, vschema); err != nil {
					return err
				}
			case *pg.SelectStmt:
				if err = mock.processSelectStmt(s, q, cluster, vschema); err != nil {
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
func (mock *PGMock) processInsertStmt(s *pg.InsertStmt, q QueryMessage, cluster *Cluster, vschema *Vschema) error {
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
	mock.logger.Log("msg", fmt.Sprintf("relation: %s, columns: %s", relation, columns))
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
				mock.logger.Log("msg", fmt.Sprintf("shard selected: %s", target.Name))
				res, err := target.Conn.Exec(context.Background(), q.String)
				if err != nil {
					return err
				}
				defer res.Close()
				results, err := res.ReadAll()
				if err != nil {
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
func (mock *PGMock) processDeleteStmt(s *pg.DeleteStmt, q QueryMessage, cluster *Cluster, vschema *Vschema) error {
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
	mock.logger.Log("msg", fmt.Sprintf("relation: %s, columns: %s", relation, whereClauseColumns))
	var concat string
	for _, val := range indexes {
		whereClauseValue := whereClauseValues[val]
		concat = appendToConcatenate(concat, whereClauseValue)
	}
	target, err := cluster.GetShardForKeyspaceId(concat)
	if err != nil {
		return fmt.Errorf("cannot select destination shard for delete statement: %w", err)
	}
	mock.logger.Log("msg", fmt.Sprintf("shard selected: %s\n", target.Name))
	res, err := target.Conn.Exec(context.Background(), q.String)
	if err != nil {
		return err
	}
	defer res.Close()
	results, err := res.ReadAll()
	if err != nil {
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
func (mock *PGMock) processUpdateStmt(s *pg.UpdateStmt, q QueryMessage, cluster *Cluster, vschema *Vschema) error {
	var updatedColumns []string
	for _, argItem := range s.TargetList.Items {
		switch arg := argItem.(type) {
		case *pg.ResTarget:
			updatedColumns = append(updatedColumns, *arg.Name)
		default:
			return fmt.Errorf("unkown type in target list. Expected ResTarget, found %T", arg)
		}
	}
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
		if stringArrayContainsValue(updatedColumns, pc) > -1 {
			return fmt.Errorf("cannot update column %s because it is part of the primary vindex", pc)
		}
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
	mock.logger.Log("msg", fmt.Sprintf("relation: %s, columns: %s", relation, whereClauseColumns))
	var concat string
	for _, val := range indexes {
		whereClauseValue := whereClauseValues[val]
		concat = appendToConcatenate(concat, whereClauseValue)
	}
	target, err := cluster.GetShardForKeyspaceId(concat)
	if err != nil {
		return fmt.Errorf("cannot select destination shard for UPDATE statement: %w", err)
	}
	mock.logger.Log("msg", fmt.Sprintf("shard selected: %s\n", target.Name))
	res, err := target.Conn.Exec(context.Background(), q.String)
	if err != nil {
		return err
	}
	defer res.Close()
	results, err := res.ReadAll()
	if err != nil {
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

func parseWhereAExpression(node *pg.A_Expr, relations []string) (table, column, value string, err error) {
	if node.Kind != 0 {
		return table, column, value, fmt.Errorf("expression kind must be 0")
	}
	switch lexpr := node.Lexpr.(type) {
	case *pg.ColumnRef:
		var fields []string
		for _, col := range lexpr.Fields.Items {
			switch columnElem := col.(type) {
			case *pg.String:
				fields = append(fields, columnElem.Str)
			default:
				return table, column, value, fmt.Errorf("left expression of where clause must contain a column name")
			}
		}
		switch len(fields) {
		case 1:
			table = relations[0]
			column = fields[0]
		case 2:
			table = fields[0]
			column = fields[1]
		default:
			return table, column, value, fmt.Errorf("only the form table.column_name is currently supported to specify clauses")
		}
		for _, expr := range node.Name.Items {
			switch exprElem := expr.(type) {
			case *pg.String:
				if table == relations[0] && exprElem.Str != "=" {
					return table, column, value,
						fmt.Errorf("only equal expression is allowed in WHERE clauses for columns part of the primary vindex of the first table of the SELECT FROM statement")
				}
			}
		}
		switch rexpr := node.Rexpr.(type) {
		case *pg.A_Const:
			switch rexprConst := rexpr.Val.(type) {
			case *pg.String:
				value = rexprConst.Str
			case *pg.Float:
				value = rexprConst.Str
			case *pg.Integer:
				value = fmt.Sprintf("%d", rexprConst.Ival)
			default:
				return table, column, value, fmt.Errorf("unknown constant type in where clause. Only type string and integer are supported")
			}
		}
	}
	return table, column, value, nil
}

func walkWhereExpressionTree(node ast.Node, relations []string, whereClauseColumns, whereClauseValues map[string][]string) error {
	switch ss := node.(type) {
	case *pg.BoolExpr:
		for _, argItem := range ss.Args.Items {
			var err error
			switch arg := argItem.(type) {
			case *pg.A_Expr:
				table, column, value, err := parseWhereAExpression(arg, relations)
				if err != nil {
					return err
				}
				// 0 = AND, 1 = OR, 2 NOT. See https://doxygen.postgresql.org/primnodes_8h.html#a27f637bf3e2c33cc8e48661a8864c7af
				if ss.Boolop > 0 && table == relations[0] {
					return fmt.Errorf("Only AND expression is allowed as a boolean operator in the WHERE clause")
				}
				whereClauseColumns[table] = append(whereClauseColumns[table], column)
				whereClauseValues[table] = append(whereClauseValues[table], value)
			case *pg.BoolExpr:
				err = walkWhereExpressionTree(arg, relations, whereClauseColumns, whereClauseValues)
			}
			if err != nil {
				return err
			}
		}
	case *pg.A_Expr:
		table, column, value, err := parseWhereAExpression(ss, relations)
		if err != nil {
			return err
		}
		whereClauseColumns[table] = append(whereClauseColumns[table], column)
		whereClauseValues[table] = append(whereClauseValues[table], value)
	default:
		return fmt.Errorf("expecting a list of columns in where clause, found unknown expression")
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
func (mock *PGMock) processSelectStmt(s *pg.SelectStmt, q QueryMessage, cluster *Cluster, vschema *Vschema) error {
	var relations []string
	for _, fromClause := range s.FromClause.Items {
		walkJoinExpressionTree(fromClause, &relations)
	}
	var whereClauseColumns = make(map[string][]string)
	var whereClauseValues = make(map[string][]string)
	err := walkWhereExpressionTree(s.WhereClause, relations, whereClauseColumns, whereClauseValues)
	if err != nil {
		return err
	}
	mock.logger.Log("msg", fmt.Sprintf("where clause columns %s, values %s", whereClauseColumns, whereClauseValues))

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
	mock.logger.Log("msg", fmt.Sprintf("shard selected: %s", target.Name))

	res, err := target.Conn.Exec(context.Background(), q.String)
	if err != nil {
		return err
	}
	defer res.Close()
	results, err := res.ReadAll()
	if err != nil {
		return err
	}
	return mock.FinaliseExecuteSequence("SELECT", results)
}

// stringArrayContainsValue returns the index of the first occurence of the value in the array,
// -1 otherwise
func stringArrayContainsValue(a []string, v string) int {
	for i, val := range a {
		if val == v {
			return i
		}
	}
	return -1
}
