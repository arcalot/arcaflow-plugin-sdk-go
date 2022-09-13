package schema

// NewIntEnumSchema creates a new enum of integer values.
func NewIntEnumSchema(validValues []int, units *Units) IntEnumSchema {
	return &intEnumSchema{
		enumSchema[int]{
			validValues,
		},
		units,
	}
}

// IntEnumSchema is an enum type with integer values.
type IntEnumSchema interface {
	EnumSchema[int]
	Units() *Units
}

type intEnumSchema struct {
	enumSchema[int] `json:",inline"`
	IntUnits        *Units `json:"units"`
}

func (i intEnumSchema) TypeID() TypeID {
	return TypeIDIntEnum
}

func (i intEnumSchema) Units() *Units {
	return i.IntUnits
}
