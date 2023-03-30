{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.SchemaSchema*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewCallableSchema({{- range $stepID, $step := .StepsValue }}
    {{ prefix (partial "schema/step" $step) "    " }},
{{ end -}}
)