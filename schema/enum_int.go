package schema

import "fmt"

// NewIntEnumSchema creates a new enum of integer values.
func NewIntEnumSchema(validValues map[int64]string, units *Units) IntEnumSchema {
	return &intEnumSchema{
		enumSchema[int64]{
			validValues,
		},
		units,
	}
}

// IntEnumSchema is an enum type with integer values.
type IntEnumSchema interface {
	EnumSchema[int64]
	Units() *Units
}

type intEnumSchema struct {
	enumSchema[int64] `json:",inline"`
	IntUnits          *Units `json:"units"`
}

func (i intEnumSchema) TypeID() TypeID {
	return TypeIDIntEnum
}

func (i intEnumSchema) Units() *Units {
	return i.IntUnits
}

// IntEnumType represents an enum type that is an integer.
type IntEnumType interface {
	EnumType[int64]
	IntEnumSchema
}

// NewIntEnumType defines a new enum that holds integer types.
func NewIntEnumType(validValues map[int64]string, units *Units) IntEnumType {
	return &intEnumType{
		enumType[int64, IntEnumSchema]{
			schemaType: NewIntEnumSchema(validValues, units),
		},
	}
}

type intEnumType struct {
	enumType[int64, IntEnumSchema] `json:",inline"`
}

func (i intEnumType) UnderlyingType() int64 {
	return int64(0)
}

func (i intEnumType) Units() *Units {
	return i.schemaType.Units()
}

func (i intEnumType) Unserialize(data any) (int64, error) {
	typedData, err := intInputMapper(data, i.Units())
	if err != nil {
		return 0, &ConstraintError{
			Message: fmt.Sprintf("'%v' (type %T) is not a valid type for a '%T' enum", data, data, typedData),
		}
	}
	return typedData, i.Validate(typedData)
}
