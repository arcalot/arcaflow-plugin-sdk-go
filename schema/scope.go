package schema

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
	Root() *string
}

// NewScopeSchema returns a new scope.
func NewScopeSchema[P PropertySchema, T ObjectSchema[P]](objects map[string]T, root *string) ScopeSchema[P, T] {
	return &scopeSchema[P, T]{
		objects,
		root,
	}
}

type scopeSchema[P PropertySchema, T ObjectSchema[P]] struct {
	ObjectsValue map[string]T `json:"objects"`
	RootValue    *string      `json:"root,omitempty"`
}

func (s scopeSchema[P, T]) TypeID() TypeID {
	return TypeIDScope
}

func (s scopeSchema[P, T]) Objects() map[string]T {
	return s.ObjectsValue
}

func (s scopeSchema[P, T]) Root() *string {
	return s.RootValue
}
