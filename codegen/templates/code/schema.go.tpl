{{- range $stepID, $step := .StepsValue -}}
{{- partial "code/type" $step.InputValue -}}
{{- range $outputID, $output := $step.Outputs -}}
{{- partial "code/type" $output.Schema -}}
{{- end -}}
{{- end -}}