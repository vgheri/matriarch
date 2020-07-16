package proxy

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/jackc/pgmock"
	"github.com/jackc/pgproto3/v2"
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
