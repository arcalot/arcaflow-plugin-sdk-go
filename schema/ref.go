package schema

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
