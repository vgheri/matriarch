package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	// "log"
	"os"

	"github.com/vgheri/matriarch/proxy"
)

var options struct {
	listenAddress   string
	hosts           string
	vschemaFilePath string
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage:  %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&options.listenAddress, "listen", "127.0.0.1:15432", "Proxy listen address")
	flag.StringVar(&options.hosts, "hosts", "localhost:5432,localhost:5433", "Comma separated list of PostgreSQL server addresses, without empty spaces")
	flag.StringVar(&options.vschemaFilePath, "vschema", "vschema.json", "Vschema file path")
	flag.Parse()

	// read vschema file
	vschema, err := readVschemaFile(options.vschemaFilePath)
	if err != nil {
		log.Fatal(err)
	}

	hosts := strings.Split(options.hosts, ",")
	// 1. Create the cluster, opening a TCP connection with each shard
	cluster, err := NewCluster(vschema.Keyspace, hosts)
	if err != nil {
		log.Fatalf("Couldn't create new Matriarch cluster: %v", err)
	}
	for i, s := range cluster.Shards {
		fmt.Printf("%d - Connected to %s - %s\n", i, s.Host, s.Name)
	}

	// 2. Start accepting connections from clients
	ln, err := net.Listen("tcp", options.listenAddress)
	if err != nil {
		log.Fatal(err)
	}

	// TODO this should be inside a for loop, creating a goroutine  for each  client connection
	clientConn, err := ln.Accept()
	if err != nil {
		log.Fatal(err)
	}
	mock := proxy.NewMock(clientConn)
	err = mock.HandleConnectionPhase()
	if err != nil {
		log.Fatalf("unable to mock pgsql and accept the request: %v", err)
	}
	// 3. For each incoming client connection, parse the query to identify the shard(s) involved and create a proxy for each backend involved, then send the query
	msg, err := mock.Receive()
	if err != nil {
		log.Fatalf("cannot receive message from client: %v", err)
	}

	var isClosed bool
	for _, s := range cluster.Shards {
		if s.Conn.IsClosed() {
			isClosed = true
			fmt.Printf("Connection for shard %s is closed\n", s.Name)
		}
	}
	if isClosed {
		log.Fatal("One or more connections to backends are closed. Quitting...")
	}
	results, pgerr, err := Process(msg, cluster, vschema)
	if err != nil {
		log.Fatalf("cannot process message: %v", err)
	}
	if pgerr != nil {
		mock.SendErrorMessage(pgerr)
	}
	// 4. Retrieve result and send it back to the client
	mock.FinaliseInsertSequence(results)
	// proxy := proxy.NewProxy(clientConn, serverConn)
	// err = proxy.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	log.Fatal(mock.Close())
}
