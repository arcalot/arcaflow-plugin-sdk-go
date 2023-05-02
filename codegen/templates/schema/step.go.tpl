{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.Step*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewCallableStep[{{ .Input.ID | structName }}](
    {{ .ID | escapeString }},
    {{ prefix (partial "schema/type" .InputValue) "    " }},
    map[string]*schema.StepOutputSchema{
        {{- range $outputID, $outputSchema := .OutputsValue }}
        {{ $outputID | escapeString }}: {{ prefix (partial "schema/output" $outputSchema) "        " }},
        {{ end -}}
    },
    {{ prefix (partial "schema/display" .DisplayValue) "    " }},
    {{ .ID | handlerFuncName }},
)