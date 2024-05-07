package schema

import (
	"fmt"
	"reflect"
	"strconv"
)

// Float holds the schema information for 64-bit floating point numbers. This dataclass only has the ability to
// hold the configuration but cannot serialize, unserialize or validate. For that functionality please use
// FloatType.
type Float interface {
	TypedType[float64]

	Min() *float64
	Max() *float64
	Units() *UnitsDefinition
}

// NewFloatSchema creates a new float schema from the specified values.
func NewFloatSchema(min *float64, max *float64, units *UnitsDefinition) *FloatSchema {
	return &FloatSchema{
		MinValue:   min,
		MaxValue:   max,
		UnitsValue: units,
	}
}

type FloatSchema struct {
	ScalarType
	MinValue   *float64         `json:"min"`
	MaxValue   *float64         `json:"max"`
	UnitsValue *UnitsDefinition `json:"units"`
}

func (f FloatSchema) ReflectedType() reflect.Type {
	return reflect.TypeOf(float64(0))
}

func (f FloatSchema) TypeID() TypeID {
	return TypeIDFloat
}

func (f FloatSchema) Min() *float64 {
	return f.MinValue
}

func (f FloatSchema) Max() *float64 {
	return f.MaxValue
}

func (f FloatSchema) Units() *UnitsDefinition {
	return f.UnitsValue
}

func (f FloatSchema) Unserialize(data any) (any, error) {
	unserialized, err := floatInputMapper(data, f.UnitsValue)
	if err != nil {
		return 0, err
	}
	return unserialized, f.Validate(unserialized)
}

func (f FloatSchema) UnserializeType(data any) (float64, error) {
	unserialized, err := f.Unserialize(data)
	if err != nil {
		return 0, err
	}
	return unserialized.(float64), nil
}

func (f FloatSchema) ValidateCompatibility(typeOrData any) error {
	// Check if it's a schema.Type. If it is, verify it. If not, verify it as data.
	schemaType, ok := typeOrData.(Type)
	if !ok {
		_, err := f.Unserialize(typeOrData)
		return err
	}

	if schemaType.TypeID() != TypeIDFloat {
		return &ConstraintError{
			Message: fmt.Sprintf("unsupported data type for 'float' type: %T", schemaType),
		}
	}
	// Verify float-specific schema values
	floatSchemaType, ok := typeOrData.(*FloatSchema)
	if ok {
		// We are just verifying compatibility. So anything is accepted except for when they are mutually exclusive
		// So that's just when the min of the tested type is greater than the max of the self type,
		// or the max of the tested type is less than the min of the self type
		// For more control over this, the ValidateCompatibility API would need to change to allow subset,
		// superset, and exact verification levels.
		if (f.MinValue != nil && floatSchemaType.MaxValue != nil && (*floatSchemaType.MinValue) > (*f.MaxValue)) ||
			(f.MaxValue != nil && floatSchemaType.MinValue != nil && (*floatSchemaType.MaxValue) < (*f.MinValue)) {
			return &ConstraintError{
				Message: "mutually exclusive min/max values between float schemas",
			}
		}
	}
	return nil
}

func (f FloatSchema) Validate(d any) error {
	_, err := f.Serialize(d)
	return err
}

func (f FloatSchema) ValidateType(data float64) error {
	return f.Validate(data)
}

func (f FloatSchema) Serialize(d any) (any, error) {
	data, err := asFloat(d)
	if err != nil {
		return data, err
	}
	if f.MinValue != nil && data < *f.MinValue {
		return data, &ConstraintError{
			Message: fmt.Sprintf("Must be at least %f", *f.MinValue),
		}
	}
	if f.MaxValue != nil && data > *f.MaxValue {
		return data, &ConstraintError{
			Message: fmt.Sprintf("Must be at most %f", *f.MaxValue),
		}
	}
	return data, nil
}

func asFloat(d any) (float64, error) {
	data, ok := d.(float64)
	if !ok {
		var i float64
		intType := reflect.TypeOf(i)
		dValue := reflect.ValueOf(d)
		if !dValue.CanConvert(intType) {
			return 0, &ConstraintError{
				Message: fmt.Sprintf("%T is not a valid data type for a float schema.", d),
			}
		}
		data = dValue.Convert(intType).Float()
	}
	return data, nil
}

func (f FloatSchema) SerializeType(data float64) (any, error) {
	return f.Serialize(data)
}

func floatInputMapper(data any, u *UnitsDefinition) (float64, error) {
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
			return float64(1), nil
		}
		return float64(0), nil
	default:
		return float64(0), fmt.Errorf("%T cannot be converted to a float64", data)
	}
}
