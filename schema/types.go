package schema

// TypeID is the identifier for types supported in the type system.
type TypeID string

const (
	// TypeIDStringEnum is a type that satisfies the StringEnumSchema.
	TypeIDStringEnum TypeID = "enum_string"
	// TypeIDIntEnum is a type that satisfies the StringIntSchema.
	TypeIDIntEnum TypeID = "enum_integer"
	// TypeIDString is a type that satisfies the StringSchema.
	TypeIDString TypeID = "string"
	// TypeIDPattern is a type that satisfies the PatternSchema.
	TypeIDPattern TypeID = "pattern"
	// TypeIDInt is a type that satisfies the IntSchema.
	TypeIDInt TypeID = "integer"
	// TypeIDFloat is a type that satisfies the FloatSchema.
	TypeIDFloat TypeID = "float"
	// TypeIDBool is a type that satisfies the BoolSchema.
	TypeIDBool TypeID = "bool"
	// TypeIDList is a type that satisfies the ListSchema.
	TypeIDList TypeID = "list"
	// TypeIDMap is a type that satisfies the MapSchema.
	TypeIDMap TypeID = "map"
	// TypeIDScope is a type that satisfies the ScopeSchema.
	TypeIDScope TypeID = "scope"
	// TypeIDObject is a type that satisfies the ObjectSchema.
	TypeIDObject TypeID = "object"
	// TypeIDOneOfString is a type that satisfies the OneOfStringSchema.
	TypeIDOneOfString TypeID = "one_of_string"
	// TypeIDOneOfInt is a type that satisfies the OneOfIntSchema.
	TypeIDOneOfInt TypeID = "one_of_string"
	// TypeIDRef is a type that references an object in a ScopeSchema.
	TypeIDRef TypeID = "ref"
)

// AbstractSchema is the minimum functionality types need to implement.
type AbstractSchema interface {
	TypeID() TypeID
}

// AbstractType describes the common methods all types need to implement.
type AbstractType[T any] interface {
	AbstractSchema
	ApplyScope(ScopeSchema[PropertyType, ObjectType[any]])
	UnderlyingType() T
	Unserialize(data any) (T, error)
	Validate(data T) error
	Serialize(data T) (any, error)
}

// MapKeyType are types that can be used as map keys.
type MapKeyType interface {
	int64 | string
}

// NumberType is a type collection of number types.
type NumberType interface {
	int64 | float64
}
