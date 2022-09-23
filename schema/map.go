package schema

import (
	"fmt"
	"reflect"
)

// Map holds the schema definition for key-value associations. This dataclass only has the ability to hold the
// configuration but cannot serialize, unserialize or validate. For that functionality please use MapType.
type Map[KeyType Type, ValueType Type] interface {
	Type

	Keys() KeyType
	Values() ValueType
	Min() *int64
	Max() *int64
}

// UntypedMap is a map schema without specific underlying types.
type UntypedMap = Map[Type, Type]

// TypedMap is a map schema that can be unserialized in its underlying components.
type TypedMap[KeyType comparable, ValueType any] interface {
	TypedType[map[KeyType]ValueType]
	Map[TypedType[KeyType], TypedType[ValueType]]
}

// NewMapSchema creates a new map schema.
func NewMapSchema(keys Type, values Type, min *int64, max *int64) *MapSchema[Type, Type] {
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

	return &MapSchema[Type, Type]{
		keys,
		values,
		min,
		max,
	}
}

// MapSchema is the implementation of tye map types.
type MapSchema[K Type, V Type] struct {
	KeysValue   K      `json:"keys"`
	ValuesValue V      `json:"values"`
	MinValue    *int64 `json:"min"`
	MaxValue    *int64 `json:"max"`
}

func (m MapSchema[K, V]) TypeID() TypeID {
	return TypeIDMap
}

func (m MapSchema[K, V]) ReflectedType() reflect.Type {
	return reflect.MapOf(m.KeysValue.ReflectedType(), m.ValuesValue.ReflectedType())
}

func (m MapSchema[K, V]) Keys() K {
	return m.KeysValue
}

func (m MapSchema[K, V]) Values() V {
	return m.ValuesValue
}

func (m MapSchema[K, V]) Min() *int64 {
	return m.MinValue
}

func (m MapSchema[K, V]) Max() *int64 {
	return m.MaxValue
}

func (m MapSchema[K, V]) ApplyScope(scope Scope) {
	m.KeysValue.ApplyScope(scope)
	m.ValuesValue.ApplyScope(scope)
}

func (m MapSchema[K, V]) Unserialize(data any) (any, error) {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Map {
		return nil, &ConstraintError{
			Message: fmt.Sprintf("Must be a map, %T given", data),
		}
	}

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

	t := m.ReflectedType()
	result := reflect.MakeMapWithSize(t, v.Len())
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
		result.SetMapIndex(reflect.ValueOf(unserializedKey), reflect.ValueOf(unserializedValue))
	}
	return result.Interface(), nil
}

func (m MapSchema[K, V]) Validate(data any) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Map {
		return &ConstraintError{
			Message: fmt.Sprintf("Must be a map, %T given", data),
		}
	}

	if m.MinValue != nil && *m.MinValue > int64(v.Len()) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at least %d items, %d given", *m.MinValue, v.Len()),
		}
	}
	if m.MaxValue != nil && *m.MaxValue < int64(v.Len()) {
		return &ConstraintError{
			Message: fmt.Sprintf("Must have at most %d items, %d given", *m.MaxValue, v.Len()),
		}
	}

	for _, k := range v.MapKeys() {
		if err := m.KeysValue.Validate(k.Interface()); err != nil {
			return ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k))
		}
		if err := m.ValuesValue.Validate(v.MapIndex(k).Interface()); err != nil {
			return ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", k))
		}
	}
	return nil
}

func (m MapSchema[K, V]) Serialize(data any) (any, error) {
	if err := m.Validate(data); err != nil {
		return nil, err
	}

	v := reflect.ValueOf(data)
	result := make(map[any]any, v.Len())
	for _, k := range v.MapKeys() {
		serializedKey, err := m.KeysValue.Serialize(k.Interface())
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("{%v}", k))
		}
		serializedValue, err := m.ValuesValue.Serialize(v.MapIndex(k).Interface())
		if err != nil {
			return nil, ConstraintErrorAddPathSegment(err, fmt.Sprintf("[%v]", k))
		}
		result[serializedKey] = serializedValue
	}
	return result, nil
}

// NewTypedMapSchema creates a new map schema with a defined underlying type.
func NewTypedMapSchema[KeyType comparable, ValueType any](
	keys TypedType[KeyType],
	values TypedType[ValueType],
	min *int64,
	max *int64,
) *TypedMapSchema[KeyType, ValueType] {
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

	return &TypedMapSchema[KeyType, ValueType]{
		MapSchema[TypedType[KeyType], TypedType[ValueType]]{
			keys,
			values,
			min,
			max,
		},
	}
}

type TypedMapSchema[KeyType comparable, ValueType any] struct {
	MapSchema[TypedType[KeyType], TypedType[ValueType]]
}

func (m TypedMapSchema[KeyType, ValueType]) UnserializeType(data any) (result map[KeyType]ValueType, err error) {
	unserialized, err := m.Unserialize(data)
	if err != nil {
		return result, err
	}
	return unserialized.(map[KeyType]ValueType), nil
}

func (m TypedMapSchema[KeyType, ValueType]) ValidateType(data map[KeyType]ValueType) error {
	return m.Validate(data)
}

func (m TypedMapSchema[KeyType, ValueType]) SerializeType(data map[KeyType]ValueType) (any, error) {
	return m.Serialize(data)
}
