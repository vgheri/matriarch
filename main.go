package main

import (
	"flag"
	"fmt"

	// "log"
	// "net"
	"os"
	// "github.com/vgheri/matriarch/proxy"
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
	flag.StringVar(&options.hosts, "hosts", "127.0.0.1:5432,127.0.0.1:5433", "Comma separated list of PostgreSQL server addresses")
	flag.StringVar(&options.vschemaFilePath, "vschema", "vschema.json", "Vschema file path")
	flag.Parse()

	// read vschema file
	// vschema, err := readVschemaFile(options.vschemaFilePath)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// ln, err := net.Listen("tcp", options.listenAddress)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// clientConn, err := ln.Accept()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// listenNetwork := "tcp"
	// if _, err := os.Stat(options.remoteAddress); err == nil {
	// 	listenNetwork = "unix"
	// }

	// serverConn, err := net.Dial(listenNetwork, options.remoteAddress)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// proxy := proxy.NewProxy(clientConn, serverConn)
	// err = proxy.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
