{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.ScopeSchema*/ -}}
{{- $root := .RootValue -}}
schema.NewScopeSchema(
    {{ prefix (partial "schema/type" (index .Objects $root)) "    " }},
{{ range $id, $object := .Objects -}}{{ if ne $id $root }}
    {{ prefix (partial "schema/type" $object) "    " }},
{{- end -}}{{- end -}}
)