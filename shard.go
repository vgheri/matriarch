package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Cluster struct {
	Shards []*Shard
}

type Shard struct {
	Host          string
	Port          int
	Name          string
	keyspaceStart uint64
	keyspaceEnd   uint64
}

const defaultPostgreSQLPort = 5432

var ErrBadHost = errors.New("bad host configuration, please specify host:port")
var ErrShardCount = errors.New("shard count must be a power of two")

func buildCluster(keyspaceName string, hosts []string) (*Cluster, error) {
	count := len(hosts)
	if count&(count-1) != 0 {
		return nil, ErrShardCount
	}
	shardRange := 0xFFFFFFFFFFFFFFFF / uint64(count)
	fmt.Printf("Shard range %d\n", shardRange)
	shards := []*Shard{}
	var start, end uint64
	start = 0x0000000000000000
	for i, host := range hosts {
		var port int
		var err error
		hostAndPort := strings.Split(host, ":")
		if len(hostAndPort) == 1 {
			port = defaultPostgreSQLPort
		} else if len(hostAndPort) > 2 {
			return nil, ErrBadHost
		} else {
			port, err = strconv.Atoi(hostAndPort[1])
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
			name = fmt.Sprintf("%s/-%s", keyspaceName, endStr[:2])
		} else if i == count-1 {
			startStr := strconv.FormatUint(start, 16)
			name = fmt.Sprintf("%s/%s-", keyspaceName, startStr[:2])
		} else {
			startStr := strconv.FormatUint(start, 16)
			endStr := strconv.FormatUint(end, 16)
			name = fmt.Sprintf("%s/%s-%s", keyspaceName, startStr[:2], endStr[:2])
		}
		shard := &Shard{
			Host:          hostAndPort[0],
			Port:          port,
			Name:          name,
			keyspaceStart: start,
			keyspaceEnd:   end,
		}
		shards = append(shards, shard)
		start = end
	}
	return &Cluster{Shards: shards}, nil
}

// func main() {
// 	shardOne := shard{name: "-40", min: 0x0000000000000000, max: 0x4000000000000000}
// 	shardTwo := shard{name: "40-80", min: 0x4000000000000000, max: 0x8000000000000000}
// 	shardThree := shard{name: "80-c0", min: 0x8000000000000000, max: 0xc000000000000000}
// 	shardFour := shard{name: "c0-", min: 0xc000000000000000, max: math.MaxUint64}

// 	shards := []shard{}
// 	shards = append(shards, shardOne, shardTwo, shardThree, shardFour)

// 	crc64Table := crc64.MakeTable(0xC96C5795D7870F42)
// 	target := crc64.Checksum([]byte("The quick brown fox jumps over the lazy dog."), crc64Table)
// 	tgtToString := strconv.FormatUint(target, 16)
// 	for _, s := range shards {
// 		if target >= s.min && target < s.max {
// 			fmt.Printf("Target %s belongs to shard %s\n", tgtToString, s.name)
// 			break
// 		}
// 	}

// }
