package schema

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// Int holds the schema information for 64-bit integers. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use IntType.
type Int interface {
	TypedType[int64]
	Min() *int64
	Max() *int64
	Units() *UnitsDefinition
}

// NewIntSchema creates a new integer schema with the specified values.
func NewIntSchema(min *int64, max *int64, units *UnitsDefinition) *IntSchema {
	return &IntSchema{
		min,
		max,
		units,
	}
}

type IntSchema struct {
	MinValue   *int64           `json:"min"`
	MaxValue   *int64           `json:"max"`
	UnitsValue *UnitsDefinition `json:"units"`
}

func (i IntSchema) ReflectedType() reflect.Type {
	return reflect.TypeOf(int64(0))
}

func (i IntSchema) ApplyScope(scope Scope, namespace string) {
}

func (i IntSchema) ValidateReferences() error {
	// Not applicable
	return nil
}

func (i IntSchema) TypeID() TypeID {
	return TypeIDInt
}

func (i IntSchema) Min() *int64 {
	return i.MinValue
}

func (i IntSchema) Max() *int64 {
	return i.MaxValue
}

func (i IntSchema) Units() *UnitsDefinition {
	return i.UnitsValue
}

func (i IntSchema) Unserialize(data any) (any, error) {
	unserialized, err := intInputMapper(data, i.UnitsValue)
	if err != nil {
		return 0, err
	}
	return unserialized, i.Validate(unserialized)
}

func (i IntSchema) Serialize(d any) (any, error) {
	data, err := asInt(d)
	if err != nil {
		return data, err
	}
	if i.MinValue != nil && data < *i.MinValue {
		return data, &ConstraintError{
			Message: fmt.Sprintf("Must be at least %d", *i.MinValue),
		}
	}
	if i.MaxValue != nil && data > *i.MaxValue {
		return data, &ConstraintError{
			Message: fmt.Sprintf("Must be at most %d", *i.MaxValue),
		}
	}
	return data, nil
}

func asInt(d any) (int64, error) {
	data, ok := d.(int64)
	if !ok {
		var i int64
		intType := reflect.TypeOf(i)
		dValue := reflect.ValueOf(d)
		if !dValue.CanConvert(intType) {
			return 0, &ConstraintError{
				Message: fmt.Sprintf("%T is not a valid data type for an int schema.", d),
			}
		}
		data = dValue.Convert(intType).Int()
	}
	return data, nil
}

func (i IntSchema) ValidateCompatibility(typeOrData any) error {
	// Check if it's a schema.Type. If it is, verify it. If not, verify it as data.
	schemaType, ok := typeOrData.(Type)
	if !ok {
		_, err := i.Unserialize(typeOrData)
		return err
	}

	if schemaType.TypeID() == TypeIDIntEnum {
		// Just accept the enums. It's possible to do more
		return nil
	}
	if schemaType.TypeID() != TypeIDInt {
		return &ConstraintError{
			Message: fmt.Sprintf("unsupported data type for 'int' type: %T", schemaType),
		}
	}
	// Verify int-specific schema values
	intSchemaType, ok := typeOrData.(*IntSchema)
	if ok {
		// We are just verifying compatibility. So anything is accepted except for when they are mutually exclusive
		// So that's just when the min of the tested type is greater than the max of the self type,
		// or the max of the tested type is less than the min of the self type
		// For more control over this, the ValidateCompatibility API would need to change to allow subset,
		// superset, and exact verification levels.
		if (i.MinValue != nil && intSchemaType.MaxValue != nil && (*intSchemaType.MinValue) > (*i.MaxValue)) ||
			(i.MaxValue != nil && intSchemaType.MinValue != nil && (*intSchemaType.MaxValue) < (*i.MinValue)) {
			return &ConstraintError{
				Message: "mutually exclusive min/max values between int schemas",
			}
		}
		// Should units be validated?
	}
	return nil
}

func (i IntSchema) Validate(d any) error {
	_, err := i.Serialize(d)
	return err
}

func (i IntSchema) UnserializeType(data any) (int64, error) {
	unserialized, err := i.Unserialize(data)
	if err != nil {
		return 0, err
	}
	return unserialized.(int64), nil
}

func (i IntSchema) ValidateType(data int64) error {
	return i.Validate(data)
}

func (i IntSchema) SerializeType(data int64) (any, error) {
	return i.Serialize(data)
}

//nolint:funlen
func intInputMapper(data any, u *UnitsDefinition) (int64, error) {
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
	case float64:
		i := int64(v)
		if v != float64(i) {
			return 0, fmt.Errorf("float64 number %f cannot be converted to an int64", v)
		}
		return i, nil
	case float32:
		i := int64(v)
		if v != float32(i) {
			return 0, fmt.Errorf("float32 number %f cannot be converted to an int64", v)
		}
		return i, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("%T cannot be converted to an int64", data)
	}
}
