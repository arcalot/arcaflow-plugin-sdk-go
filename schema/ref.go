package schema

import (
	"fmt"
	"reflect"
)

// RefSchema holds the definition of a reference to a scope-wide object. The ref must always be inside a scope,
// either directly or indirectly. If several scopes are embedded within each other, the Ref references the object
// in the current scope.
//
// This dataclass only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use :class:`RefType`.
type RefSchema interface {
	AbstractSchema
	ID() string
	Display() *DisplayValue
}

// NewRefSchema creates a new reference to an object in a wrapping ScopeSchema by ID.
func NewRefSchema(id string, display *DisplayValue) RefSchema {
	return &refSchema{
		id,
		display,
	}
}

type refSchema struct {
	IDValue      string        `json:"id"`
	DisplayValue *DisplayValue `json:"display"`
}

func (r refSchema) TypeID() TypeID {
	return TypeIDRef
}

func (r refSchema) ID() string {
	return r.IDValue
}

func (r refSchema) Display() *DisplayValue {
	return r.DisplayValue
}

// RefType is a serializable version of RefSchema.
type RefType[T any] interface {
	RefSchema
	AbstractType[T]
	HasProperty(propertyID string) bool

	Anonymous() RefType[any]
}

// NewRefType creates a serializable reference to a scope. The ApplyScope function must be called after creation to link
// it with the scope.
func NewRefType[T any](
	id string,
	display *DisplayValue,
) RefType[T] {
	return &refType[T]{
		refSchema{
			id,
			display,
		},
		nil,
	}
}

type refType[T any] struct {
	refSchema             `json:",inline"`
	referencedObjectCache ObjectType[any]
}

func (r *refType[T]) HasProperty(propertyID string) bool {
	_, ok := r.referencedObjectCache.Properties()[propertyID]
	return ok
}

func (r *refType[T]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	objects := s.Objects()
	referencedObject, ok := objects[r.IDValue]
	if !ok {
		panic(BadArgumentError{
			Message: fmt.Sprintf("Referenced object %s not found in scope", r.IDValue),
		})
	}
	underlyingType := referencedObject.UnderlyingType()
	underlyingTypeType := reflect.TypeOf(underlyingType)
	var defaultValue T
	defaultValueType := reflect.TypeOf(defaultValue)
	if underlyingTypeType.Kind() != defaultValueType.Kind() {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"Referenced object '%s' underlying type '%T' does not match reference type '%T",
				r.IDValue,
				underlyingType,
				defaultValue,
			),
		})
	}
	r.referencedObjectCache = referencedObject
}

func (r *refType[T]) UnderlyingType() T {
	var defaultValue T
	return defaultValue
}

func (r *refType[T]) Unserialize(data any) (T, error) {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Unserialize called before ApplyScope. Did you add your RefType to a scope?",
		})
	}
	result, err := r.referencedObjectCache.Unserialize(data)
	return result.(T), err
}

func (r *refType[T]) Validate(data T) error {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Unserialize called before ApplyScope. Did you add your RefType to a scope?",
		})
	}
	return r.referencedObjectCache.Validate(data)
}

func (r *refType[T]) Serialize(data T) (any, error) {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Unserialize called before ApplyScope. Did you add your RefType to a scope?",
		})
	}
	return r.referencedObjectCache.Serialize(data)
}

func (r *refType[T]) Anonymous() RefType[any] {
	return &anonymousRefType[T]{
		*r,
	}
}

type anonymousRefType[T any] struct {
	refType[T] `json:",inline"`
}

func (a *anonymousRefType[T]) UnderlyingType() any {
	return a.refType.UnderlyingType()
}

func (a *anonymousRefType[T]) Unserialize(data any) (any, error) {
	return a.refType.Unserialize(data)
}

func (a *anonymousRefType[T]) Validate(data any) error {
	typedData, ok := data.(T)
	if !ok {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not usable as %T", data, typedData),
		}
	}
	return a.refType.Validate(typedData)
}

func (a *anonymousRefType[T]) Serialize(data any) (any, error) {
	typedData, ok := data.(T)
	if !ok {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("%T is not usable as %T", data, typedData),
		}
	}
	return a.refType.Serialize(typedData)
}
