package schema

import "regexp"

// StringSchema holds schema information for strings. This dataclass only has the ability to hold the configuration but
// cannot serialize, unserialize or validate. For that functionality please use StringType.
type StringSchema interface {
	AbstractSchema
	Min() *int64
	Max() *int64
	Pattern() *regexp.Regexp
}

// NewStringSchema creates a new string schema.
func NewStringSchema(min *int64, max *int64, pattern *regexp.Regexp) StringSchema {
	return &stringSchema{
		min,
		max,
		pattern,
	}
}

type stringSchema struct {
	MinValue     *int64         `json:"min"`
	MaxValue     *int64         `json:"max"`
	PatternValue *regexp.Regexp `json:"pattern"`
}

func (s stringSchema) TypeID() TypeID {
	return TypeIDString
}

func (s stringSchema) Min() *int64 {
	return s.MinValue
}

func (s stringSchema) Max() *int64 {
	return s.MaxValue
}

func (s stringSchema) Pattern() *regexp.Regexp {
	return s.PatternValue
}
