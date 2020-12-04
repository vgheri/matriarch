package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/google/uuid"
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

	// Signals management
	signals := make(chan os.Signal, 1)
	quit := make(chan bool, 1)
	defer close(signals)
	defer close(quit)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go listenForSignals(signals, quit, level.Warn(logger))

	// read vschema file
	vschema, err := readVschemaFile(options.vschemaFilePath)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot read vschema file: %s", err.Error()))
		os.Exit(1)
	}

	hosts := strings.Split(options.hosts, ",")
	// Create the cluster, opening a TCP connection with each shard
	cluster, err := NewCluster(vschema.Keyspace, hosts)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot create new cluster: %s", err.Error()))
		os.Exit(1)
	}
	defer cluster.Shutdown()
	for _, s := range cluster.Shards {
		level.Info(logger).Log("msg", fmt.Sprintf("Connected to %s - %s\n", s.Host, s.Name))
	}

	// Start accepting connections from clients
	ln, err := net.Listen("tcp", options.listenAddress)
	if err != nil {
		level.Error(logger).Log("msg", fmt.Sprintf("cannot listen on port %s: %s", options.listenAddress, err.Error()))
		os.Exit(1)
	}
	// Close the listener
	var wg sync.WaitGroup
	var quitChannels []chan bool
	// Close the listener on quit signal and signal all goroutines processing client requests to terminate their inflight requests and then return
	go func() {
		<-quit
		level.Info(logger).Log("msg", "stop accepting incoming connections")
		if err = ln.Close(); err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("cannot stop accepting incoming connections: %s", err.Error()))
		}
		for _, c := range quitChannels {
			c <- true // TODO replace with close(c)
		}
	}()

	level.Info(logger).Log("msg", fmt.Sprintf("Matriarch started and listening on port %s", options.listenAddress))

	go cluster.Stats(level.Debug(logger))
	// main control loop
	for {
		// wait for a new client connection
		clientConn, err := ln.Accept()
		if err != nil {
			level.Error(logger).Log("msg", fmt.Sprintf("cannot accept incoming client connection: %s", err.Error()))
			break
		}
		scopedLogger := level.Debug(logger)
		connId, err := uuid.NewRandom()
		if err == nil {
			scopedLogger = log.With(scopedLogger, "connection-id", connId.String())
		}
		wg.Add(1)
		quitChan := make(chan bool)
		quitChannels = append(quitChannels, quitChan)
		// each client connection lifecycle is managed in its own goroutine
		go func(clientConn net.Conn, wg *sync.WaitGroup, q chan bool, logger log.Logger) {
			mock := NewMock(clientConn, logger)
			defer func() {
				if !mock.IsClosed() {
					if err := mock.Close(); err != nil {
						logger.Log("msg", fmt.Sprintf("cannot close client connection: %s", err.Error()))
					}
				}
				wg.Done()
			}()
			// Receive goroutine quit signal and close the mock
			go func() {
				<-q
				if !mock.IsClosed() {
					if err := mock.Close(); err != nil {
						logger.Log("msg", fmt.Sprintf("cannot close client connection: %s", err.Error()))
					}
				}
			}()
			err = mock.HandleConnectionPhase()
			if err != nil {
				logger.Log("msg", fmt.Sprintf("cannot handle connection phase of incoming new client connection: %s", err.Error()))
				return
			}
			for {
				msg, err := mock.Receive()
				if err != nil {
					logger.Log("msg", fmt.Sprintf("cannot receive message from client: %s", err.Error()))
					mock.SendError(err)
					return
				}
				// For each incoming client connection, parse the query to identify the shard(s) involved and create a proxy for each backend involved, then send the query
				err = mock.Process(msg, cluster, vschema)
				if err != nil {
					logger.Log("msg", fmt.Sprintf("cannot process message from client: %s", err.Error()))
					mock.SendError(err)
					return
				}
				if mock.IsClosed() {
					return
				}
			}
		}(clientConn, &wg, quitChan, scopedLogger)
	}

	level.Info(logger).Log("msg", "draining active connections")
	wg.Wait()
	level.Info(logger).Log("msg", "all active connections have been drained, now quitting")
	os.Exit(0)
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

func listenForSignals(signals chan os.Signal, quit chan bool, logger log.Logger) {
	sig := <-signals
	logger.Log("msg", fmt.Sprintf("received signal %s", sig.String()))
	quit <- true // TODO replace with close(quit)
}
