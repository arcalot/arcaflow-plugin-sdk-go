package schema

// ScopeSchema is a container for holding objects that can be referenced. It also optionally holds a reference to the
// root object of the current scope. References within the scope must always reference IDs in a scope. Scopes can be
// embedded into other objects, and scopes can have subscopes. Each RefSchema will reference objects in its current
// scope.
//
// This schema only has the ability to hold the configuration but cannot serialize, unserialize or validate. For
// that functionality please use ScopeType.
type ScopeSchema interface {
	AbstractSchema
	Objects() map[string]ObjectSchema
	Root() *string
}

// NewScopeSchema returns a new scope.
func NewScopeSchema(objects map[string]ObjectSchema, root *string) ScopeSchema {
	return &scopeSchema{
		objects,
		root,
	}
}

type scopeSchema struct {
	ObjectsValue map[string]ObjectSchema `json:"objects"`
	RootValue    *string                 `json:"root,omitempty"`
}

func (s scopeSchema) TypeID() TypeID {
	return TypeIDScope
}

func (s scopeSchema) Objects() map[string]ObjectSchema {
	return s.ObjectsValue
}

func (s scopeSchema) Root() *string {
	return s.RootValue
}
