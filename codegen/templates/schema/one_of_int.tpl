{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.OneOfSchema[int64]*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewOneOfIntSchema(
    map[int64]schema.Object{{ "{" }}{{ range $key, $value := .TypesValue }}}
        {{ $key }}: {{ prefix (partial "schema/type" $value) "        " }},
        {{- end -}}
    },
    {{ .DiscriminatorFieldNameValue | escapeString }}
)