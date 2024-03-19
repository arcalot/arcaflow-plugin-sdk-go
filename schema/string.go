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
func NewStringSchema(minLen *int64, maxLen *int64, pattern *regexp.Regexp) *StringSchema {
	return &StringSchema{
		minLen,
		maxLen,
		pattern,
	}
}

type StringSchema struct {
	MinValue     *int64         `json:"min"`
	MaxValue     *int64         `json:"max"`
	PatternValue *regexp.Regexp `json:"pattern"`
}

func (s StringSchema) TypeID() TypeID {
	return TypeIDString
}

func (s StringSchema) ReflectedType() reflect.Type {
	return reflect.TypeOf("")
}

// Min returns the min length of the string.
func (s StringSchema) Min() *int64 {
	return s.MinValue
}

// Max returns the max length of the string.
func (s StringSchema) Max() *int64 {
	return s.MaxValue
}

func (s StringSchema) Pattern() *regexp.Regexp {
	return s.PatternValue
}

func (s StringSchema) ApplyScope(scope Scope, namespace string) {
}

func (s StringSchema) ValidateReferences() error {
	// No references in this type. No work to do.
	return nil
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

func (s StringSchema) ValidateCompatibility(typeOrData any) error {
	// Check if it's a schema.Type. If it is, verify it. If not, verify it as data.
	schemaType, ok := typeOrData.(Type)
	if !ok {
		// Just verify strings, since everything is string from raw values.
		stringType, ok := typeOrData.(string)
		if !ok {
			return &ConstraintError{
				Message: fmt.Sprintf("unsupported data type for 'string' type: %T", typeOrData),
			}
		} else {
			_, err := s.Unserialize(stringType)
			return err
		}
	}

	if schemaType.TypeID() == TypeIDStringEnum {
		// For now, just accept the enums. Consider more validations later.
		return nil
	} else if schemaType.TypeID() != TypeIDString {
		return &ConstraintError{
			Message: fmt.Sprintf("unsupported data type for 'string' type: %T", schemaType),
		}
	}
	// Verify string-specific schema values
	stringSchemaType, ok := typeOrData.(*StringSchema)
	if ok {
		// We are just verifying compatibility. So anything is accepted except for when they are mutually exclusive
		// So that's just when the min of the tested type is greater than the max of the self type,
		// or the max of the tested type is less than the min of the self type
		// For more control over this, the ValidateCompatibility API would need to change to allow subset,
		// superset, and exact verification levels.
		if (s.MinValue != nil && stringSchemaType.MaxValue != nil && (*stringSchemaType.MinValue) > (*s.MaxValue)) ||
			(s.MaxValue != nil && stringSchemaType.MinValue != nil && (*stringSchemaType.MaxValue) < (*s.MinValue)) {
			return &ConstraintError{
				Message: "mutually exclusive string lengths between string schemas",
			}
		}
		// Is it possible to validate the patterns in some way that makes sense?
	}
	return nil
}

func (s StringSchema) Validate(d any) error {
	_, err := s.Serialize(d)
	return err
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
			Message: fmt.Sprintf("String '%s' must match the pattern '%s'", data, (*s.PatternValue).String()),
		}
	}
	return nil
}

func (s StringSchema) Serialize(d any) (any, error) {
	data, err := asString(d)
	if err != nil {
		return data, err
	}
	if s.MinValue != nil && int64(len(data)) < *s.MinValue {
		return data, &ConstraintError{
			Message: fmt.Sprintf("String must be at least %d characters, %d given", *s.MinValue, int64(len(data))),
		}
	}
	if s.MaxValue != nil && int64(len(data)) > *s.MaxValue {
		return data, &ConstraintError{
			Message: fmt.Sprintf("String must be at most %d characters, %d given", *s.MaxValue, int64(len(data))),
		}
	}
	if s.PatternValue != nil && !(*s.PatternValue).MatchString(data) {
		return data, &ConstraintError{
			Message: fmt.Sprintf("String '%s' must match the pattern '%s'", data, (*s.PatternValue).String()),
		}
	}
	return data, nil
}

func asString(d any) (string, error) {
	data, ok := d.(string)
	if !ok {
		var i string
		stringType := reflect.TypeOf(i)
		dValue := reflect.ValueOf(d)
		if !dValue.CanConvert(stringType) {
			return "", &ConstraintError{
				Message: fmt.Sprintf("%T is not a valid data type for a string schema.", d),
			}
		}
		data = dValue.Convert(stringType).String()
	}
	return data, nil
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
