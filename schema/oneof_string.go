package schema

// OneOfStringSchema holds the definition of variable types with a string discriminator. This type acts as a split for a
// case where multiple possible object types can be present in a field. This type requires that there be a common field
// (the discriminator) which tells a parsing party which type it is. The field type in this case is a string.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use OneOfStringType.
type OneOfStringSchema interface {
	OneOfSchema[string]
}

// NewOneOfStringSchema creates a new OneOf-type with integer discriminators.
func NewOneOfStringSchema(
	types map[string]RefSchema,
	discriminatorFieldName string,
) OneOfStringSchema {
	return &oneOfStringSchema{
		oneOfSchema[string]{
			types,
			discriminatorFieldName,
		},
	}
}

type oneOfStringSchema struct {
	oneOfSchema[string] `json:",inline"`
}

func (o oneOfStringSchema) TypeID() TypeID {
	return TypeIDOneOfString
}
