package schema

import "reflect"

// OneOfInt holds the definition of variable types with an integer discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
type OneOfInt interface {
	OneOf[int64]
}

// NewOneOfIntSchema creates a new OneOf-type with integer discriminators.
func NewOneOfIntSchema[ItemsInterface any](
	types map[int64]Object,
	discriminatorFieldName string,
	discriminatorInlined bool,
) *OneOfSchema[int64] {
	var defaultValue ItemsInterface
	oos := &OneOfSchema[int64]{
		reflect.TypeOf(&defaultValue).Elem(),
		types,
		discriminatorFieldName,
		discriminatorInlined,
	}
	//err := oos.ValidateSubtypeDiscriminatorInlineFields()
	//if err != nil {
	//	panic(err)
	//}
	return oos
}
