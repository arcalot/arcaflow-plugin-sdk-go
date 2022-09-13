package schema

// ListSchema holds the schema definition for lists. This dataclass only has the ability to hold the configuration but
// cannot serialize, unserialize or validate. For that functionality please use ListType.
type ListSchema interface {
	AbstractSchema
	Items() AbstractSchema
	Min() *int64
	Max() *int64
}

// NewListSchema creates a new list schema from the specified values.
func NewListSchema(items AbstractSchema, min *int64, max *int64) ListSchema {
	return &listSchema{}
}

type listSchema struct {
	ItemsValue AbstractSchema
	MinValue   *int64
	MaxValue   *int64
}

func (l listSchema) TypeID() TypeID {
	return TypeIDList
}

func (l listSchema) Items() AbstractSchema {
	return l.ItemsValue
}

func (l listSchema) Min() *int64 {
	return l.MinValue
}

func (l listSchema) Max() *int64 {
	return l.MaxValue
}
