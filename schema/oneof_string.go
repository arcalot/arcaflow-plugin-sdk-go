package schema

import "reflect"

// OneOfString holds the definition of variable types with an integer discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
type OneOfString interface {
	OneOf[string]
}

// NewOneOfStringSchema creates a new OneOf-type with integer discriminators.
func NewOneOfStringSchema[ItemsInterface any](
	types map[string]Object,
	discriminatorFieldName string,
	discriminatorInlined bool,
) *OneOfSchema[string] {
	var defaultValue ItemsInterface
	return &OneOfSchema[string]{
		reflect.TypeOf(&defaultValue).Elem(),
		types,
		discriminatorFieldName,
		discriminatorInlined,
	}
}
