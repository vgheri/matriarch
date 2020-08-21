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

// VIndexType models the type of vindex in Matriarch.
type VIndexType string

const (
	// Primary means the index is used to physically distribute data amongst shards.
	Primary VIndexType = "primary"
	// Secondary means the index is only used to make queries faster.
	Secondary VIndexType = "secondary"
)

// IsValid checks the VTableType assigned value is valid.
func (v VIndexType) IsValid() error {
	switch v {
	case Primary, Secondary:
		return nil
	}
	return errors.New("Inalid VIndexType")
}

// VIndex models a VIndex section in the vschema file
type VIndex struct {
	// List of columns part of the vschema.
	Columns []string
	// Type of the  vindex. Can be either "primary" or "secondary".
	// Only one of the two must be set.
	Type VIndexType
}

// Table models a Table section in the vschema file.
type Table struct {
	// The name of the table.
	Name string
	// The type of the table. Can be either "sharded" or "reference".
	// Only one of the two must be set.
	Type VTableType

	VIndexes []VIndex
}

// Vschema models the content of the vschema file, with all its options.
type Vschema struct {
	// The name of the keyspace. It will be used to name the databases created in each shard.
	Keyspace string `json:"keyspace"`
	// The list of tables part of the vschema
	Tables []Table `json:"tables"`
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

func (v *Vschema) GetTable(name string) *Table {
	for _, t := range v.Tables {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func (t *Table) GetPrimaryVIndex() *VIndex {
	var index VIndex
	for _, i := range t.VIndexes {
		if i.Type == Primary {
			index = i
		}
	}
	return &index
}
