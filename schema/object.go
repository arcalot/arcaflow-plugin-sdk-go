package schema

// ObjectSchema holds the definition for objects comprised of defined fields. This dataclass only has the ability to hold
// the configuration but cannot serialize, unserialize or validate. For that functionality please use
// PropertyType.
type ObjectSchema interface {
	AbstractSchema
	ID() string
	Properties() map[string]PropertySchema
}

// NewObjectSchema creates a new object definition.
func NewObjectSchema(id string, properties map[string]PropertySchema) ObjectSchema {
	return &objectSchema{
		id,
		properties,
	}
}

type objectSchema struct {
	IDValue         string                    `json:"id"`
	PropertiesValue map[string]PropertySchema `json:"properties"`
}

func (o objectSchema) TypeID() TypeID {
	return TypeIDObject
}

func (o objectSchema) ID() string {
	return o.IDValue
}

func (o objectSchema) Properties() map[string]PropertySchema {
	return o.PropertiesValue
}
