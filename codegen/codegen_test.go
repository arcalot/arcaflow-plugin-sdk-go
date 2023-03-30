package codegen_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"go.arcalot.io/assert"
	"go.flow.arcalot.io/pluginsdk/codegen"
	"go.flow.arcalot.io/pluginsdk/schema"
)

var testSchema = schema.NewSchema(
	map[string]*schema.StepSchema{
		"test1": schema.NewStepSchema(
			"test1",
			schema.NewScopeSchema(
				schema.NewObjectSchema(
					"input",
					map[string]*schema.PropertySchema{
						"name": schema.NewPropertySchema(
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
			),
			map[string]*schema.StepOutputSchema{},
			nil,
		),
	},
)

func TestCodeGeneration(t *testing.T) {
	generator := codegen.New()
	generated, err := generator.Generate(
		"testpackage",
		testSchema,
		map[string]codegen.ExternalReference{},
	)
	assert.NoError(t, err)
	code := string(generated)

	fs := token.NewFileSet()
	if _, err := parser.ParseFile(fs, "", code, parser.SkipObjectResolution); err != nil {
		t.Fatalf("failed to parse generated Go code (%v)\n%s", err, code)
	}

	if !strings.Contains(code, "package testpackage") {
		t.Fatalf("missing package header in code:\n%s", code)
	}
	if !strings.Contains(code, "type Input struct {") {
		t.Fatalf("missing input struct in code:\n%s", code)
	}
	t.Logf("Generated code:\n%s", code)
}
