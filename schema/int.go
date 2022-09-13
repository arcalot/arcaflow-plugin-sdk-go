package schema

import (
	"fmt"
	"math"
	"strconv"
)

// IntSchema holds the schema information for 64-bit integers. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use IntType.
type IntSchema interface {
	AbstractSchema
	Min() *int64
	Max() *int64
	Units() *Units
}

// NewIntSchema creates a new integer schema with the specified values.
func NewIntSchema(min *int64, max *int64, units *Units) IntSchema {
	return &intSchema{
		min,
		max,
		units,
	}
}

type intSchema struct {
	MinValue   *int64 `json:"min"`
	MaxValue   *int64 `json:"max"`
	UnitsValue *Units `json:"units"`
}

func (i intSchema) TypeID() TypeID {
	return TypeIDInt
}

func (i intSchema) Min() *int64 {
	return i.MinValue
}

func (i intSchema) Max() *int64 {
	return i.MaxValue
}

func (i intSchema) Units() *Units {
	return i.UnitsValue
}

// IntType is the version of the IntSchema that supports serialization.
type IntType interface {
	AbstractType[int64]
	IntSchema
}

// NewIntType defines a serializable integer representation.
func NewIntType(min *int64, max *int64, units *Units) IntType {
	return &intType{
		intSchema{
			min,
			max,
			units,
		},
	}
}

type intType struct {
	intSchema `json:",inline"`
}

func (i intType) TypeID() TypeID {
	return TypeIDInt
}

func (i intType) Unserialize(data any) (int64, error) {
	unserialized, err := intInputMapper(data, i.UnitsValue)
	if err != nil {
		return 0, err
	}
	return unserialized, i.Validate(unserialized)
}

func (i intType) Validate(data int64) error {
	if i.MinValue != nil && data < *i.MinValue {
		return ConstraintError{
			Message: fmt.Sprintf("Must be at least %d", *i.MinValue),
		}
	}
	if i.MaxValue != nil && data > *i.MaxValue {
		return ConstraintError{
			Message: fmt.Sprintf("Must be at most %d", *i.MaxValue),
		}
	}
	return nil
}

func (i intType) Serialize(data int64) (any, error) {
	return data, i.Validate(data)
}

func intInputMapper(data any, u *Units) (int64, error) {
	switch v := data.(type) {
	case string:
		if u != nil {
			return (*u).ParseInt(v)
		}
		return strconv.ParseInt(v, 10, 64)
	case int64:
		return v, nil
	case uint64:
		if v > math.MaxInt64 {
			return 0, fmt.Errorf("number is too large for an int64: %d", v)
		}
		return int64(v), nil
	case int:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%T cannot be converted to an int64", data)
	}
}
