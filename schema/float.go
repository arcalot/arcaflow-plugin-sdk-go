package schema

// FloatSchema holds the schema information for 64-bit floating point numbers. This dataclass only has the ability to
// hold the configuration but cannot serialize, unserialize or validate. For that functionality please use
// FloatType.
type FloatSchema interface {
	AbstractSchema

	Min() *float64
	Max() *float64
	Units() *Units
}

// NewFloatSchema creates a new float schema from the specified values.
func NewFloatSchema(min *float64, max *float64, units *Units) FloatSchema {
	return &floatSchema{
		min,
		max,
		units,
	}
}

type floatSchema struct {
	MinValue   *float64 `json:"min"`
	MaxValue   *float64 `json:"max"`
	UnitsValue *Units   `json:"units"`
}

func (f floatSchema) TypeID() TypeID {
	return TypeIDFloat
}

func (f floatSchema) Min() *float64 {
	return f.MinValue
}

func (f floatSchema) Max() *float64 {
	return f.MaxValue
}

func (f floatSchema) Units() *Units {
	return f.UnitsValue
}
