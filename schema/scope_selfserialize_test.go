package schema_test

import (
	"go.arcalot.io/assert"
	"regexp"
	"testing"

	"go.flow.arcalot.io/pluginsdk/schema"
)

type Step struct {
	Needs []string `json:"needs"`
	Input any      `json:"input"`
}

type PluginStep struct {
	Step `json:",inline"`

	Plugin string `json:"plugin"`
	Deploy any    `json:"deploy"`
}

type Workflow struct {
	Input  *schema.ScopeSchema `json:"input"`
	Steps  map[string]Step     `json:"steps"`
	Output any                 `json:"output"`
}

func getPluginStepSchema() schema.Object { //nolint:funlen
	return schema.NewStructMappedObjectSchema[PluginStep](
		"PluginStep",
		map[string]*schema.PropertySchema{
			"needs": schema.NewPropertySchema(
				schema.NewListSchema(
					schema.NewStringSchema(schema.IntPointer(1), nil, nil),
					nil,
					nil,
				),
				schema.NewDisplayValue(
					schema.PointerTo("Needs"),
					schema.PointerTo("A list of expressions that should be evaluated for step dependencies."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				schema.PointerTo("[]"),
				nil,
			),
			"input": schema.NewPropertySchema(
				// We use an any schema here because it will need to evaluate expressions before applying the schema.
				schema.NewAnySchema(),
				schema.NewDisplayValue(
					schema.PointerTo("Input"),
					schema.PointerTo("Input data for this step."),
					nil,
				),
				false,
				nil,
				nil,
				nil,
				schema.PointerTo("[]"),
				nil,
			),
			"plugin": schema.NewPropertySchema(
				schema.NewStringSchema(schema.IntPointer(1), schema.IntPointer(255), nil),
				schema.NewDisplayValue(
					schema.PointerTo("Plugin"),
					schema.PointerTo("The plugin container image in fully qualified form."),
					nil,
				),
				true,
				nil,
				nil,
				nil,
				nil,
				nil,
			),
			"deploy": schema.NewPropertySchema(
				schema.NewAnySchema(),
				schema.NewDisplayValue(
					schema.PointerTo("Deployment"),
					schema.PointerTo("Deployment configuration for this plugin."),
					nil,
				),
				true,
				[]string{"plugin"},
				nil,
				nil,
				schema.PointerTo(`{"type": "docker"}`),
				nil,
			),
		},
	)
}

func getWorkflowSchema() *schema.TypedScopeSchema[*Workflow] {
	return schema.NewTypedScopeSchema[*Workflow](
		schema.NewStructMappedObjectSchema[*Workflow](
			"Workflow",
			map[string]*schema.PropertySchema{
				"input": schema.NewPropertySchema(
					schema.DescribeScope(),
					schema.NewDisplayValue(
						schema.PointerTo("Input"),
						schema.PointerTo("Input definitions for this workflow. These are used to render the form for starting the workflow."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"steps": schema.NewPropertySchema(
					schema.NewMapSchema(
						schema.NewStringSchema(
							schema.IntPointer(1),
							schema.IntPointer(255),
							regexp.MustCompile("^[$@a-zA-Z0-9-_]+$"),
						),
						getPluginStepSchema(),
						schema.IntPointer(1),
						nil,
					),
					schema.NewDisplayValue(
						schema.PointerTo("Steps"),
						schema.PointerTo("Workflow steps to execute in this workflow."),
						nil,
					),
					true,
					nil,
					nil,
					nil,
					nil,
					nil,
				),
				"output": schema.NewPropertySchema(
					schema.NewAnySchema(),
					schema.NewDisplayValue(
						schema.PointerTo("Output"),
						schema.PointerTo("Output data structure with expressions to pull in output data from steps."),
						nil,
					),
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
}

func TestSelfSerializationEndToEnd(t *testing.T) {
	s := getWorkflowSchema()
	_, err := s.SelfSerialize()
	assert.NoError(t, err)
}
