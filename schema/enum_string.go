package schema

// NewStringEnumSchema creates a new enum of string values.
func NewStringEnumSchema(validValues []string) StringEnumSchema {
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
