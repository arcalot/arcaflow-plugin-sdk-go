package schema

// PatternSchema holds the schema information for regular expression patterns. This dataclass only has the ability to
// hold the configuration but cannot serialize, unserialize or validate. For that functionality please use
// PatternType.
type PatternSchema interface {
	AbstractSchema
}

// NewPatternSchema creates a new pattern schema.
func NewPatternSchema() PatternSchema {
	return &patternSchema{}
}

type patternSchema struct {
}

func (p patternSchema) TypeID() TypeID {
	return TypeIDPattern
}
