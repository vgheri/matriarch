package main

import (
	"context"
	"errors"
	"fmt"
	"hash/crc64"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/jackc/pgconn"
	"github.com/vgheri/matriarch/pgpool"
)

type Cluster struct {
	Shards []*Shard
}

type Shard struct {
	Host          string
	Name          string
	KeyspaceStart uint64
	KeyspaceEnd   uint64
	Conn          *pgpool.Pool
}

const defaultPostgreSQLPort = 5432

var ErrBadHost = errors.New("bad host configuration, please specify host:port")
var ErrShardCount = errors.New("shard count must be a power of two")
var ErrCannotFindTargetShard = errors.New("cannot find shard owner of desired keyspace id value")

func NewCluster(keyspaceName string, hosts []string) (*Cluster, error) {
	shards, err := buildShards(keyspaceName, hosts)
	if err != nil {
		return nil, err
	}
	err = connect(shards)
	if err != nil {
		return nil, err
	}
	return &Cluster{Shards: shards}, nil
}

func (c *Cluster) Shutdown() {
	for _, shard := range c.Shards {
		shard.Conn.Close()
	}
}

func (c *Cluster) Stats(logger log.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			for _, shard := range c.Shards {
				stats := shard.Conn.Stat()
				logger.Log("shard", shard.Name, "ac", stats.AcquiredConns(),
					"tac", stats.AcquireCount(),
					"eac", stats.EmptyAcquireCount(),
					"ic", stats.IdleConns(),
					"tc", stats.TotalConns())
			}
		}
	}
}

func buildShards(keyspaceName string, hosts []string) ([]*Shard, error) {
	count := len(hosts)
	if count&(count-1) != 0 {
		return nil, ErrShardCount
	}
	shardRange := 0xFFFFFFFFFFFFFFFF / uint64(count)
	shards := []*Shard{}
	var start, end uint64
	start = 0x0000000000000000
	for i, host := range hosts {
		hostAndPort := strings.Split(host, ":")
		if len(hostAndPort) == 1 {
			return nil, ErrBadHost
		} else if len(hostAndPort) > 2 {
			return nil, ErrBadHost
		} else {
			_, err := strconv.Atoi(hostAndPort[1])
			if err != nil {
				return nil, fmt.Errorf("cannot build cluster, invalid port number %s: %w", hostAndPort[1], err)
			}
		}
		var name string
		if i != count-1 {
			end = start + shardRange + 1
		} else {
			end = start + shardRange
		}
		if i == 0 {
			endStr := strconv.FormatUint(end, 16)
			name = fmt.Sprintf("%s_$%s", keyspaceName, endStr[:2])
		} else if i == count-1 {
			startStr := strconv.FormatUint(start, 16)
			name = fmt.Sprintf("%s_%s$", keyspaceName, startStr[:2])
		} else {
			startStr := strconv.FormatUint(start, 16)
			endStr := strconv.FormatUint(end, 16)
			name = fmt.Sprintf("%s_%s$%s", keyspaceName, startStr[:2], endStr[:2])
		}
		shard := &Shard{
			Host:          host,
			Name:          name,
			KeyspaceStart: start,
			KeyspaceEnd:   end,
		}
		shards = append(shards, shard)
		start = end
	}
	return shards, nil
}

// TODO parallelize
// For each shard,
//  1. Establish the connection (and send STARTUP message) -> next step open a pool of connections
// 	2. Send commands to check if DB exists already, otherwise create it
func connect(shards []*Shard) error {
	for _, shard := range shards {
		// 1. Establish the connection and send the startup message
		ctx := context.Background()
		pgConn, err := pgconn.Connect(ctx, fmt.Sprintf("postgres://%s", shard.Host))
		if err != nil {
			return fmt.Errorf("error trying to connect to %s: %w", shard.Host, err)
		}
		// 	2. Send commands to check if DB exists already, otherwise create it
		result := pgConn.ExecParams(ctx, "SELECT 1 FROM pg_database WHERE datname=$1", [][]byte{[]byte(shard.Name)}, nil, nil, nil)
		var dbAlreadyExists bool
		for result.NextRow() {
			res := string(result.Values()[0])
			if res == "1" {
				dbAlreadyExists = true
			}
			break
		}
		_, err = result.Close()
		if err != nil {
			return fmt.Errorf("error reading result from shard %s: %w", shard.Host, err)
		}
		if !dbAlreadyExists {
			command := fmt.Sprintf("CREATE DATABASE %s", shard.Name)
			result = pgConn.ExecParams(ctx, command, nil, nil, nil, nil)
			for result.NextRow() {
				return fmt.Errorf("shouldn't have received any result! Got %s", string(result.Values()[0]))
			}
			if result != nil {
				_, err = result.Close()
				if err != nil {
					return fmt.Errorf("error reading result from shard %s: %w", shard.Host, err)
				}
			}
		}
		_ = pgConn.Close(ctx)
		connString := fmt.Sprintf("postgres://%s/%s", shard.Host, shard.Name)
		pool, err := pgpool.Connect(ctx, connString)
		if err != nil {
			return fmt.Errorf("error trying to connect to %s/%s: %w", shard.Host, shard.Name, err)
		}
		shard.Conn = pool
	}
	return nil
}

var crc64Table = crc64.MakeTable(0xC96C5795D7870F42)

// GetShardForKeyspaceId returns the shard owning a specific keyspace id, which is
// calculated as the result crc64 checksum of the input string
func (c *Cluster) GetShardForKeyspaceId(value string) (*Shard, error) {
	keyspaceId := crc64.Checksum([]byte(value), crc64Table)
	for _, s := range c.Shards {
		if keyspaceId >= s.KeyspaceStart && keyspaceId < s.KeyspaceEnd {
			return s, nil
		}
	}
	return nil, ErrCannotFindTargetShard
}
