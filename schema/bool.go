package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// Bool holds the schema information for boolean types.
type Bool interface {
	TypedType[bool]
}

// NewBoolSchema creates a new boolean representation.
func NewBoolSchema() *BoolSchema {
	return &BoolSchema{}
}

// BoolSchema holds the schema information for boolean types.
type BoolSchema struct {
	Type TypeID `json:"type_id" yaml:"type_id"`
}

func (b BoolSchema) Unserialize(data any) (any, error) {
	intConverter := func(data int64) (bool, error) {
		switch data {
		case 1:
			return true, nil
		case 0:
			return false, nil
		default:
			return false, fmt.Errorf("'%d' is not a valid boolean value", data)
		}
	}
	switch v := data.(type) {
	case bool:
		return v, nil
	case string:
		lowerStr := strings.ToLower(v)
		if serializedValue, ok := boolStringValues[lowerStr]; ok {
			return serializedValue, nil
		}
	case int:
		return intConverter(int64(v))
	case uint:
		return intConverter(int64(v))
	case int64:
		return intConverter(v)
	case uint64:
		return intConverter(int64(v))
	case int32:
		return intConverter(int64(v))
	case uint32:
		return intConverter(int64(v))
	case int16:
		return intConverter(int64(v))
	case uint16:
		return intConverter(int64(v))
	case int8:
		return intConverter(int64(v))
	case uint8:
		return intConverter(int64(v))
	}
	return false, fmt.Errorf("'%v' is not a valid boolean value", data)
}

func (b BoolSchema) UnserializeType(data any) (bool, error) {
	unserialized, err := b.Unserialize(data)
	if err != nil {
		return false, err
	}
	return unserialized.(bool), nil
}

func (b BoolSchema) Validate(data any) error {
	_, err := b.Serialize(data)
	return err
}

func (b BoolSchema) ValidateType(data bool) error {
	return b.Validate(data)
}

func (b BoolSchema) Serialize(d any) (any, error) {
	return asBool(d)
}

func asBool(d any) (bool, error) {
	data, ok := d.(bool)
	if !ok {
		var i bool
		intType := reflect.TypeOf(i)
		dValue := reflect.ValueOf(d)
		if !dValue.CanConvert(intType) {
			return false, &ConstraintError{
				Message: fmt.Sprintf("%T is not a valid data type for a bool schema.", d),
			}
		}
		data = dValue.Convert(intType).Bool()
	}
	return data, nil
}

func (b BoolSchema) SerializeType(data bool) (any, error) {
	return b.Serialize(data)
}

func (b BoolSchema) ApplyScope(scope Scope) {

}

func (b BoolSchema) TypeID() TypeID {
	return TypeIDBool
}

func (b BoolSchema) ReflectedType() reflect.Type {
	return reflect.TypeOf(false)
}

var boolStringValues = map[string]bool{
	"1":        true,
	"yes":      true,
	"y":        true,
	"on":       true,
	"true":     true,
	"enable":   true,
	"enabled":  true,
	"0":        false,
	"no":       false,
	"n":        false,
	"off":      false,
	"false":    false,
	"disable":  false,
	"disabled": false,
}
