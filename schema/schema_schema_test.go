package schema_test

import (
	_ "embed"
	"testing"

	"go.arcalot.io/assert"
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
	assert.NotNil(t, unserializedData.Steps())
	steps := unserializedData.Steps()
	assert.NotNil(t, steps["hello-world"])
	helloWorldStep := steps["hello-world"]
	assert.NotNil(t, helloWorldStep.Display())
	display := helloWorldStep.Display()
	assert.NotNil(t, display.Name())
	assert.Equals(t, *display.Name(), "Hello world!")

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
	assert.NotNil(t, unserializedData.Steps())
	steps := unserializedData.Steps()
	assert.NotNil(t, steps["hello-world"])
	helloWorldStep := steps["hello-world"]
	assert.NotNil(t, helloWorldStep.Display())
	display := helloWorldStep.Display()
	assert.NotNil(t, display.Name())
	assert.Equals(t, *display.Name(), "Hello world!")

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
	assert.NotNil(t, unserializedData.Steps())
	steps := unserializedData.Steps()
	assert.NotNil(t, steps["hello-world"])
	helloWorldStep := steps["hello-world"]
	assert.NotNil(t, helloWorldStep.Display())
	display := helloWorldStep.Display()
	assert.NotNil(t, display.Name())
	assert.Equals(t, *display.Name(), "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assert.NoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string])
	assert.Equals(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDScope)
}
