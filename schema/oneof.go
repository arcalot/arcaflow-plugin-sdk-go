package schema

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
)

// OneOf is the root interface for one-of types. It should not be used directly but is provided for convenience.
type OneOf[KeyType int64 | string] interface {
	Type

	Types() map[KeyType]Object
	DiscriminatorFieldName() string
}

type OneOfSchema[KeyType int64 | string] struct {
	interfaceType               reflect.Type
	TypesValue                  map[KeyType]Object `json:"types"`
	DiscriminatorFieldNameValue string             `json:"discriminator_field_name"`
	// whether or not the discriminator is inlined in the underlying objects' schema
	DiscriminatorInlined bool `json:"discriminator_inlined"`
}

func (o OneOfSchema[KeyType]) TypeID() TypeID {
	var defaultValue KeyType
	switch any(defaultValue).(type) {
	case int64:
		return TypeIDOneOfInt
	case string:
		return TypeIDOneOfString
	default:
		panic(BadArgumentError{Message: fmt.Sprintf("Unexpected key type: %T", defaultValue)})
	}
}

func (o OneOfSchema[KeyType]) Types() map[KeyType]Object {
	return o.TypesValue
}

func (o OneOfSchema[KeyType]) DiscriminatorFieldName() string {
	return o.DiscriminatorFieldNameValue
}

func (o OneOfSchema[KeyType]) ApplyScope(scope Scope) {
	for _, t := range o.TypesValue {
		t.ApplyScope(scope)
	}
	// scope must be applied before we can access the subtypes' properties
	err := o.ValidateSubtypeDiscriminatorInlineFields()
	if err != nil {
		panic(err)
	}
}

func (o OneOfSchema[KeyType]) ReflectedType() reflect.Type {
	if o.interfaceType == nil {
		var defaultValue any
		return reflect.TypeOf(&defaultValue).Elem()
	}
	return o.interfaceType
}

//nolint:funlen
func (o OneOfSchema[KeyType]) UnserializeType(data any) (result any, err error) {
	reflectedValue := reflect.ValueOf(data)
	if reflectedValue.Kind() != reflect.Map {
		return result, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of type: '%s'. Expected map.",
				reflect.TypeOf(data).Name(),
			),
		}
	}

	discriminatorValue := reflectedValue.MapIndex(reflect.ValueOf(o.DiscriminatorFieldNameValue))
	if !discriminatorValue.IsValid() {
		return result, &ConstraintError{
			Message: fmt.Sprintf("Missing discriminator field '%s' in '%v'", o.DiscriminatorFieldNameValue, data),
		}
	}
	discriminator := discriminatorValue.Interface()
	typedDiscriminator, err := o.getTypedDiscriminator(discriminator)
	if err != nil {
		return result, err
	}
	typedData := make(map[string]any, reflectedValue.Len())
	for _, k := range reflectedValue.MapKeys() {
		v := reflectedValue.MapIndex(k)
		keyString, ok := k.Interface().(string)
		if !ok {
			return result, &ConstraintError{
				Message: fmt.Sprintf(
					"Invalid key type for one-of: '%T'",
					k.Interface(),
				),
			}
		}
		typedData[keyString] = v.Interface()
	}

	selectedType, ok := o.TypesValue[typedDiscriminator]
	if !ok {
		validDiscriminators := make([]string, len(o.TypesValue))
		i := 0
		for k := range o.TypesValue {
			validDiscriminators[i] = fmt.Sprintf("%v", k)
			i++
		}
		return result, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid value for '%s', expected one of: %s",
				o.DiscriminatorFieldNameValue,
				strings.Join(validDiscriminators, ", "),
			),
		}
	}

	cloneData := maps.Clone(typedData)
	if !o.DiscriminatorInlined {
		delete(cloneData, o.DiscriminatorFieldNameValue)
	}

	unserializedData, err := selectedType.Unserialize(cloneData)
	if err != nil {
		return result, err
	}

	unserializedMap, ok := unserializedData.(map[string]any)
	if ok {
		unserializedMap[o.DiscriminatorFieldNameValue] = discriminator
		return unserializedMap, nil
	}

	if o.interfaceType == nil {
		return unserializedData, nil
	}
	return saveConvertTo(unserializedData, o.interfaceType)
}

func (o OneOfSchema[KeyType]) ValidateType(data any) error {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		return err
	}
	if err := underlyingType.Validate(data); err != nil {
		return ConstraintErrorAddPathSegment(err, fmt.Sprintf("{oneof[%v]}", discriminatorValue))
	}
	return nil
}

func (o OneOfSchema[KeyType]) SerializeType(data any) (any, error) {
	discriminatorValue, underlyingType, err := o.findUnderlyingType(data)
	if err != nil {
		return nil, err
	}
	dataMap, ok := data.(map[string]any)
	if ok {
		cloneData := maps.Clone(dataMap)
		if !o.DiscriminatorInlined {
			delete(cloneData, o.DiscriminatorFieldNameValue)
		}
		data = cloneData
	}
	serializedData, err := underlyingType.Serialize(data)
	if err != nil {
		return nil, err
	}
	mapData := serializedData.(map[string]any)
	if _, ok := mapData[o.DiscriminatorFieldNameValue]; !ok {
		mapData[o.DiscriminatorFieldNameValue] = discriminatorValue
	}
	return mapData, nil
}

func (o OneOfSchema[KeyType]) Unserialize(data any) (any, error) {
	return o.UnserializeType(data)
}

func (o OneOfSchema[KeyType]) ValidateCompatibility(typeOrData any) error {
	// If a schema is given, validate that it's a oneof schema. If it isn't, fail.
	// If a schema is not given, validate as data.

	// Check if it's a map. If it is, verify it. If not, check if it's a schema, if it is, verify it.
	// If not, verify it as data.
	inputAsMap, ok := typeOrData.(map[string]any)
	if ok {
		_, _, err := o.validateMap(inputAsMap)
		return err
	}
	value := reflect.ValueOf(typeOrData)
	if reflect.Indirect(value).Kind() != reflect.Struct {
		// Validate as data
		return o.Validate(typeOrData)
	}

	inputAsIndirectInterface := reflect.Indirect(value).Interface()

	// Validate the oneof and key types
	schemaType, ok := inputAsIndirectInterface.(OneOfSchema[KeyType])
	if !ok {
		return &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for OneOfSchema. Found type (%T) does not match expected type (%T)",
				inputAsIndirectInterface, o),
		}
	}

	return o.validateSchema(schemaType)
}

func (o OneOfSchema[KeyType]) validateSchema(otherSchema OneOfSchema[KeyType]) error {
	// Validate that the discriminator fields match, and all other values match.

	// Validate the discriminator field name
	if otherSchema.DiscriminatorFieldName() != o.DiscriminatorFieldName() {
		return &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for OneOfSchema. Discriminator field name (%s) does not match expected field name (%s)",
				otherSchema.DiscriminatorFieldName(), o.DiscriminatorFieldName()),
		}
	}
	// Validate the key values and matching types
	for key, typeValue := range o.Types() {
		matchingTypeValue := otherSchema.Types()[key]
		if matchingTypeValue == nil {
			return &ConstraintError{
				Message: fmt.Sprintf(
					"validation failed for OneOfSchema. OneOf key '%v' is not present in given type", key),
			}
		}
		err := typeValue.ValidateCompatibility(matchingTypeValue)
		if err != nil {
			return &ConstraintError{
				Message: fmt.Sprintf(
					"validation failed for OneOfSchema. OneOf key '%v' does not have a compatible object schema (%s) ",
					key, err),
			}
		}
	}
	return nil
}

func (o OneOfSchema[KeyType]) validateMap(data map[string]any) (KeyType, Object, error) {
	var defaultKey KeyType
	// Validate that it has the discriminator field.
	// If it doesn't, fail
	// If it does, pass the non-discriminator fields into the ValidateCompatibility method for the object
	selectedTypeID := data[o.DiscriminatorFieldNameValue]
	if selectedTypeID == nil {
		return defaultKey, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for OneOfSchema. Discriminator field '%s' missing", o.DiscriminatorFieldNameValue),
		}
	}
	// Ensure it's the correct type
	selectedTypeIDAsserted, ok := selectedTypeID.(KeyType)
	if !ok {
		return defaultKey, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for OneOfSchema. Discriminator field '%v' has invalid type '%T'. Expected %T",
				o.DiscriminatorFieldNameValue, selectedTypeID, selectedTypeIDAsserted),
		}
	}
	// Find the object that's associated with the selected type
	selectedSchema := o.TypesValue[selectedTypeIDAsserted]
	if selectedSchema == nil {
		return defaultKey, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for OneOfSchema. Discriminator value '%v' is invalid. Expected one of: %v",
				selectedTypeIDAsserted, o.getTypeValues()),
		}
	}
	cloneData := maps.Clone(data)
	if selectedSchema.Properties()[o.DiscriminatorFieldNameValue] == nil { // Check to see if the discriminator is part of the sub-object.
		delete(cloneData, o.DiscriminatorFieldNameValue) // The discriminator isn't part of the object.
	}
	err := selectedSchema.ValidateCompatibility(cloneData)
	if err != nil {
		return defaultKey, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"validation failed for OneOfSchema. Failed to validate as selected schema type '%T' from discriminator value '%v' (%s)",
				selectedSchema, selectedTypeIDAsserted, err),
		}
	}
	return selectedTypeIDAsserted, selectedSchema, nil
}

func (o OneOfSchema[KeyType]) getTypeValues() []KeyType {
	output := make([]KeyType, len(o.TypesValue))
	i := 0
	for key := range o.TypesValue {
		output[i] = key
		i += 1
	}
	return output
}

func (o OneOfSchema[KeyType]) Validate(data any) error {
	if o.interfaceType == nil {
		return o.ValidateType(data)
	}
	d, err := saveConvertTo(data, o.interfaceType)
	if err != nil {
		return err
	}
	return o.ValidateType(d)
}

func (o OneOfSchema[KeyType]) Serialize(data any) (result any, err error) {
	if o.interfaceType == nil {
		return nil, o.ValidateType(data)
	}
	d, err := saveConvertTo(data, o.interfaceType)
	if err != nil {
		return nil, err
	}
	return o.SerializeType(d)
}

func (o OneOfSchema[KeyType]) getTypedDiscriminator(discriminator any) (KeyType, error) {
	var typedDiscriminator KeyType
	switch any(typedDiscriminator).(type) {
	case int64:
		intDiscriminator, err := intInputMapper(discriminator, nil)
		if err != nil {
			return typedDiscriminator, &ConstraintError{
				Message: fmt.Sprintf(
					"Invalid type %T for field %s, expected %T",
					discriminator,
					o.DiscriminatorFieldNameValue,
					typedDiscriminator,
				),
				Cause: err,
			}
		}
		typedDiscriminator = any(intDiscriminator).(KeyType)
	case string:
		stringDiscriminator, err := stringInputMapper(discriminator)
		if err != nil {
			return typedDiscriminator, &ConstraintError{
				Message: fmt.Sprintf(
					"Invalid type %T for field %s, expected %T",
					discriminator,
					o.DiscriminatorFieldNameValue,
					typedDiscriminator,
				),
				Cause: err,
			}
		}
		typedDiscriminator = any(stringDiscriminator).(KeyType)
	}
	return typedDiscriminator, nil
}

func (o OneOfSchema[KeyType]) findUnderlyingType(data any) (KeyType, Object, error) {
	var defaultValue KeyType
	reflectedType := reflect.TypeOf(data)
	if reflectedType.Kind() != reflect.Struct &&
		reflectedType.Kind() != reflect.Map &&
		(reflectedType.Kind() != reflect.Pointer || reflectedType.Elem().Kind() != reflect.Struct) {

		return defaultValue, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of type: '%s' expected struct or map.",
				reflect.TypeOf(data).Name(),
			),
		}
	}

	var foundKey *KeyType
	if reflectedType.Kind() == reflect.Map {
		myKey, mySchemaObj, err := o.validateMap(data.(map[string]any))
		if err != nil {
			return defaultValue, nil, err
		}
		return myKey, mySchemaObj, nil
	} else {
		for key, ref := range o.TypesValue {
			underlyingReflectedType := ref.ReflectedType()
			if underlyingReflectedType == reflectedType {
				keyValue := key
				foundKey = &keyValue
			}
		}
	}

	if foundKey == nil {
		dataType := reflect.TypeOf(data)
		values := make([]string, len(o.TypesValue))
		i := 0
		for _, ref := range o.TypesValue {
			values[i] = ref.ReflectedType().String()
			if values[i] == "" {
				panic(fmt.Errorf("bug: reflected type name is empty"))
			}
			i++
		}
		return defaultValue, nil, &ConstraintError{
			Message: fmt.Sprintf(
				"Invalid type for one-of schema: '%s' (valid types are: %s)",
				dataType.String(),
				strings.Join(values, ", "),
			),
		}
	}
	return *foundKey, o.TypesValue[*foundKey], nil
}

func (o OneOfSchema[KeyType]) ValidateSubtypeDiscriminatorInlineFields() error {
	for key, typeValue := range o.TypesValue {
		typeValueDiscriminatorValue, hasDiscriminator := typeValue.Properties()[o.DiscriminatorFieldNameValue]
		if !o.DiscriminatorInlined && hasDiscriminator {
			return fmt.Errorf(
				"object id %q has conflicting field %q; either remove that field or set inline to true for %T[%T]",
				typeValue.ID(), o.DiscriminatorFieldNameValue, o, key)
		} else if o.DiscriminatorInlined && !hasDiscriminator {
			return fmt.Errorf(
				"object id %q needs discriminator field %q; either add that field or set inline to false for %T[%T]",
				typeValue.ID(), o.DiscriminatorFieldNameValue, o, key)
		} else if o.DiscriminatorInlined && hasDiscriminator && (typeValueDiscriminatorValue.ReflectedType().Kind() != reflect.TypeOf(key).Kind()) {
			return fmt.Errorf(
				"the type of object id %v's discriminator field %q does not match OneOfSchema discriminator type; expected %v got %T",
				typeValue.ID(), o.DiscriminatorFieldNameValue, typeValueDiscriminatorValue.TypeID(), key)
		}
	}
	return nil
}
