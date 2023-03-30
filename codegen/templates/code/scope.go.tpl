{{- range $id, $object := .Objects -}}
    {{ partial "code/type" $object }}
{{- end -}}