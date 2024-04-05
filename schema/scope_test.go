package schema_test

import (
	"encoding/json"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type scopeTestObjectEmpty struct {
}

type scopeTestObjectB struct {
	C string `json:"c"`
}

type scopeTestObjectA struct {
	B scopeTestObjectB `json:"b"`
}

type scopeTestObjectAPtr struct {
	B *scopeTestObjectB `json:"b"`
}

var scopeTestObjectEmptySchema = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[scopeTestObjectEmpty](
		"scopeTestObjectEmpty",
		map[string]*schema.PropertySchema{},
	),
)
var scopeTestObjectEmptySchemaRenamed = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[scopeTestObjectEmpty](
		"scopeTestObjectEmptyRenamed",
		map[string]*schema.PropertySchema{},
	),
)

var scopeTestObjectCStrSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"scopeTestObjectC",
		map[string]*schema.PropertySchema{
			"d": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
)
var scopeTestObjectCIntSchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"scopeTestObjectC",
		map[string]*schema.PropertySchema{
			"d": schema.NewPropertySchema(
				schema.NewIntSchema(nil, nil, nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
)

var scopeTestObjectASchema = schema.NewScopeSchema(
	schema.NewObjectSchema(
		"scopeTestObjectA",
		map[string]*schema.PropertySchema{
			"b": schema.NewPropertySchema(
				schema.NewRefSchema("scopeTestObjectB", nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
	schema.NewObjectSchema(
		"scopeTestObjectB",
		map[string]*schema.PropertySchema{
			"c": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
)

var scopeTestObjectAType = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[scopeTestObjectA](
		"scopeTestObjectA",
		map[string]*schema.PropertySchema{
			"b": schema.NewPropertySchema(
				schema.NewRefSchema("scopeTestObjectB", nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
	schema.NewStructMappedObjectSchema[scopeTestObjectB](
		"scopeTestObjectB",
		map[string]*schema.PropertySchema{
			"c": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
)

var scopeTestObjectATypePtr = schema.NewScopeSchema(
	schema.NewStructMappedObjectSchema[*scopeTestObjectAPtr](
		"scopeTestObjectA",
		map[string]*schema.PropertySchema{
			"b": schema.NewPropertySchema(
				schema.NewRefSchema("scopeTestObjectB", nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
	schema.NewStructMappedObjectSchema[*scopeTestObjectB](
		"scopeTestObjectB",
		map[string]*schema.PropertySchema{
			"c": schema.NewPropertySchema(
				schema.NewStringSchema(nil, nil, nil),
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		},
	),
)

func TestScopeConstructor(t *testing.T) {
	assert.Equals(t, scopeTestObjectASchema.TypeID(), schema.TypeIDScope)
	assert.Equals(t, scopeTestObjectAType.TypeID(), schema.TypeIDScope)
}

func TestUnserialization(t *testing.T) {
	// Test unserialization of composition of two objects
	data := `{"b":{"c": "Hello world!"}}`
	var input any
	assert.NoError(t, json.Unmarshal([]byte(data), &input))

	result, err := scopeTestObjectAType.Unserialize(input)
	assert.NoError(t, err)
	assert.InstanceOf[scopeTestObjectA](t, result.(scopeTestObjectA))
	assert.Equals(t, result.(scopeTestObjectA).B.C, "Hello world!")
	serialized, err := scopeTestObjectAType.Serialize(result)
	assert.NoError(t, err)
	unserialized2, err := scopeTestObjectAType.Unserialize(serialized)
	assert.NoError(t, err)
	// test reversibility
	assert.Equals(t, unserialized2, result)

	// Now as a ptr
	resultPtr, err := scopeTestObjectATypePtr.Unserialize(input)
	assert.NoError(t, err)
	assert.InstanceOf[*scopeTestObjectAPtr](t, resultPtr.(*scopeTestObjectAPtr))
	assert.Equals(t, resultPtr.(*scopeTestObjectAPtr).B.C, "Hello world!")
	serialized, err = scopeTestObjectATypePtr.Serialize(resultPtr)
	assert.NoError(t, err)
	unserialized2, err = scopeTestObjectATypePtr.Unserialize(serialized)
	assert.NoError(t, err)
	// test reversiblity
	assert.Equals(t, unserialized2, resultPtr)

	// Test empty object
	data = `{}`
	assert.NoError(t, json.Unmarshal([]byte(data), &input))
	result, err = scopeTestObjectEmptySchema.Unserialize(input)
	assert.NoError(t, err)
	assert.InstanceOf[scopeTestObjectEmpty](t, result.(scopeTestObjectEmpty))
	serialized, err = scopeTestObjectEmptySchema.Serialize(result)
	assert.NoError(t, err)
	unserialized2, err = scopeTestObjectEmptySchema.Unserialize(serialized)
	assert.NoError(t, err)
	// test reversiblity
	assert.Equals(t, unserialized2, result)
}

func TestValidation(t *testing.T) {
	// Note: The scopeTestObject var used must be NewStructMappedObjectSchema,
	// or else it will be a dict instead of a struct, causing problems.
	// Test composition of two objects
	err := scopeTestObjectAType.Validate(scopeTestObjectA{
		scopeTestObjectB{
			"Hello world!",
		},
	})
	assert.NoError(t, err)

	// Test empty scope object
	err = scopeTestObjectEmptySchema.Validate(scopeTestObjectEmpty{})
	assert.NoError(t, err)
}
func TestCompatibilityValidationWithData(t *testing.T) {
	err := scopeTestObjectAType.ValidateCompatibility(map[string]any{
		"b": map[string]any{
			"c": "Hello world!",
		},
	})
	assert.NoError(t, err)

	// Replace the actual value with a schema
	err = scopeTestObjectAType.ValidateCompatibility(map[string]any{
		"b": map[string]any{
			"c": schema.NewStringSchema(nil, nil, nil),
		},
	})
	assert.NoError(t, err)

	// Test empty scope object
	// The ValidateCompatibility method should behave like Validate when data is passed in
	err = scopeTestObjectEmptySchema.ValidateCompatibility(map[string]any{})
	assert.NoError(t, err)
}

func TestCompatibilityValidationWithSchema(t *testing.T) {
	// Note: The scopeTestObject var used must be NewStructMappedObjectSchema,
	// or else it will be a dict instead of a struct, causing problems.
	// Test composition of two objects

	// Note: Doesn't support the non-pointer, dereferenced version of the scope type.
	err := scopeTestObjectAType.ValidateCompatibility(scopeTestObjectAType)
	assert.NoError(t, err)

	// Test empty scope object
	// Note: Doesn't support the non-pointer version.
	err = scopeTestObjectEmptySchema.ValidateCompatibility(scopeTestObjectEmptySchema)
	assert.NoError(t, err)

	// Now mismatched
	err = scopeTestObjectAType.ValidateCompatibility(scopeTestObjectEmptySchema)
	assert.Error(t, err)
	err = scopeTestObjectEmptySchema.ValidateCompatibility(scopeTestObjectAType)
	assert.Error(t, err)

	// Similar, but with a simple difference
	// Mismatching IDs
	err = scopeTestObjectEmptySchema.ValidateCompatibility(scopeTestObjectEmptySchemaRenamed)
	assert.Error(t, err)
	err = scopeTestObjectEmptySchemaRenamed.ValidateCompatibility(scopeTestObjectEmptySchema)
	assert.Error(t, err)
	// Mismatching type in one field, but with the field ID matching
	err = scopeTestObjectCStrSchema.ValidateCompatibility(scopeTestObjectCIntSchema)
	assert.Error(t, err)
	err = scopeTestObjectCIntSchema.ValidateCompatibility(scopeTestObjectCStrSchema)
	assert.Error(t, err)
}

func TestSerialization(t *testing.T) {
	serialized, err := scopeTestObjectAType.Serialize(scopeTestObjectA{
		scopeTestObjectB{
			"Hello world!",
		},
	})
	assert.NoError(t, err)
	assert.Equals(t, serialized.(map[string]any)["b"].(map[string]any)["c"].(string), "Hello world!")
	unserialized, err := scopeTestObjectAType.Unserialize(serialized)
	assert.NoError(t, err)
	serialized2, err := scopeTestObjectAType.Serialize(unserialized)
	assert.NoError(t, err)

	// test reversiblity
	assert.Equals(t, serialized2, serialized)
}

func TestSelfSerialization(t *testing.T) {
	serializedScope, err := scopeTestObjectAType.SelfSerialize()
	assert.NoError(t, err)
	serializedScopeMap := serializedScope.(map[string]any)
	if serializedScopeMap["root"] != "scopeTestObjectA" {
		t.Fatalf("Unexpected root object: %s", serializedScopeMap["root"])
	}
}

//nolint:funlen
func TestApplyingExternalNamespace(t *testing.T) {
	// This test tests applying a scope to a schema that contains scopes,
	// properties, objects, maps, and lists.
	// The applied scope must be passed down to all of those types, validating
	// that the scope gets applied down and that errors are propagated up.
	refRefSchema := schema.NewNamespacedRefSchema("scopeTestObjectB", "test-namespace", nil)

	refProperty := schema.NewPropertySchema(
		refRefSchema,
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	withObjectRefSchema := schema.NewNamespacedRefSchema("scopeTestObjectB", "test-namespace", nil)
	nestedObjectProperty := schema.NewPropertySchema(
		schema.NewObjectSchema("level-2-object", map[string]*schema.PropertySchema{
			"a": schema.NewPropertySchema(
				withObjectRefSchema,
				nil,
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
		}),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	withListRefSchema := schema.NewNamespacedRefSchema("scopeTestObjectB", "test-namespace", nil)
	listProperty := schema.NewPropertySchema(
		schema.NewListSchema(
			withListRefSchema,
			nil,
			nil,
		),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	withMapRefSchema := schema.NewNamespacedRefSchema("scopeTestObjectB", "test-namespace", nil)
	mapProperty := schema.NewPropertySchema(
		schema.NewMapSchema(
			schema.NewIntSchema(nil, nil, nil),
			withMapRefSchema,
			nil,
			nil,
		),
		nil,
		true,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
	testScopes := map[string]struct {
		scope *schema.ScopeSchema
		ref   schema.Ref
	}{
		"withRef": {
			schema.NewScopeSchema(
				schema.NewObjectSchema(
					"scopeTestObjectA",
					map[string]*schema.PropertySchema{
						"ref-b": refProperty,
					},
				),
			),
			refRefSchema,
		},
		"withObjectInObject": {
			schema.NewScopeSchema(
				schema.NewObjectSchema(
					"scopeTestObjectA",
					map[string]*schema.PropertySchema{
						"ref-b": nestedObjectProperty,
					},
				),
			),
			withObjectRefSchema,
		},
		"withList": {
			schema.NewScopeSchema(
				schema.NewObjectSchema(
					"scopeTestObjectA",
					map[string]*schema.PropertySchema{
						"list-type": listProperty,
					},
				),
			),
			withListRefSchema,
		},
		"withMap": {
			schema.NewScopeSchema(
				schema.NewObjectSchema(
					"scopeTestObjectA",
					map[string]*schema.PropertySchema{
						"map-type": mapProperty,
					},
				),
			),
			withMapRefSchema,
		},
	}

	var externalScope = schema.NewScopeSchema(
		schema.NewObjectSchema(
			"scopeTestObjectB",
			map[string]*schema.PropertySchema{
				"c": schema.NewPropertySchema(
					schema.NewStringSchema(nil, nil, nil),
					nil,
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
			},
		),
	)
	for testName, td := range testScopes {
		testData := td
		t.Run(testName, func(t *testing.T) {
			// Not applied yet
			// Outermost
			err := testData.scope.ValidateReferences()
			assert.Error(t, err)
			// Innermost
			err = testData.ref.ValidateReferences()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "missing its link")
			assert.Equals(t, testData.ref.ObjectReady(), false)
			testData.scope.ApplyScope(externalScope, "test-namespace")
			assert.Equals(t, testData.ref.ObjectReady(), true)
			// Now it's applied, so the error should be resolved.
			// Outermost
			assert.NoError(t, testData.scope.ValidateReferences())
			// Innermost
			assert.NoError(t, testData.ref.ValidateReferences())
		})
	}
}

func TestApplyingExternalNamespaceToNonRefTypes(t *testing.T) {
	testCases := map[string]schema.Type{
		"int":      schema.NewIntSchema(nil, nil, nil),
		"float":    schema.NewFloatSchema(nil, nil, nil),
		"bool":     schema.NewBoolSchema(),
		"string":   schema.NewStringSchema(nil, nil, nil),
		"any":      schema.NewAnySchema(),
		"str_enum": schema.NewStringEnumSchema(map[string]*schema.DisplayValue{}),
		"int_enum": schema.NewIntEnumSchema(map[int64]*schema.DisplayValue{}, nil),
		"pattern":  schema.NewPatternSchema(),
	}
	for testName, tt := range testCases {
		testType := tt
		t.Run(testName, func(t *testing.T) {
			// Not applied yet
			// Should be no error because none of the given types can contain references.
			assert.NoError(t, testType.ValidateReferences())
			testType.ApplyScope(scopeTestObjectEmptySchema, "test-namespace")
			// Should still be no error.
			assert.NoError(t, testType.ValidateReferences())
		})
	}
}
