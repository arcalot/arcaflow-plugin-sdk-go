package schema

import "fmt"

// NewStringEnumSchema creates a new enum of string values.
func NewStringEnumSchema(validValues map[string]*DisplayValue) *StringEnumSchema {
	return &StringEnumSchema{
		EnumSchema[string]{
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
	EnumSchema[string] `json:",inline"`
}

func (s StringEnumSchema) TypeID() TypeID {
	return TypeIDStringEnum
}

func (s StringEnumSchema) Unserialize(data any) (any, error) {
	typedData, err := stringInputMapper(data)
	if err != nil {
		return "", &ConstraintError{
			Message: fmt.Sprintf("'%v' (type %T) is not a valid type for a '%T' enum", data, data, typedData),
		}
	}
	return typedData, s.Validate(typedData)
}

func (s StringEnumSchema) UnserializeType(data any) (string, error) {
	unserialized, err := s.Unserialize(data)
	if err != nil {
		return "", err
	}
	return unserialized.(string), nil
}

func (s StringEnumSchema) SerializeForHuman(args map[string]any) any {
	return map[string]any{string(s.TypeID()): s.EnumSchema.SerializeForHuman(args)}
}
