package schema

import "fmt"

// NewStringEnumSchema creates a new enum of string values.
func NewStringEnumSchema(validValues map[string]string) StringEnumSchema {
	return &stringEnumSchema{
		enumSchema[string]{
			validValues,
		},
	}
}

// StringEnumSchema is an enum type with string values.
type StringEnumSchema interface {
	EnumSchema[string]
}

type stringEnumSchema struct {
	enumSchema[string] `json:",inline"`
}

func (s stringEnumSchema) TypeID() TypeID {
	return TypeIDStringEnum
}

// StringEnumType represents an enum type that is a string.
type StringEnumType interface {
	EnumType[string]
	StringEnumSchema
}

// NewStringEnumType creates an enum type that holds strings.
func NewStringEnumType(validValues map[string]string) StringEnumType {
	return &stringEnumType{
		enumType[string, StringEnumSchema]{
			schemaType: NewStringEnumSchema(validValues),
		},
	}
}

type stringEnumType struct {
	enumType[string, StringEnumSchema] `json:",inline"`
}

func (s stringEnumType) Unserialize(data any) (string, error) {
	typedData, err := stringInputMapper(data)
	if err != nil {
		return "", &ConstraintError{
			Message: fmt.Sprintf("'%v' (type %T) is not a valid type for a '%T' enum", data, data, typedData),
		}
	}
	return typedData, s.Validate(typedData)
}
