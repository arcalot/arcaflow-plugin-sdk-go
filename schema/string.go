package schema

import (
	"fmt"
	"regexp"
)

// StringSchema holds schema information for strings. This dataclass only has the ability to hold the configuration but
// cannot serialize, unserialize or validate. For that functionality please use StringType.
type StringSchema interface {
	AbstractSchema
	Min() *int64
	Max() *int64
	Pattern() *regexp.Regexp
}

// NewStringSchema creates a new string schema.
func NewStringSchema(min *int64, max *int64, pattern *regexp.Regexp) StringSchema {
	return &stringSchema{
		min,
		max,
		pattern,
	}
}

type stringSchema struct {
	MinValue     *int64         `json:"min"`
	MaxValue     *int64         `json:"max"`
	PatternValue *regexp.Regexp `json:"pattern"`
}

func (s stringSchema) TypeID() TypeID {
	return TypeIDString
}

func (s stringSchema) Min() *int64 {
	return s.MinValue
}

func (s stringSchema) Max() *int64 {
	return s.MaxValue
}

func (s stringSchema) Pattern() *regexp.Regexp {
	return s.PatternValue
}

// StringType is the serializable variant of StringSchema.
type StringType interface {
	AbstractType[string]
	StringSchema
}

// NewStringType creates a new string type definition with the given constraints.
func NewStringType(min *int64, max *int64, pattern *regexp.Regexp) StringType {
	return &stringType{
		stringSchema{min, max, pattern},
	}
}

type stringType struct {
	stringSchema `json:",inline"`
}

func (s stringType) Unserialize(data any) (string, error) {
	unserialized, err := stringInputMapper(data)
	if err != nil {
		return "", err
	}
	return unserialized, s.Validate(unserialized)
}

func (s stringType) Validate(data string) error {
	if s.MinValue != nil && int64(len(data)) < *s.MinValue {
		return ConstraintError{
			Message: fmt.Sprintf("String must be at least %d characters, %d given", *s.MinValue, int64(len(data))),
		}
	}
	if s.MaxValue != nil && int64(len(data)) > *s.MaxValue {
		return ConstraintError{
			Message: fmt.Sprintf("String must be at most %d characters, %d given", *s.MaxValue, int64(len(data))),
		}
	}
	if s.PatternValue != nil && !s.PatternValue.MatchString(data) {
		return ConstraintError{
			Message: fmt.Sprintf("String must match the pattern %s", s.PatternValue.String()),
		}
	}
	return nil
}

func (s stringType) Serialize(data string) (any, error) {
	return data, s.Validate(data)
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
