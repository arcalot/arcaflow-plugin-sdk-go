package schema

import (
	"fmt"
	"reflect"
)

// OneOfIntSchema holds the definition of variable types with an integer discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use OneOfIntType.
type OneOfIntSchema[RefSchemaType RefSchema] interface {
	OneOfSchema[int64, RefSchemaType]
}

// NewOneOfIntSchema creates a new OneOf-type with integer discriminators.
func NewOneOfIntSchema(
	types map[int64]RefSchema,
	discriminatorFieldName string,
) OneOfIntSchema[RefSchema] {
	return &oneOfSchema[int64, RefSchema]{
		types,
		discriminatorFieldName,
	}
}

// OneOfIntType is a serializable version of OneOfIntSchema.
type OneOfIntType[T any] interface {
	OneOfIntSchema[RefType[T]]
	AbstractType[T]
}

// NewOneOfIntType creates a new unserializable polymorphic type with an integer key. The type parameter should
// be an interface describing the underlying types, or any.
func NewOneOfIntType[T any](
	types map[int64]RefType[T],
	discriminatorFieldName string,
) OneOfIntType[T] {
	var defaultValue T
	reflectedType := reflect.TypeOf(defaultValue)
	if reflectedType.Kind() != reflect.Interface {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"The type variable for NewOneOfIntType must be an interface or an any type, %T given",
				defaultValue,
			),
		})
	}

	return &oneOfType[int64, T]{
		oneOfSchema[int64, RefType[T]]{
			types,
			discriminatorFieldName,
		},
	}
}
