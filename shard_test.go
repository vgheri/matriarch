package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestBuildShards(t *testing.T) {

	twoShards := []*Shard{
		{
			Host:          "localhost:5432",
			Name:          "ecommerce_$80",
			KeyspaceStart: 0x0000000000000000,
			KeyspaceEnd:   0x8000000000000000,
		},
		{
			Host:          "localhost:5433",
			Name:          "ecommerce_80$",
			KeyspaceStart: 0x8000000000000000,
			KeyspaceEnd:   0xFFFFFFFFFFFFFFFF,
		},
	}

	fourShards := []*Shard{
		{
			Host:          "localhost:5432",
			Name:          "ecommerce_$40",
			KeyspaceStart: 0x0000000000000000,
			KeyspaceEnd:   0x4000000000000000,
		},
		{
			Host:          "localhost:5433",
			Name:          "ecommerce_40$80",
			KeyspaceStart: 0x4000000000000000,
			KeyspaceEnd:   0x8000000000000000,
		},
		{
			Host:          "localhost:5434",
			Name:          "ecommerce_80$c0",
			KeyspaceStart: 0x8000000000000000,
			KeyspaceEnd:   0xc000000000000000,
		},
		{
			Host:          "localhost:5435",
			Name:          "ecommerce_c0$",
			KeyspaceStart: 0xc000000000000000,
			KeyspaceEnd:   0xFFFFFFFFFFFFFFFF,
		},
	}

	tests := []struct {
		name          string
		keyspaceName  string
		hosts         []string
		expected      []*Shard
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
			hosts:         []string{"localhost:5432", "localhost:5433"},
			expected:      twoShards,
			expectedError: nil,
		},
		{
			name:          "it should build the cluster of 4 shards",
			keyspaceName:  "ecommerce",
			hosts:         []string{"localhost:5432", "localhost:5433", "localhost:5434", "localhost:5435"},
			expected:      fourShards,
			expectedError: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observed, err := buildShards(tt.keyspaceName, tt.hosts)
			if tt.expectedError != nil {
				if !errors.As(err, &tt.expectedError) {
					t.Fatalf("expected error type %v, got %v", tt.expectedError, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected test to succeed, got error %v", err)
				}
				if len(tt.expected) != len(observed) {
					t.Fatalf("expected has %d shards, observed has %d instead", len(tt.expected), len(observed))
				}
				for i, s := range observed {
					if !reflect.DeepEqual(*tt.expected[i], *s) {
						t.Fatalf("expected shard %+v to equal %+v", s, tt.expected[i])
					}
				}
			}
		})
	}
}
