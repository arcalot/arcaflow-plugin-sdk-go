package schema

import (
	"fmt"
	"reflect"
)

// Scope is a container for holding objects that can be referenced. It also optionally holds a reference to the
// root object of the current scope. References within the scope must always reference IDs in a scope. Scopes can be
// embedded into other objects, and scopes can have subscopes. Each Ref will reference objects in its current
// scope.
type Scope interface {
	Type
	Objects() map[string]*ObjectSchema
	Root() string
}

// NewScopeSchema returns a new scope.
func NewScopeSchema(rootObject *ObjectSchema, objects ...*ObjectSchema) *ScopeSchema {
	objectMap := make(map[string]*ObjectSchema, len(objects)+1)
	root := rootObject.ID()
	objectMap[rootObject.ID()] = rootObject
	for _, object := range objects {
		objectMap[object.ID()] = object
	}

	schema := &ScopeSchema{
		objectMap,
		root,
	}

	for _, v := range objectMap {
		v.ApplyScope(schema)
	}

	return schema
}

type ScopeSchema struct {
	ObjectsValue map[string]*ObjectSchema `json:"objects"`
	RootValue    string                   `json:"root,omitempty"`
}

func (s ScopeSchema) ReflectedType() reflect.Type {
	return s.ObjectsValue[s.RootValue].ReflectedType()
}

func (s ScopeSchema) Unserialize(data any) (any, error) {
	return s.ObjectsValue[s.RootValue].Unserialize(data)
}

func (s ScopeSchema) Validate(data any) error {
	return s.ObjectsValue[s.RootValue].Validate(data)
}

func (s ScopeSchema) Serialize(data any) (any, error) {
	return s.ObjectsValue[s.RootValue].Serialize(data)
}

func (s ScopeSchema) ApplyScope(_ Scope) {

}

func (s ScopeSchema) TypeID() TypeID {
	return TypeIDScope
}

func (s ScopeSchema) Objects() map[string]*ObjectSchema {
	return s.ObjectsValue
}

func (s ScopeSchema) Root() string {
	return s.RootValue
}

// NewTypedScopeSchema returns a new scope that is typed.
func NewTypedScopeSchema[T any](rootObject *ObjectSchema, objects ...*ObjectSchema) *TypedScopeSchema[T] {
	var defaultValue T
	if reflect.TypeOf(defaultValue) != rootObject.ReflectedType() {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"Incorrect type definition: type %T does not match the root object type of %s",
				defaultValue,
				rootObject.ReflectedType().Name(),
			),
		})
	}

	return &TypedScopeSchema[T]{
		*NewScopeSchema(rootObject, objects...),
	}
}

// TypedScopeSchema is a typed variant of the ScopeSchema, allowing for direct type use. This should not be used in full
// schema definitions as the type parameter will prevent it from being added to lists thanks to the simplistic
// generics system in Go.
type TypedScopeSchema[T any] struct {
	ScopeSchema `json:",inline"`
}

func (t TypedScopeSchema[T]) UnserializeType(data any) (result T, err error) {
	untypedResult, err := t.Unserialize(data)
	if err != nil {
		return result, err
	}
	return untypedResult.(T), nil
}

func (t TypedScopeSchema[T]) ValidateType(data T) error {
	return t.Validate(data)
}

func (t TypedScopeSchema[T]) SerializeType(data T) (any, error) {
	return t.Serialize(data)
}
