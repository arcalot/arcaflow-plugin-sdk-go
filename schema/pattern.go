package schema

import "regexp"

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

// PatternType is the serializable version of PatternSchema.
type PatternType interface {
	PatternSchema
	AbstractType[*regexp.Regexp]
}

// NewPatternType creates a new pattern type for serialization/unserialization.
func NewPatternType() PatternType {
	return &patternType{
		patternSchema{},
	}
}

type patternType struct {
	patternSchema `json:",inline"`
}

func (p patternType) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
}

func (p patternType) UnderlyingType() *regexp.Regexp {
	return &regexp.Regexp{}
}

func (p patternType) Unserialize(data any) (*regexp.Regexp, error) {
	str, err := stringInputMapper(data)
	if err != nil {
		return nil, err
	}
	pattern, err := regexp.Compile(str)
	if err != nil {
		return nil, &ConstraintError{
			Message: "Invalid pattern",
			Cause:   err,
		}
	}
	return pattern, nil
}

func (p patternType) Validate(data *regexp.Regexp) error {
	if data == nil {
		return &ConstraintError{
			Message: "Pattern value should not be nil.",
		}
	}
	return nil
}

func (p patternType) Serialize(data *regexp.Regexp) (any, error) {
	if err := p.Validate(data); err != nil {
		return nil, err
	}
	return data.String(), nil
}
