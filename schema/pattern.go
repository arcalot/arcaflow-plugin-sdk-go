package schema

import (
	"fmt"
	"reflect"
	"regexp"
)

// Pattern holds the schema information for regular expression patterns. This dataclass only has the ability to
// hold the configuration but cannot serialize, unserialize or validate. For that functionality please use
// PatternType.
type Pattern interface {
	TypedType[*regexp.Regexp]
}

// NewPatternSchema creates a new pattern schema.
func NewPatternSchema() *PatternSchema {
	return &PatternSchema{}
}

type PatternSchema struct {
	Type TypeID `json:"type_id" yaml:"type_id"`
}

func (p PatternSchema) TypeID() TypeID {
	return TypeIDPattern
}

func (p PatternSchema) ApplyScope(scope Scope) {
}

func (p PatternSchema) ReflectedType() reflect.Type {
	return reflect.TypeOf(&regexp.Regexp{})
}

func (p PatternSchema) Unserialize(data any) (any, error) {
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

func (p PatternSchema) Validate(d any) error {
	if d == nil {
		return &ConstraintError{
			Message: "Pattern value should not be nil.",
		}
	}

	_, ok := d.(*regexp.Regexp)
	if !ok {
		return &ConstraintError{
			Message: fmt.Sprintf("%T is not a valid data type for a float schema.", d),
		}
	}
	return nil
}

func (p PatternSchema) Serialize(data any) (any, error) {
	if err := p.Validate(data); err != nil {
		return nil, err
	}
	return data.(*regexp.Regexp).String(), nil
}

func (p PatternSchema) UnserializeType(data any) (*regexp.Regexp, error) {
	result, err := p.Unserialize(data)
	if err != nil {
		return nil, err
	}
	return result.(*regexp.Regexp), nil
}

func (p PatternSchema) ValidateType(data *regexp.Regexp) error {
	return p.Validate(data)
}

func (p PatternSchema) SerializeType(data *regexp.Regexp) (any, error) {
	return p.Serialize(data)
}
