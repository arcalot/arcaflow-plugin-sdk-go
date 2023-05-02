{{- /*gotype:go.flow.arcalot.io/pluginsdk/schema.UnitsDefinition*/ -}}
{{- import "go.flow.arcalot.io/pluginsdk/schema" -}}
schema.NewUnits(
    {{ prefix (partial "schema/unit" .BaseUnit) "    " }},
    map[int64]:*schema.UnitDefinition{
    {{- range $multiplier, $unit := .MultipliersValue -}}
        {{ $multiplier }}: {{ prefix (partial "schema/unit" $unit) "        " }}
    {{ end }}
    },
)