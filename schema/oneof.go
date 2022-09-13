package schema

// OneOfSchema is the root interface for one-of types. It should not be used directly but is provided for convenience.
type OneOfSchema[T int64 | string] interface {
	AbstractSchema
	Types() map[T]RefSchema
	DiscriminatorFieldName() string
}

type oneOfSchema[T int64 | string] struct {
	TypesValue                  map[T]RefSchema `json:"types"`
	DiscriminatorFieldNameValue string          `json:"discriminator_field_name"`
}

func (o oneOfSchema[T]) Types() map[T]RefSchema {
	return o.TypesValue
}

func (o oneOfSchema[T]) DiscriminatorFieldName() string {
	return o.DiscriminatorFieldNameValue
}
