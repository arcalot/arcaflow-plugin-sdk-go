package schema

// OneOfString holds the definition of variable types with an integer discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
type OneOfString[ItemsInterface any] interface {
	OneOf[string, ItemsInterface]
}

// NewOneOfStringSchema creates a new OneOf-type with integer discriminators.
func NewOneOfStringSchema[ItemsInterface any](
	types map[string]Object,
	discriminatorFieldName string,
) *OneOfSchema[string, ItemsInterface] {
	return &OneOfSchema[string, ItemsInterface]{
		types,
		discriminatorFieldName,
	}
}
