{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.DisplayValue*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
{{- with . }}
schema.NewDisplayValue(
    {{ with .NameValue }}schema.PointerTo({{ . | escapeString}}){{ else }}nil{{end}},
    {{ with .DescriptionValue }}schema.PointerTo({{ . | escapeString}}){{ else }}nil{{end}},
    {{ with .IconValue }}schema.PointerTo({{ . | escapeString}}){{ else }}nil{{end}},
)
{{ else -}}
nil
{{- end -}}