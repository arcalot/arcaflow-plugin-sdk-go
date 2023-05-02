{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.UnitDefinition*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewUnit(
    {{ .NameShortSingularValue | escapeString }},
    {{ .NameShortPluralValue | escapeString }},
    {{ .NameLongSingularValue | escapeString }},
    {{ .NameLongPluralValue | escapeString }},
)