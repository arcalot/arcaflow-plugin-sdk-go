package schema_test

import (
	_ "embed"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
	"gopkg.in/yaml.v3"
)

//go:embed testdata/hello_world_plugin.yaml
var helloWorldPluginSchema []byte

func TestSchemaUnserializationHelloWorld(t *testing.T) {
	data := map[string]any{}
	assertNoError(t, yaml.Unmarshal(helloWorldPluginSchema, &data))
	unserializedData, err := schema.UnserializeSchema(data)
	assertNoError(t, err)
	steps := assertNotNil(t, unserializedData.Steps())
	helloWorldStep := assertNotNil(t, steps["hello-world"])
	display := assertNotNil(t, helloWorldStep.Display())
	name := assertNotNil(t, display.Name())
	assertEqual(t, *name, "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assertNoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string, schema.Object])
	assertEqual(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDRef)
}

//go:embed testdata/embedded_objects.yaml
var embeddedSchema []byte

func TestSchemaUnserializationEmbeddedObjects(t *testing.T) {
	data := map[string]any{}
	assertNoError(t, yaml.Unmarshal(embeddedSchema, &data))
	unserializedData, err := schema.UnserializeSchema(data)
	assertNoError(t, err)
	steps := assertNotNil(t, unserializedData.Steps())
	helloWorldStep := assertNotNil(t, steps["hello-world"])
	display := assertNotNil(t, helloWorldStep.Display())
	name := assertNotNil(t, display.Name())
	assertEqual(t, *name, "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assertNoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string, schema.Object])
	assertEqual(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDObject)
}

//go:embed testdata/super_scoped.yaml
var superScopedSchema []byte

func TestSchemaUnserializationSuperScoped(t *testing.T) {
	data := map[string]any{}
	assertNoError(t, yaml.Unmarshal(superScopedSchema, &data))
	unserializedData, err := schema.UnserializeSchema(data)
	assertNoError(t, err)
	steps := assertNotNil(t, unserializedData.Steps())
	helloWorldStep := assertNotNil(t, steps["hello-world"])
	display := assertNotNil(t, helloWorldStep.Display())
	name := assertNotNil(t, display.Name())
	assertEqual(t, *name, "Hello world!")

	_, err = unserializedData.SelfSerialize()
	assertNoError(t, err)

	nameType := unserializedData.StepsValue["hello-world"].InputValue.Objects()["InputParams"].Properties()["name"].Type().(*schema.OneOfSchema[string, schema.Object])
	assertEqual(t, nameType.Types()["fullname"].TypeID(), schema.TypeIDScope)
}
