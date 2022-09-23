package schema

import "reflect"

// TypeID is the identifier for types supported in the type system.
type TypeID string

const (
	// TypeIDStringEnum is a type that satisfies the StringEnum.
	TypeIDStringEnum TypeID = "enum_string"
	// TypeIDIntEnum is a type that satisfies the StringIntSchema.
	TypeIDIntEnum TypeID = "enum_integer"
	// TypeIDString is a type that satisfies the String.
	TypeIDString TypeID = "string"
	// TypeIDPattern is a type that satisfies the Pattern.
	TypeIDPattern TypeID = "pattern"
	// TypeIDInt is a type that satisfies the Int.
	TypeIDInt TypeID = "integer"
	// TypeIDFloat is a type that satisfies the Float.
	TypeIDFloat TypeID = "float"
	// TypeIDBool is a type that satisfies the BoolSchema.
	TypeIDBool TypeID = "bool"
	// TypeIDList is a type that satisfies the List.
	TypeIDList TypeID = "list"
	// TypeIDMap is a type that satisfies the Map.
	TypeIDMap TypeID = "map"
	// TypeIDScope is a type that satisfies the Scope.
	TypeIDScope TypeID = "scope"
	// TypeIDObject is a type that satisfies the Object.
	TypeIDObject TypeID = "object"
	// TypeIDOneOfString is a type that satisfies the OneOfStringSchema.
	TypeIDOneOfString TypeID = "one_of_string"
	// TypeIDOneOfInt is a type that satisfies the OneOfInt.
	TypeIDOneOfInt TypeID = "one_of_string"
	// TypeIDRef is a type that references an object in a Scope.
	TypeIDRef TypeID = "ref"
)

// Serializable describes the minimum feature set a part of the schema hierarchy must implement.
type Serializable interface {
	// ReflectedType returns the underlying unserialized type.
	ReflectedType() reflect.Type
	// Unserialize unserializes the provided data.
	Unserialize(data any) (any, error)
	// Validate validates the specified data in accordance with the schema.
	Validate(data any) error
	// Serialize serializes the provided data.
	Serialize(data any) (any, error)
	// ApplyScope notifies the current schema being added to a scope.
	ApplyScope(scope Scope)
}

// Type adds the type ID to Serializable as part of the Schema tree.
type Type interface {
	Serializable

	// TypeID returns the type of the current schema entry.
	TypeID() TypeID
}

// TypedType provides additional functionality for unserializing types in a type-safe manner.
type TypedType[T any] interface {
	Type

	UnserializeType(data any) (T, error)
	ValidateType(data T) error
	SerializeType(data T) (any, error)
}

// MapKeyType are types that can be used as map keys.
type MapKeyType interface {
	int64 | string
}

// NumberType is a type collection of number types.
type NumberType interface {
	int64 | float64
}
