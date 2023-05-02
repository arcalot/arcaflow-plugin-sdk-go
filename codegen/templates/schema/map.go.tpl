{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.MapSchema[any,any]*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewMapSchema(
    {{ prefix (partial "schema/type" .KeysValue) "    " }},
    {{ prefix (partial "schema/type" .ValuesValue) "    " }},
    {{with .MinValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
    {{with .MaxValue}}schema.PointerTo({{.}}){{ else }}nil{{end}},
)