package proxy

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/go-kit/kit/log"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgmock"
	"github.com/jackc/pgproto3/v2"
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
	p.connectionClosed = true
	return p.frontendConn.Close()
}

func (p *PGMock) IsClosed() bool {
	return p.connectionClosed
}
