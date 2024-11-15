package schema

import "fmt"

// NewStringEnumSchema creates a new enum of string values.
func NewStringEnumSchema(validValues map[string]*DisplayValue) *StringEnumSchema {
	return &StringEnumSchema{
		TypedStringEnumSchema[string]{
			EnumSchema[string, string]{
				ValidValuesMap: validValues,
			},
		},
	}
}

// NewTypedStringEnumSchema allows the use of a type with string as an underlying type.
// Useful for external APIs that are being mapped to a schema that use string enums.
func NewTypedStringEnumSchema[T ~string](validValues map[T]*DisplayValue) *TypedStringEnumSchema[T] {
	return &TypedStringEnumSchema[T]{
		EnumSchema[string, T]{
			ValidValuesMap: validValues,
		},
	}
}

// StringEnum is an enum type with string values.
type StringEnum interface {
	Enum[string]
}

// StringEnumSchema is an enum type with string values.
type StringEnumSchema struct {
	TypedStringEnumSchema[string] `json:",inline"`
}

// TypedStringEnumSchema is an enum type with string values, but with a generic
// element for golang enums that have an underlying string type.
type TypedStringEnumSchema[T ~string] struct {
	EnumSchema[string, T] `json:",inline"`
}

func (s TypedStringEnumSchema[T]) TypeID() TypeID {
	return TypeIDStringEnum
}

func (s TypedStringEnumSchema[T]) Unserialize(data any) (any, error) {
	strData, err := stringInputMapper(data)
	typedData := T(strData)
	if err != nil {
		return "", &ConstraintError{
			Message: fmt.Sprintf("'%v' (type %T) is not a valid type for a '%T' enum", data, data, typedData),
		}
	}
	return typedData, s.Validate(typedData)
}

func (s TypedStringEnumSchema[T]) UnserializeType(data any) (string, error) {
	unserialized, err := s.Unserialize(data)
	if err != nil {
		return "", err
	}
	return unserialized.(string), nil
}
