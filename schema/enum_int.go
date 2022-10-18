package schema

import "fmt"

// NewIntEnumSchema creates a new enum of integer values.
func NewIntEnumSchema(validValues map[int64]*DisplayValue, units *UnitsDefinition) *IntEnumSchema {
	return &IntEnumSchema{
		EnumSchema[int64]{
			validValues,
		},
		units,
	}
}

// IntEnum is an enum type with integer values.
type IntEnum interface {
	Enum[int64]
	Units() *UnitsDefinition
}

// IntEnumSchema is an enum type with integer values.
type IntEnumSchema struct {
	EnumSchema[int64] `json:",inline"`
	IntUnits          *UnitsDefinition `json:"UnitsDefinition"`
}

func (i IntEnumSchema) TypeID() TypeID {
	return TypeIDIntEnum
}

func (i IntEnumSchema) Units() *UnitsDefinition {
	return i.IntUnits
}

func (i IntEnumSchema) Unserialize(data any) (any, error) {
	typedData, err := intInputMapper(data, i.Units())
	if err != nil {
		return 0, &ConstraintError{
			Message: fmt.Sprintf("'%v' (type %T) is not a valid type for a '%T' enum", data, data, typedData),
		}
	}
	return typedData, i.Validate(typedData)
}

func (e IntEnumSchema) UnserializeType(data any) (int64, error) {
	unserialized, err := e.Unserialize(data)
	if err != nil {
		return 0, err
	}
	return unserialized.(int64), nil
}
