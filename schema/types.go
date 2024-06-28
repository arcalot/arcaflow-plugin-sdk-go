package schema

import (
	"fmt"
	"reflect"
)

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
	TypeIDOneOfInt TypeID = "one_of_int"
	// TypeIDRef is a type that references an object in a Scope.
	TypeIDRef TypeID = "ref"
	// TypeIDAny refers to an any type. This type essentially amounts to unchecked types, as long as they are:
	//
	// - maps
	// - lists
	// - int64
	// - float64
	// - string
	// - bool
	//
	// No other types are accepted.
	TypeIDAny TypeID = "any"
)

const SelfNamespace string = ""

// Serializable describes the minimum feature set a part of the schema hierarchy must implement.
type Serializable interface {
	// ReflectedType returns the underlying unserialized type.
	ReflectedType() reflect.Type
	// Unserialize unserializes the provided data.
	Unserialize(data any) (any, error)
	// Validate validates the specified unserialized data in accordance with the schema.
	Validate(data any) error
	// ValidateCompatibility validates the specified serialized data or schema is compatible with the schema.
	ValidateCompatibility(typeOrData any) error
	// Serialize serializes the provided data.
	Serialize(data any) (any, error)
	// ApplyNamespace makes namespace object available to resolve references.
	ApplyNamespace(objects map[string]*ObjectSchema, namespace string)
	// ValidateReferences validates that all references had their referenced objects found.
	// Useful to ensure the error is caught early rather than later when it's used.
	ValidateReferences() error
	SerializeForHuman(args map[string]any) any
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

// ScalarType is a struct that provides default implementations for
// ApplyNamespace and ValidateReferences for types that cannot contain
// references.
type ScalarType struct {
}

func (s *ScalarType) ApplyNamespace(_ map[string]*ObjectSchema, _ string) {
	// Scalar types have no references, so the namespace can be ignored.
}

func (s *ScalarType) ValidateReferences() error {
	return nil // Scalar types have no references, so no work to do.
}

// MapKeyType are types that can be used as map keys.
type MapKeyType interface {
	int64 | string
}

// NumberType is a type collection of number types.
type NumberType interface {
	int64 | float64
}

func saveConvertTo(value any, to reflect.Type) (any, error) {
	var recoveredError error
	var result any
	func() {
		defer func() {
			e := recover()
			if e != nil {
				var ok bool
				recoveredError, ok = e.(error)
				if !ok {
					recoveredError = fmt.Errorf("%v", e)
				}
			}
		}()
		result = reflect.ValueOf(value).Convert(to).Interface()
	}()
	if recoveredError != nil {
		return nil, &ConstraintError{
			Message: fmt.Sprintf(
				"%T cannot be converted to %s",
				value,
				to.String(),
			),
			Cause: recoveredError,
		}
	}
	return result, nil
}

//func (s Serializable) SerializeForHuman(args map[string]any) map[string]any {
//
//}
