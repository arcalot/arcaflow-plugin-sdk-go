package schema

import (
	"encoding/json"
	"fmt"
	"strings"
)

// BoolSchema holds the schema information for boolean types. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use BoolType.
type BoolSchema interface {
	AbstractSchema
}

// NewBoolSchema creates a new boolean representation.
func NewBoolSchema() BoolSchema {
	return &boolSchema{}
}

type boolSchema struct {
}

func (b boolSchema) TypeID() TypeID {
	return TypeIDBool
}

// BoolType contains the unserialization and serialization routines for boolean values.
type BoolType interface {
	BoolSchema
	AbstractType[bool]
}

// NewBoolType creates a new boolean representation.
func NewBoolType() BoolType {
	return &boolType{
		NewBoolSchema(),
	}
}

type boolType struct {
	schema BoolSchema
}

func (b *boolType) UnderlyingType() bool {
	return false
}

func (b *boolType) ApplyScope(_ ScopeSchema[PropertyType, ObjectType[any]]) {
}

func (b *boolType) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.schema)
}

func (b *boolType) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, b.schema)
}

func (b *boolType) TypeID() TypeID {
	return b.schema.TypeID()
}

var boolStringValues = map[string]bool{
	"1":        true,
	"yes":      true,
	"y":        true,
	"on":       true,
	"true":     true,
	"enable":   true,
	"enabled":  true,
	"0":        false,
	"no":       false,
	"n":        false,
	"off":      false,
	"false":    false,
	"disable":  false,
	"disabled": false,
}

func (b *boolType) Unserialize(data any) (bool, error) {
	intConverter := func(data int64) (bool, error) {
		switch data {
		case 1:
			return true, nil
		case 0:
			return false, nil
		default:
			return false, fmt.Errorf("'%d' is not a valid boolean value", data)
		}
	}
	switch v := data.(type) {
	case bool:
		return v, nil
	case string:
		lowerStr := strings.ToLower(v)
		if serializedValue, ok := boolStringValues[lowerStr]; ok {
			return serializedValue, nil
		}
	case int:
		return intConverter(int64(v))
	case uint:
		return intConverter(int64(v))
	case int64:
		return intConverter(v)
	case uint64:
		return intConverter(int64(v))
	case int32:
		return intConverter(int64(v))
	case uint32:
		return intConverter(int64(v))
	case int16:
		return intConverter(int64(v))
	case uint16:
		return intConverter(int64(v))
	case int8:
		return intConverter(int64(v))
	case uint8:
		return intConverter(int64(v))
	}
	return false, fmt.Errorf("'%v' is not a valid boolean value", data)
}

func (b *boolType) Validate(_ bool) error {
	return nil
}

func (b *boolType) Serialize(data bool) (any, error) {
	return data, nil
}
