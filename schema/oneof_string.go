package schema

import (
	"fmt"
	"reflect"
)

// OneOfStringSchema holds the definition of variable types with a string discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use OneOfStringType.
type OneOfStringSchema[RefSchemaType RefSchema] interface {
	OneOfSchema[string, RefSchemaType]
}

// NewOneOfStringSchema creates a new OneOf-type with integer discriminators.
func NewOneOfStringSchema(
	types map[string]RefSchema,
	discriminatorFieldName string,
) OneOfStringSchema[RefSchema] {
	return &oneOfSchema[string, RefSchema]{
		types,
		discriminatorFieldName,
	}
}

// OneOfStringType is a serializable version of OneOfStringSchema.
type OneOfStringType[T any] interface {
	OneOfStringSchema[RefType[T]]
	AbstractType[T]
}

// NewOneOfStringType creates a new unserializable polymorphic type with a string key. The type parameter should
// be an interface describing the underlying types, or any.
func NewOneOfStringType[T any](
	types map[string]RefType[T],
	discriminatorFieldName string,
) OneOfStringType[T] {
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

	return &oneOfType[string, T]{
		oneOfSchema[string, RefType[T]]{
			types,
			discriminatorFieldName,
		},
	}
}
