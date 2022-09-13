package schema

import "fmt"

// MapSchema holds the schema definition for key-value associations. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use MapType.
type MapSchema interface {
	AbstractSchema
	Keys() AbstractSchema
	Values() AbstractSchema
	Min() *int64
	Max() *int64
}

// NewMapSchema creates a new map schema.
func NewMapSchema(keys AbstractSchema, values AbstractSchema, min *int64, max *int64) MapSchema {
	switch keys.TypeID() {
	case TypeIDString:
	case TypeIDInt:
	case TypeIDStringEnum:
	case TypeIDIntEnum:
	default:
		panic(BadArgumentError{
			Message: fmt.Sprintf("Invalid type ID for map: %s, expected one of: string, int", keys.TypeID()),
		})
	}

	return &mapSchema{
		keys,
		values,
		min,
		max,
	}
}

type mapSchema struct {
	KeysValue   AbstractSchema
	ValuesValue AbstractSchema
	MinValue    *int64
	MaxValue    *int64
}

func (m mapSchema) TypeID() TypeID {
	return TypeIDMap
}

func (m mapSchema) Keys() AbstractSchema {
	return m.KeysValue
}

func (m mapSchema) Values() AbstractSchema {
	return m.ValuesValue
}

func (m mapSchema) Min() *int64 {
	return m.MinValue
}

func (m mapSchema) Max() *int64 {
	return m.MaxValue
}
