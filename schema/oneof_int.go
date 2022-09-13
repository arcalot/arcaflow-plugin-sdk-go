package schema

// OneOfIntSchema holds the definition of variable types with an integer discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use OneOfIntType.
type OneOfIntSchema interface {
	OneOfSchema[int64]
}

// NewOneOfIntSchema creates a new OneOf-type with integer discriminators.
func NewOneOfIntSchema(
	types map[int64]RefSchema,
	discriminatorFieldName string,
) OneOfIntSchema {
	return &oneOfIntSchema{
		oneOfSchema[int64]{
			types,
			discriminatorFieldName,
		},
	}
}

type oneOfIntSchema struct {
	oneOfSchema[int64] `json:",inline"`
}

func (o oneOfIntSchema) TypeID() TypeID {
	return TypeIDOneOfInt
}
