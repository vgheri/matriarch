package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"

	"github.com/vgheri/matriarch/proxy"
)

var options struct {
	listenAddress   string
	hosts           string
	vschemaFilePath string
	logLevel        string
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage:  %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&options.listenAddress, "listen", "127.0.0.1:15432", "Proxy listen address")
	flag.StringVar(&options.hosts, "hosts", "localhost:5432,localhost:5433", "Comma separated list of PostgreSQL server addresses, without empty spaces")
	flag.StringVar(&options.vschemaFilePath, "vschema", "vschema.json", "Vschema file path")
	flag.StringVar(&options.logLevel, "loglevel", "INFO", "Allowed levels: ALL, DEBUG, INFO, WARN, ERROR, NONE")
	flag.Parse()

	logger := configureLogger(options.logLevel)

	// read vschema file
	vschema, err := readVschemaFile(options.vschemaFilePath)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot read vschema file: %s", err.Error()))
		os.Exit(1)
	}

	hosts := strings.Split(options.hosts, ",")
	// 1. Create the cluster, opening a TCP connection with each shard
	cluster, err := NewCluster(vschema.Keyspace, hosts)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot create new Matriarch cluster: %s", err.Error()))
		os.Exit(1)
	}
	for i, s := range cluster.Shards {
		level.Info(logger).Log("msg", fmt.Sprintf("%d - Connected to %s - %s\n", i, s.Host, s.Name))
	}

	// 2. Start accepting connections from clients
	ln, err := net.Listen("tcp", options.listenAddress)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot listen on port %s: %s", options.listenAddress, err.Error()))
		os.Exit(1)
	}

	level.Info(logger).Log("msg", fmt.Sprintf("Matriarch started and listening on port %s", options.listenAddress))
	// main control loop
	for {
		// wait for a new client connection
		clientConn, err := ln.Accept()
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("cannot accept incoming client connection: %s", err.Error()))
			continue
		}
		scopedLogger := level.Debug(logger)
		if connId, err := uuid.NewRandom(); err == nil {
			scopedLogger = log.With(scopedLogger, "connection-id", connId.String())
		}
		// each client connection lifecycle is managed in its own goroutine
		go func(clientConn net.Conn) {
			mock := proxy.NewMock(clientConn, scopedLogger)
			err = mock.HandleConnectionPhase()
			if err != nil {
				level.Error(logger).Log("msg", fmt.Sprintf("cannot handle connection phase of incoming client connection: %s", err.Error()))
				return
			}
			for {
				msg, err := mock.Receive()
				if err != nil {
					level.Error(logger).Log("msg", fmt.Sprintf("cannot receive message from client: %s", err.Error()))
					mock.SendError(err)
					return
				}

				// For each incoming client connection, parse the query to identify the shard(s) involved and create a proxy for each backend involved, then send the query
				err = Process(msg, mock, cluster, vschema, level.Debug(logger))
				if err != nil {
					level.Error(logger).Log("msg", fmt.Sprintf("cannot process message from client: %s", err.Error()))
					mock.SendError(err)
					return
				}
				if mock.IsClosed() {
					return
				}
			}
		}(clientConn)
	}
}

func configureLogger(minimumLevel string) log.Logger {
	var logger log.Logger
	logger = log.NewLogfmtLogger(os.Stdout)
	var levelOption level.Option
	switch minimumLevel {
	case "NONE":
		levelOption = level.AllowNone()
	case "DEBUG":
		levelOption = level.AllowDebug()
	case "INFO":
		levelOption = level.AllowInfo()
	case "WARN":
		levelOption = level.AllowWarn()
	case "ERROR":
		levelOption = level.AllowError()
	default:
		levelOption = level.AllowAll()
	}
	logger = level.NewFilter(logger, levelOption)
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	return logger
}
