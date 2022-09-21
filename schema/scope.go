package schema

import (
	"fmt"
	"reflect"
)

// ScopeSchema is a container for holding objects that can be referenced. It also optionally holds a reference to the
// root object of the current scope. References within the scope must always reference IDs in a scope. Scopes can be
// embedded into other objects, and scopes can have subscopes. Each RefSchema will reference objects in its current
// scope.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use ScopeType.
type ScopeSchema[P PropertySchema, T ObjectSchema[P]] interface {
	AbstractSchema
	Objects() map[string]T
	Root() string
}

// NewScopeSchema returns a new scope.
func NewScopeSchema[P PropertySchema, T ObjectSchema[P]](objects map[string]T, root string) ScopeSchema[P, T] {
	return &abstractScopeSchema[P, T]{
		objects,
		root,
	}
}

type abstractScopeSchema[P PropertySchema, T ObjectSchema[P]] struct {
	ObjectsValue map[string]T `json:"objects"`
	RootValue    string       `json:"root,omitempty"`
}

//nolint:unused
type scopeSchema struct {
	abstractScopeSchema[*propertySchema, *objectSchema] `json:",inline"`
}

func (s abstractScopeSchema[P, T]) TypeID() TypeID {
	return TypeIDScope
}

func (s abstractScopeSchema[P, T]) Objects() map[string]T {
	return s.ObjectsValue
}

func (s abstractScopeSchema[P, T]) Root() string {
	return s.RootValue
}

// ScopeType is a serializable version of ScopeSchema.
type ScopeType[T any] interface {
	AbstractType[T]
	ScopeSchema[PropertyType, ObjectType[any]]

	Any() ScopeType[any]
}

// NewScopeType declares a new scope that can be unserialized using the passed objects.
func NewScopeType[T any](objects map[string]ObjectType[any], root string) ScopeType[T] {
	rootObject, ok := objects[root]
	if !ok {
		panic(BadArgumentError{
			Message: fmt.Sprintf("Declared root object '%s' not found in scope", root),
		})
	}

	rootUnderlyingType := rootObject.UnderlyingType()
	var defaultValue T
	if reflect.TypeOf(rootUnderlyingType).Kind() != reflect.TypeOf(defaultValue).Kind() {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"The declared root object %s (type: %T) does not match the object in the scope (%T)",
				root,
				rootUnderlyingType,
				defaultValue,
			),
		})
	}

	schema := abstractScopeSchema[PropertyType, ObjectType[any]]{
		objects,
		root,
	}

	for _, v := range objects {
		v.ApplyScope(schema)
	}

	return &scopeType[T]{
		schema,
		rootObject,
	}
}

type scopeType[T any] struct {
	abstractScopeSchema[PropertyType, ObjectType[any]] `json:",inline"`
	rootObject                                         ObjectType[any]
}

func (s scopeType[T]) Any() ScopeType[any] {
	return &scopeType[any]{
		abstractScopeSchema: s.abstractScopeSchema,
		rootObject:          s.rootObject,
	}
}

func (s scopeType[T]) ApplyScope(_ ScopeSchema[PropertyType, ObjectType[any]]) {

}

func (s scopeType[T]) UnderlyingType() T {
	return s.rootObject.UnderlyingType().(T)
}

func (s scopeType[T]) Unserialize(data any) (typedResult T, err error) {
	result, err := s.rootObject.Unserialize(data)
	if err != nil {
		return typedResult, err
	}
	typedResult, ok := result.(T)
	if !ok {
		return typedResult, &ConstraintError{
			Message: fmt.Sprintf("Failed to convert %T to %T", result, typedResult),
		}
	}
	return typedResult, nil
}

func (s scopeType[T]) Validate(data T) error {
	return s.rootObject.Validate(data)
}

func (s scopeType[T]) Serialize(data T) (any, error) {
	return s.rootObject.Serialize(data)
}
