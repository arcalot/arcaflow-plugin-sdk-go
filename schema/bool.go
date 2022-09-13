package schema

// BoolSchema holds the schema information for boolean types. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use BoolType.
type BoolSchema interface {
	AbstractSchema
}

// NewBoolSchema creates a new boolean representation.
func NewBoolSchema() BoolSchema {
	return &boolSchema{}
}

type boolSchema struct {
}

func (b boolSchema) TypeID() TypeID {
	return TypeIDBool
}
