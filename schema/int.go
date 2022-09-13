package schema

// IntSchema holds the schema information for 64-bit integers. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use IntType.
type IntSchema interface {
	AbstractSchema
	Min() *int64
	Max() *int64
	Units() *Units
}

// NewIntSchema creates a new integer schema with the specified values.
func NewIntSchema(min *int64, max *int64, units *Units) IntSchema {
	return &intSchema{
		min,
		max,
		units,
	}
}

type intSchema struct {
	MinValue   *int64 `json:"min"`
	MaxValue   *int64 `json:"max"`
	UnitsValue *Units `json:"units"`
}

func (i intSchema) TypeID() TypeID {
	return TypeIDInt
}

func (i intSchema) Min() *int64 {
	return i.MinValue
}

func (i intSchema) Max() *int64 {
	return i.MaxValue
}

func (i intSchema) Units() *Units {
	return i.UnitsValue
}
