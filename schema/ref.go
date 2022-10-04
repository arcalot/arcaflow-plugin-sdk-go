package schema

import (
	"fmt"
	"reflect"
)

// Ref holds the definition of a reference to a scope-wide object. The ref must always be inside a scope,
// either directly or indirectly. If several scopes are embedded within each other, the Ref references the object
// in the current scope.
type Ref interface {
	Object

	ID() string
	Display() Display
	GetObject() Object
}

// NewRefSchema creates a new reference to an object in a wrapping Scope by ID.
func NewRefSchema(id string, display Display) *RefSchema {
	return &RefSchema{
		id,
		display,

		nil,
	}
}

type RefSchema struct {
	IDValue      string  `json:"id"`
	DisplayValue Display `json:"display"`

	referencedObjectCache Object
}

func (r *RefSchema) Properties() map[string]*PropertySchema {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Properties was called before ApplyScope!",
		})
	}
	return r.referencedObjectCache.Properties()
}

func (r *RefSchema) GetDefaults() map[string]any {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "GetDefaults was called before ApplyScope!",
		})
	}
	return r.referencedObjectCache.GetDefaults()
}

func (r *RefSchema) TypeID() TypeID {
	return TypeIDRef
}

func (r *RefSchema) GetObject() Object {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "GetObject was called before ApplyScope!",
		})
	}
	return r.referencedObjectCache
}

func (r *RefSchema) ReflectedType() reflect.Type {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "ReflectedType was called before ApplyScope!",
		})
	}
	return r.referencedObjectCache.ReflectedType()
}

func (r *RefSchema) ID() string {
	return r.IDValue
}

func (r *RefSchema) Display() Display {
	return r.DisplayValue
}

func (r *RefSchema) ApplyScope(scope Scope) {
	objects := scope.Objects()
	referencedObject, ok := objects[r.IDValue]
	if !ok {
		panic(BadArgumentError{
			Message: fmt.Sprintf("Referenced object %s not found in scope", r.IDValue),
		})
	}
	r.referencedObjectCache = referencedObject
}

func (r *RefSchema) Unserialize(data any) (any, error) {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Unserialize called before ApplyScope. Did you add your RefType to a scope?",
		})
	}
	return r.referencedObjectCache.Unserialize(data)
}

func (r *RefSchema) Validate(data any) error {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Unserialize called before ApplyScope. Did you add your RefType to a scope?",
		})
	}
	return r.referencedObjectCache.Validate(data)
}

func (r *RefSchema) Serialize(data any) (any, error) {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: "Unserialize called before ApplyScope. Did you add your RefType to a scope?",
		})
	}
	return r.referencedObjectCache.Serialize(data)
}
