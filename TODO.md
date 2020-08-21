TODO:

- [ ] Insert statement use case
  - [ ] query_processor.Process:
    - [ ] Properly handle error after Exec
    - [ ] Investigate how we can use either ExecParams either the low level connection to PGSQL by hijacking the connection
  - [ ] Send result to client via the mock backend.SendMessage to client
- [ ] main.go:56 clientConn, err := ln.Accept() should be part of a for loop and spawn go routines
