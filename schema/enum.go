package schema

type enumValue interface {
	int | string
}

// EnumSchema is an abstract schema for enumerated types.
type EnumSchema[T enumValue] interface {
	AbstractSchema
	ValidValues() []T
}

type enumSchema[T enumValue] struct {
	ValidValuesList []T `json:"valid_values"`
}

func (e enumSchema[T]) ValidValues() []T {
	return e.ValidValuesList
}
