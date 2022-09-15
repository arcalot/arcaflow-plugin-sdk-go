package schema

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
type OneOfIntType interface {
	OneOfIntSchema[RefType[any]]
	AbstractType[any]
}

// NewOneOfIntType creates a new unserializable polymorphic type with an integer key. The type parameter should
// be an interface describing the underlying types, or any.
func NewOneOfIntType(
	types map[int64]RefType[any],
	discriminatorFieldName string,
) OneOfIntType {
	return &oneOfType[int64]{
		oneOfSchema[int64, RefType[any]]{
			types,
			discriminatorFieldName,
		},
	}
}
