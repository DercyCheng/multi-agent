{{- define "llm-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "llm-service.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "llm-service.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
