package schema

import (
	"fmt"
	"reflect"
)

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

	return &abstractMapSchema[AbstractSchema, AbstractSchema]{
		keys,
		values,
		min,
		max,
	}
}

type abstractMapSchema[K AbstractSchema, V AbstractSchema] struct {
	KeysValue   K      `json:"keys"`
	ValuesValue V      `json:"values"`
	MinValue    *int64 `json:"min"`
	MaxValue    *int64 `json:"max"`
}

//nolint:unused
type mapSchema struct {
	abstractMapSchema[AbstractSchema, AbstractSchema] `json:",inline"`
}

func (m abstractMapSchema[K, V]) TypeID() TypeID {
	return TypeIDMap
}

func (m abstractMapSchema[K, V]) Keys() AbstractSchema {
	return m.KeysValue
}

func (m abstractMapSchema[K, V]) Values() AbstractSchema {
	return m.ValuesValue
}

func (m abstractMapSchema[K, V]) Min() *int64 {
	return m.MinValue
}

func (m abstractMapSchema[K, V]) Max() *int64 {
	return m.MaxValue
}

// MapType is a serializable version of MapSchema.
type MapType[K ~int64 | ~string, V any] interface {
	MapSchema
	AbstractType[map[K]V]

	TypedKeys() AbstractType[K]
	TypedValues() AbstractType[V]
}

// NewMapType defines a serializable version of MapSchema.
func NewMapType[K ~int64 | ~string, V any](
	keys AbstractType[K],
	values AbstractType[V],
	min *int64, max *int64,
) MapType[K, V] {
	return &mapType[K, V]{
		abstractMapSchema[AbstractType[K], AbstractType[V]]{
			KeysValue:   keys,
			ValuesValue: values,
			MinValue:    min,
			MaxValue:    max,
		},
	}
}

type mapType[K ~int64 | ~string, V any] struct {
	abstractMapSchema[AbstractType[K], AbstractType[V]] `json:",inline"`
}

func (m mapType[K, V]) ApplyScope(s ScopeSchema[PropertyType, ObjectType[any]]) {
	m.KeysValue.ApplyScope(s)
	m.ValuesValue.ApplyScope(s)
}

func (m mapType[K, V]) UnderlyingType() map[K]V {
	return map[K]V{}
}

func (m mapType[K, V]) Unserialize(data any) (map[K]V, error) {
	var result map[K]V
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Map:
		if m.MinValue != nil && *m.MinValue > int64(v.Len()) {
			return nil, &ConstraintError{
				Message: fmt.Sprintf("Must have at least %d items, %d given", *m.MinValue, v.Len()),
			}
		}
		if m.MaxValue != nil && *m.MaxValue < int64(v.Len()) {
			return nil, &ConstraintError{
				Message: fmt.Sprintf("Must have at most %d items, %d given", *m.MaxValue, v.Len()),
			}
		}

		result = make(map[K]V, v.Len())
		for _, k := range v.MapKeys() {
			val := v.MapIndex(k)

			unserializedKey, err := m.KeysValue.Unserialize(k.Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k.Interface()))
			}
			unserializedValue, err := m.ValuesValue.Unserialize(val.Interface())
			if err != nil {
				return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", k.Interface()))
			}
			result[unserializedKey] = unserializedValue
		}
		return result, nil
	default:
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must be a map, %T given", data),
		}
	}
}

func (m mapType[K, V]) Validate(data map[K]V) error {
	if m.MinValue != nil && *m.MinValue > int64(len(data)) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at least %d items, %d given", *m.MinValue, len(data)),
		}
	}
	if m.MaxValue != nil && *m.MaxValue < int64(len(data)) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at most %d items, %d given", *m.MaxValue, len(data)),
		}
	}

	for k, v := range data {
		if err := m.KeysValue.Validate(k); err != nil {
			return ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k))
		}
		if err := m.ValuesValue.Validate(v); err != nil {
			return ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", k))
		}
	}
	return nil
}

func (m mapType[K, V]) Serialize(data map[K]V) (any, error) {
	if m.MinValue != nil && *m.MinValue > int64(len(data)) {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must have at least %d items, %d given", *m.MinValue, len(data)),
		}
	}
	if m.MaxValue != nil && *m.MaxValue < int64(len(data)) {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must have at most %d items, %d given", *m.MaxValue, len(data)),
		}
	}

	result := make(map[any]any, len(data))
	for k, v := range data {
		serializedKey, err := m.KeysValue.Serialize(k)
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k))
		}
		serializedValue, err := m.ValuesValue.Serialize(v)
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", k))
		}
		result[serializedKey] = serializedValue
	}
	return result, nil
}

func (m mapType[K, V]) TypedKeys() AbstractType[K] {
	return m.KeysValue
}

func (m mapType[K, V]) TypedValues() AbstractType[V] {
	return m.ValuesValue
}
