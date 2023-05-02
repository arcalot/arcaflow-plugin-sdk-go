{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.OneOfSchema[string]*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewOneOfStringSchema(
    map[string]schema.Object{{ "{" }}{{ range $key, $value := .TypesValue }}}
        {{ $key | escapeString }}: {{ prefix (partial "schema/type" $value) "        " }},
        {{- end -}}
    },
    {{ .DiscriminatorFieldNameValue | escapeString }}
)