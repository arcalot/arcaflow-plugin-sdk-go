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
		DEFAULT_NAMESPACE,
		nil,
	}
}

// NewNamespacedRefSchema creates a new reference to an object in a wrapping Scope by ID and namespace.
func NewNamespacedRefSchema(id string, namespace string, display Display) *RefSchema {
	return &RefSchema{
		id,
		display,
		namespace,
		nil,
	}
}

type RefSchema struct {
	IDValue         string  `json:"id"`
	DisplayValue    Display `json:"display"`
	ObjectNamespace string  `json:"namespace"`

	referencedObjectCache Object
}

func (r *RefSchema) Properties() map[string]*PropertySchema {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf("Properties was called before ApplyScope with namespace %q", r.ObjectNamespace),
		})
	}
	return r.referencedObjectCache.Properties()
}

func (r *RefSchema) GetDefaults() map[string]any {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf("GetDefaults was called before ApplyScope with namespace %q", r.ObjectNamespace),
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
			Message: fmt.Sprintf("GetObject was called before ApplyScope with namespace %q", r.ObjectNamespace),
		})
	}
	return r.referencedObjectCache
}

func (r *RefSchema) ReflectedType() reflect.Type {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf("ReflectedType was called before ApplyScope with namespace %q", r.ObjectNamespace),
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

func (r *RefSchema) ApplyScope(scope Scope, namespace string) {
	if namespace != r.ObjectNamespace {
		return // The scope does not apply to this reference.
	}
	objects := scope.Objects()
	referencedObject, ok := objects[r.IDValue]
	if !ok {
		panic(BadArgumentError{
			Message: fmt.Sprintf("Referenced object '%s' not found in scope with namespace %q", r.IDValue, namespace),
		})
	}
	r.referencedObjectCache = referencedObject
}

func (r *RefSchema) ValidateReferences() error {
	if r.referencedObjectCache == nil {
		return BadArgumentError{
			Message: fmt.Sprintf(
				"Ref object reference could not find an object with ID %q in namespace %q",
				r.IDValue,
				r.ObjectNamespace,
			),
		}
	}

	return nil
}

func (r *RefSchema) Unserialize(data any) (any, error) {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"Unserialize called before ApplyScope. Did you add your RefType to a scope with the namespace %q?",
				r.ObjectNamespace,
			),
		})
	}
	return r.referencedObjectCache.Unserialize(data)
}

func (r *RefSchema) Validate(data any) error {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"Validate called before ApplyScope. Did you add your RefType to a scope with the namespace %q?",
				r.ObjectNamespace,
			),
		})
	}
	return r.referencedObjectCache.Validate(data)
}

func (r *RefSchema) ValidateCompatibility(typeOrData any) error {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ValidateCompatibility called before ApplyScope. Did you add your RefType to a scope with the namespace %q?",
				r.ObjectNamespace,
			),
		})
	}
	schemaType, ok := typeOrData.(*RefSchema)
	if ok {
		return r.referencedObjectCache.ValidateCompatibility(schemaType.referencedObjectCache)
	}
	return r.referencedObjectCache.ValidateCompatibility(typeOrData)
}

func (r *RefSchema) Serialize(data any) (any, error) {
	if r.referencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"Serialize called before ApplyScope. Did you add your RefType to a scope with the namespace %q?",
				r.ObjectNamespace,
			),
		})
	}
	return r.referencedObjectCache.Serialize(data)
}
