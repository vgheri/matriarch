package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestBuildCluster(t *testing.T) {

	clusters := []*Cluster{
		// 2 shards cluster
		{
			Shards: []*Shard{
				{
					Host:          "localhost",
					Port:          5432,
					Name:          "ecommerce/-80",
					keyspaceStart: 0x0000000000000000,
					keyspaceEnd:   0x8000000000000000,
				},
				{
					Host:          "localhost",
					Port:          5433,
					Name:          "ecommerce/80-",
					keyspaceStart: 0x8000000000000000,
					keyspaceEnd:   0xFFFFFFFFFFFFFFFF,
				},
			},
		},
		// 4 shards cluster
		{
			Shards: []*Shard{
				{
					Host:          "localhost",
					Port:          5432,
					Name:          "ecommerce/-40",
					keyspaceStart: 0x0000000000000000,
					keyspaceEnd:   0x4000000000000000,
				},
				{
					Host:          "localhost",
					Port:          5433,
					Name:          "ecommerce/40-80",
					keyspaceStart: 0x4000000000000000,
					keyspaceEnd:   0x8000000000000000,
				},
				{
					Host:          "localhost",
					Port:          5434,
					Name:          "ecommerce/80-c0",
					keyspaceStart: 0x8000000000000000,
					keyspaceEnd:   0xc000000000000000,
				},
				{
					Host:          "localhost",
					Port:          5435,
					Name:          "ecommerce/c0-",
					keyspaceStart: 0xc000000000000000,
					keyspaceEnd:   0xFFFFFFFFFFFFFFFF,
				},
			},
		},
	}

	tests := []struct {
		name          string
		keyspaceName  string
		hosts         []string
		expected      *Cluster
		expectedError error
	}{
		// {
		// 	name:          "invalid port should error",
		// 	keyspaceName:  "ecommerce",
		// 	hosts:         []string{"localhost:abcd", "localhost:1234"},
		// 	expected:      nil,
		// 	expectedError: ,
		// },
		{
			name:          "malformed host and port should error",
			keyspaceName:  "ecommerce",
			hosts:         []string{"localhost:1234:123", "localhost:1234"},
			expected:      nil,
			expectedError: ErrBadHost,
		},
		{
			name:          "it should build the cluster of 2 shards",
			keyspaceName:  "ecommerce",
			hosts:         []string{"localhost", "localhost:5433"},
			expected:      clusters[0],
			expectedError: nil,
		},
		{
			name:          "it should build the cluster of 4 shards",
			keyspaceName:  "ecommerce",
			hosts:         []string{"localhost", "localhost:5433", "localhost:5434", "localhost:5435"},
			expected:      clusters[1],
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observed, err := buildCluster(tt.keyspaceName, tt.hosts)
			if tt.expectedError != nil {
				if !errors.As(err, &tt.expectedError) {
					t.Fatalf("expected error type %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected test to succeed, got error %v", err)
				}
				if len(tt.expected.Shards) != len(observed.Shards) {
					t.Fatalf("expected has %d shards, observed has %d instead", len(tt.expected.Shards), len(observed.Shards))
				}
				for i, s := range observed.Shards {
					if !reflect.DeepEqual(*tt.expected.Shards[i], *s) {
						t.Fatalf("expected shard %+v to equal %+v", s, tt.expected.Shards[i])
					}
				}
			}
		})
	}
}
