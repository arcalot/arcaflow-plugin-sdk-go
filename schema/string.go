package schema

import (
	"fmt"
	"reflect"
	"regexp"
)

// String holds schema information for strings. This dataclass only has the ability to hold the configuration but
// cannot serialize, unserialize or validate. For that functionality please use StringType.
type String interface {
	TypedType[string]

	Min() *int64
	Max() *int64
	Pattern() *regexp.Regexp
}

// NewStringSchema creates a new string schema.
func NewStringSchema(min *int64, max *int64, pattern *regexp.Regexp) *StringSchema {
	return &StringSchema{
		min,
		max,
		pattern,
	}
}

type StringSchema struct {
	MinValue     *int64         `json:"min,omitempty"`
	MaxValue     *int64         `json:"max,omitempty"`
	PatternValue *regexp.Regexp `json:"pattern,omitempty"`
}

func (s StringSchema) TypeID() TypeID {
	return TypeIDString
}

func (s StringSchema) ReflectedType() reflect.Type {
	return reflect.TypeOf("")
}

func (s StringSchema) Min() *int64 {
	return s.MinValue
}

func (s StringSchema) Max() *int64 {
	return s.MaxValue
}

func (s StringSchema) Pattern() *regexp.Regexp {
	return s.PatternValue
}

func (s StringSchema) ApplyScope(scope Scope) {
}

func (s StringSchema) Unserialize(data any) (any, error) {
	return s.UnserializeType(data)
}

func (s StringSchema) UnserializeType(data any) (string, error) {
	unserialized, err := stringInputMapper(data)
	if err != nil {
		return "", err
	}
	return unserialized, s.ValidateType(unserialized)
}

func (s StringSchema) Validate(data any) error {
	d, ok := data.(string)
	if !ok {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for a string schema.", data),
		}
	}
	return s.ValidateType(d)
}

func (s StringSchema) ValidateType(data string) error {
	if s.MinValue != nil && int64(len(data)) < *s.MinValue {
		return &ConstraintError{
			Message: fmt.Sprintf("String must be at least %d characters, %d given", *s.MinValue, int64(len(data))),
		}
	}
	if s.MaxValue != nil && int64(len(data)) > *s.MaxValue {
		return &ConstraintError{
			Message: fmt.Sprintf("String must be at most %d characters, %d given", *s.MaxValue, int64(len(data))),
		}
	}
	if s.PatternValue != nil && !(*s.PatternValue).MatchString(data) {
		return &ConstraintError{
			Message: fmt.Sprintf("String must match the pattern %s", (*s.PatternValue).String()),
		}
	}
	return nil
}

func (s StringSchema) Serialize(data any) (any, error) {
	d, ok := data.(string)
	if !ok {
		return "", &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for a string schema.", d),
		}
	}
	return s.SerializeType(d)
}

func (s StringSchema) SerializeType(data string) (any, error) {
	return data, s.ValidateType(data)
}

func stringInputMapper(data any) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case uint:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case uint64:
		return fmt.Sprintf("%d", v), nil
	case int32:
		return fmt.Sprintf("%d", v), nil
	case uint32:
		return fmt.Sprintf("%d", v), nil
	case int16:
		return fmt.Sprintf("%d", v), nil
	case uint16:
		return fmt.Sprintf("%d", v), nil
	case int8:
		return fmt.Sprintf("%d", v), nil
	case uint8:
		return fmt.Sprintf("%d", v), nil
	case float64:
		return fmt.Sprintf("%f", v), nil
	case float32:
		return fmt.Sprintf("%f", v), nil
	default:
		return "", fmt.Errorf("%T cannot be converted to a string", data)
	}
}
