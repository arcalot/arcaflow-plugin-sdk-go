package schema_test

import (
	_ "embed"
	"go.arcalot.io/assert"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/hello_world_plugin.yaml
var helloWorldPluginSchema []byte

func TestSchemaUnserializationHelloWorld(t *testing.T) {
	data := map[string]any{}
	assert.NoError(t, yaml.Unmarshal(helloWorldPluginSchema, &data))
	unserializedData, err := schema.UnserializeSchema(data)
	assert.NoError(t, err)
	steps := assert.NotNilR(t, unserializedData.Steps())
	helloWorldStep := assert.NotNilR(t, steps["hello-world"])
	display := assert.NotNilR(t, helloWorldStep.Display())
	name := assert.NotNilR(t, display.Name())
	assert.Equals(t, *name, "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assert.NoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string])
	assert.Equals(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDRef)
}

//go:embed testdata/embedded_objects.yaml
var embeddedSchema []byte

func TestSchemaUnserializationEmbeddedObjects(t *testing.T) {
	data := map[string]any{}
	assert.NoError(t, yaml.Unmarshal(embeddedSchema, &data))
	unserializedData, err := schema.UnserializeSchema(data)
	assert.NoError(t, err)
	steps := assert.NotNilR(t, unserializedData.Steps())
	helloWorldStep := assert.NotNilR(t, steps["hello-world"])
	display := assert.NotNilR(t, helloWorldStep.Display())
	name := assert.NotNilR(t, display.Name())
	assert.Equals(t, *name, "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assert.NoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string])
	assert.Equals(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDObject)
}

//go:embed testdata/super_scoped.yaml
var superScopedSchema []byte

func TestSchemaUnserializationSuperScoped(t *testing.T) {
	data := map[string]any{}
	assert.NoError(t, yaml.Unmarshal(superScopedSchema, &data))
	unserializedData, err := schema.UnserializeSchema(data)
	assert.NoError(t, err)
	steps := assert.NotNilR(t, unserializedData.Steps())
	helloWorldStep := assert.NotNilR(t, steps["hello-world"])
	display := assert.NotNilR(t, helloWorldStep.Display())
	name := assert.NotNilR(t, display.Name())
	assert.Equals(t, *name, "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assert.NoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string])
	assert.Equals(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDScope)
}

func TestStepOutputSchema(t *testing.T) {
	stepOutputSchema := schema.DescribeStepOutput()
	unserializedStepOutput, err := stepOutputSchema.Unserialize(map[string]any{
		"schema": map[string]any{
			"root": "A",
			"objects": map[string]any{
				"A": map[string]any{
					"id": "A",
					"properties": map[string]any{
						"foo": map[string]any{
							"type": map[string]any{
								"type_id": "string",
							},
							"required": true,
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	unserializedStepOutputOutput, err := unserializedStepOutput.(*schema.StepOutputSchema).Unserialize(map[string]any{
		"foo": "bar",
	})
	if err != nil {
		t.Fatalf("%v", err)
	}
	if unserializedStepOutputOutput.(map[string]any)["foo"] != "bar" {
		t.Fatalf("Incorrect unserialized output: %s", unserializedStepOutputOutput.(map[any]any)["foo"])
	}
}
