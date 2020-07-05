package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// VTableType models the type of tables in Matriarch.
type VTableType string

const (
	// Sharded means the content of the table is scattered amongst each shard.
	Sharded VTableType = "sharded"
	// Reference means the content of the table is duplicated on each shard.
	Reference VTableType = "reference"
)

// IsValid checks the VTableType assigned value is valid.
func (v VTableType) IsValid() error {
	switch v {
	case Sharded, Reference:
		return nil
	}
	return errors.New("Inalid VTableType")
}

// Table models a Table section in the vschema file.
type Table struct {
	// The name of the table.
	Name string
	// The type of the table. Can be either "sharded" or "reference".
	// Only one of the two must be set.
	Type VTableType
}

// Vschema models the content of the vschema file, with all its options.
type Vschema struct {
	// The name of the keyspace. It will be used to name the databases created in each shard.
	Keyspace string `json:"keyspace"`
}

func readVschemaFile(path string) (*Vschema, error) {
	var vschema Vschema
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open %v: %w", path, err)
	}
	jsonParser := json.NewDecoder(file)
	if err = jsonParser.Decode(&vschema); err != nil {
		return nil, fmt.Errorf("cannot decode content of vschema file: %w", err)
	}

	// TODO
	// Use https://pkg.go.dev/gopkg.in/go-playground/validator.v9?tab=doc to validate struct content
	return &vschema, nil
}
