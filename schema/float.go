package schema

import (
	"fmt"
	"strconv"
)

// FloatSchema holds the schema information for 64-bit floating point numbers. This dataclass only has the ability to
// hold the configuration but cannot serialize, unserialize or validate. For that functionality please use
// FloatType.
type FloatSchema interface {
	AbstractSchema

	Min() *float64
	Max() *float64
	Units() *Units
}

// NewFloatSchema creates a new float schema from the specified values.
func NewFloatSchema(min *float64, max *float64, units *Units) FloatSchema {
	return &floatSchema{
		min,
		max,
		units,
	}
}

type floatSchema struct {
	MinValue   *float64 `json:"min"`
	MaxValue   *float64 `json:"max"`
	UnitsValue *Units   `json:"units"`
}

func (f floatSchema) TypeID() TypeID {
	return TypeIDFloat
}

func (f floatSchema) Min() *float64 {
	return f.MinValue
}

func (f floatSchema) Max() *float64 {
	return f.MaxValue
}

func (f floatSchema) Units() *Units {
	return f.UnitsValue
}

// FloatType is the version of the FloatSchema that supports serialization.
type FloatType interface {
	AbstractType[float64]
	FloatSchema
}

// NewFloatType defines a serializable integer representation.
func NewFloatType(min *float64, max *float64, units *Units) FloatType {
	return &floatType{
		floatSchema{
			min,
			max,
			units,
		},
	}
}

type floatType struct {
	floatSchema `json:",inline"`
}

func (f floatType) TypeID() TypeID {
	return TypeIDFloat
}

func (f floatType) Unserialize(data any) (float64, error) {
	unserialized, err := floatInputMapper(data, f.UnitsValue)
	if err != nil {
		return 0, err
	}
	return unserialized, f.Validate(unserialized)
}

func (f floatType) Validate(data float64) error {
	if f.MinValue != nil && data < *f.MinValue {
		return &ConstraintError{
			Message: fmt.Sprintf("Must be at least %f", *f.MinValue),
		}
	}
	if f.MaxValue != nil && data > *f.MaxValue {
		return &ConstraintError{
			Message: fmt.Sprintf("Must be at most %f", *f.MaxValue),
		}
	}
	return nil
}

func (f floatType) Serialize(data float64) (any, error) {
	return data, f.Validate(data)
}

func floatInputMapper(data any, u *Units) (float64, error) {
	switch v := data.(type) {
	case string:
		if u != nil {
			return (*u).ParseFloat(v)
		}
		return strconv.ParseFloat(v, 64)
	case int64:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case int:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%T cannot be converted to a float64", data)
	}
}
