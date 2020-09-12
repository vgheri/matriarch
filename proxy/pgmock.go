package proxy

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgmock"
	"github.com/jackc/pgproto3/v2"
	// pg_query "github.com/lfittl/pg_query_go"
	// nodes "github.com/lfittl/pg_query_go/nodes"
)

type PGMock struct {
	backend      *pgproto3.Backend
	frontendConn net.Conn
}

func NewMock(frontendConn net.Conn) *PGMock {
	backend := pgproto3.NewBackend(pgproto3.NewChunkReader(frontendConn), frontendConn)

	mock := &PGMock{
		backend:      backend,
		frontendConn: frontendConn,
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
	fmt.Println("B", string(buf))
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
	fmt.Println("F", string(buf))

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
	fmt.Println("F", string(buf))
	return msg, nil
}

func (m *PGMock) SendErrorMessage(err *pgconn.PgError) error {
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

func (m *PGMock) FinaliseInsertSequence(results []*pgconn.Result) {
	if results != nil {
		// Send RowDescription and then DataRow messages
		// B {"Type":"RowDescription","Fields":[{"Name":"id","TableOID":16386,"TableAttributeNumber":1,"DataTypeOID":23,"DataTypeSize":4,"TypeModifier":-1,"Format":0}]}
		// B {"Type":"DataRow","Values":[{"text":"5"}]}
	}
	// Send command complete
	// B {"Type":"CommandComplete","CommandTag":"INSERT 0 1"}
	cmdCompleteMsg := &pgproto3.CommandComplete{
		CommandTag: []byte("INSERT 0 1"),
	}
	m.backend.Send(cmdCompleteMsg)
	// Send ReadyForQuery
	// B {"Type":"ReadyForQuery","TxStatus":"I"}
	readForQueryMsg := &pgproto3.ReadyForQuery{
		TxStatus: 'I',
	}
	m.backend.Send(readForQueryMsg)
}

// func (m *PGMock) Process(msg pgproto3.FrontendMessage) error {
// 	buf, err := json.Marshal(msg)
// 	if err != nil {
// 		return fmt.Errorf("cannot marshal message into JSON: %w", err)
// 	}
// 	switch msg.(type) {
// 	case *pgproto3.Query:
// 		var q QueryMessage
// 		err = json.NewDecoder(strings.NewReader(string(buf))).Decode(&q)
// 		if err != nil {
// 			return fmt.Errorf("cannot decode frontend Query message into QueryMessage struct: %w", err)
// 		}
// 		fmt.Printf("Received query %s\n", q.String)
// 		stmts, err := engine.NewParser().Parse(strings.NewReader(q.String))
// 		if err != nil {
// 			return fmt.Errorf("cannot parse frontend Query message: %w", err)
// 		}
// 		for _, stmt := range stmts {
// 			switch s := stmt.Raw.Stmt.(type) {

// 			case *pg.InsertStmt:
// 				relation := *s.Relation.Relname
// 				var columns []string
// 				for _, item := range s.Cols.Items {
// 					t := item.(*pg.ResTarget)
// 					columns = append(columns, *t.Name)
// 				}
// 				fmt.Printf("Relation: %s, columns: %s\n", relation, columns)
// 				switch ss := s.SelectStmt.(type) {
// 				case *pg.SelectStmt:
// 					for _, v := range ss.ValuesLists.Items {
// 						switch t := v.(type) {
// 						case *ast.List:
// 							for i, vv := range t.Items {
// 								switch tt := vv.(type) {
// 								case *pg.A_Const:
// 									switch vt := tt.Val.(type) {
// 									case *pg.Integer:
// 										fmt.Printf("Value for item %d, %d\n", i, vt.Ival)
// 									case *pg.String:
// 										fmt.Printf("Value for item %d, %s\n", i, vt.Str)
// 									case *pg.Null:
// 										fmt.Printf("Item %d has null value\n", i)
// 									default:
// 										fmt.Printf("Item %d, value %+v\n", i, vv)
// 									}
// 								default:
// 									fmt.Printf("Unknown type in InsertStmt->SelectStmt->ValuesList.Items %#v\n", t)
// 								}
// 							}
// 						default:
// 							fmt.Printf("Unknown type in InsertStmt->SelectStmt->ValuesList %#v\n", t)
// 						}

// 					}
// 				default:
// 					fmt.Printf("Unknown type in InsertStmt->SelectStmt %#v\n", ss)
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

func (p *PGMock) Close() error {
	return p.frontendConn.Close()
}
