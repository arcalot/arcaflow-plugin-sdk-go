package schema

import (
	"fmt"
	"reflect"
)

// Ref holds the definition of a reference to a scope-wide object. The ref must always be inside a scope,
// either directly or indirectly. If several scopes are embedded within each other, the Ref references the object
// in the scope specified. SelfNamespace refers to the current scope.
type Ref interface {
	Object

	ID() string
	Namespace() string
	Display() Display
	GetObject() Object
	ObjectReady() bool
}

// NewRefSchema creates a new reference to an object in a wrapping Scope by ID.
func NewRefSchema(id string, display Display) *RefSchema {
	return NewNamespacedRefSchema(id, SelfNamespace, display)
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

	ReferencedObjectCache Object
}

func (r *RefSchema) Properties() map[string]*PropertySchema {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in Properties; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace),
		})
	}
	return r.ReferencedObjectCache.Properties()
}

func (r *RefSchema) GetDefaults() map[string]any {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in GetDefaults; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace),
		})
	}
	return r.ReferencedObjectCache.GetDefaults()
}

func (r *RefSchema) TypeID() TypeID {
	return TypeIDRef
}

func (r *RefSchema) GetObject() Object {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in GetObject; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace),
		})
	}
	return r.ReferencedObjectCache
}

func (r *RefSchema) ReflectedType() reflect.Type {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in ReflectedType; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace),
		})
	}
	return r.ReferencedObjectCache.ReflectedType()
}

func (r *RefSchema) ObjectReady() bool {
	return r.ReferencedObjectCache != nil
}

func (r *RefSchema) ID() string {
	return r.IDValue
}

func (r *RefSchema) Namespace() string {
	return r.ObjectNamespace
}

func (r *RefSchema) Display() Display {
	return r.DisplayValue
}

// ApplyNamespace links the reference to the object if the given namespace
// matches the ref's namespace. Other namespaces are skipped.
func (r *RefSchema) ApplyNamespace(objects map[string]*ObjectSchema, namespace string) {
	if namespace != r.ObjectNamespace {
		return // This reference does not refer to an object in the supplied namespace.
	}
	referencedObject, ok := objects[r.IDValue]
	if !ok {
		availableObjects := ""
		for objectID := range objects {
			availableObjects += objectID + "\n"
		}
		panic(BadArgumentError{
			Message: fmt.Sprintf("Referenced object '%s' not found in scope with namespace %q; available:\n%s", r.IDValue, namespace, availableObjects),
		})
	}
	r.ReferencedObjectCache = referencedObject
}

func (r *RefSchema) ValidateReferences() error {
	if r.ReferencedObjectCache != nil {
		return nil // Success
	}
	// The only way, unless there is a bug, for it to get here is if ApplyNamespace was not called with the
	// correct namespace, or if the code disregards the error returned by ApplyNamespace. ApplyNamespace should
	// always set ReferencedObjectCache or return an error if the correct namespace is applied.
	return BadArgumentError{
		Message: fmt.Sprintf(
			"Ref object reference missing its link to object with ID %q in namespace %q. Namespace not valid (not applied).",
			r.IDValue,
			r.ObjectNamespace,
		),
	}
}

func (r *RefSchema) Unserialize(data any) (any, error) {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in Unserialize; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace,
			),
		})
	}
	return r.ReferencedObjectCache.Unserialize(data)
}

func (r *RefSchema) Validate(data any) error {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in Validate; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace,
			),
		})
	}
	return r.ReferencedObjectCache.Validate(data)
}

func (r *RefSchema) ValidateCompatibility(typeOrData any) error {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in ValidateCompatibility; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace,
			),
		})
	}
	schemaType, ok := typeOrData.(*RefSchema)
	if ok {
		return r.ReferencedObjectCache.ValidateCompatibility(schemaType.ReferencedObjectCache)
	}
	return r.ReferencedObjectCache.ValidateCompatibility(typeOrData)
}

func (r *RefSchema) Serialize(data any) (any, error) {
	if r.ReferencedObjectCache == nil {
		panic(BadArgumentError{
			Message: fmt.Sprintf(
				"ref type not linked to its object with ID %q in Serialize; scope with namespace %q was not applied successfully",
				r.IDValue, r.ObjectNamespace,
			),
		})
	}
	return r.ReferencedObjectCache.Serialize(data)
}

func (r *RefSchema) SerializeForHuman(args map[string]any) any {
	return r.ReferencedObjectCache.SerializeForHuman(args)
}
