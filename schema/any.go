package schema

import (
	"fmt"
	"reflect"
)

// NewAnySchema creates an AnySchema which is a wildcard allowing maps, lists, integers, strings, bools. and floats.
func NewAnySchema() *AnySchema {
	return &AnySchema{}
}

// AnySchema is a wildcard allowing maps, lists, integers, strings, bools. and floats.
type AnySchema struct {
	Type TypeID `json:"type_id" yaml:"type_id"`
}

func (a *AnySchema) ReflectedType() reflect.Type {
	var defaultValue any
	return reflect.TypeOf(&defaultValue).Elem()
}

func (a *AnySchema) Unserialize(data any) (any, error) {
	return a.checkAndConvert(data)
}

func (a *AnySchema) Validate(data any) error {
	_, err := a.checkAndConvert(data)
	return err
}

func (a *AnySchema) Serialize(data any) (any, error) {
	return a.checkAndConvert(data)
}

func (a *AnySchema) ApplyScope(_ Scope) {

}

func (a *AnySchema) TypeID() TypeID {
	return TypeIDAny
}

//nolint:funlen
func (a *AnySchema) checkAndConvert(data any) (any, error) {
	t := reflect.ValueOf(data)
	switch t.Kind() {
	case reflect.Int:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return intInputMapper(data, nil)
	case reflect.Int64:
		return data.(int64), nil
	case reflect.Float32:
		return floatInputMapper(data, nil)
	case reflect.Float64:
		return asFloat(data)
	case reflect.String:
		return data.(string), nil
	case reflect.Bool:
		return asBool(data)
	case reflect.Slice:
		result := make([]any, t.Len())
		for i := 0; i < t.Len(); i++ {
			val, err := a.checkAndConvert(t.Index(i).Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%d]", i))
			}
			result[i] = val
		}
		return result, nil
	case reflect.Map:
		result := make(map[any]any, t.Len())
		for _, k := range t.MapKeys() {
			key, err := a.checkAndConvert(k.Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k))
			}
			v := t.MapIndex(k)
			value, err := a.checkAndConvert(v.Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", key))
			}
			result[key] = value
		}
		return result, nil
	default:
		return nil, &ConstraintError{
			Message: fmt.Sprintf("unsupported data type for 'any' type: %T", data),
		}
	}
}
