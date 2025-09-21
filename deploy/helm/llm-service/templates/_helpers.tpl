{{/*
Expand the name of the chart.
*/}}
{{- define "llm-service.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "llm-service.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "llm-service.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "llm-service.labels" -}}
helm.sh/chart: {{ include "llm-service.chart" . }}
{{ include "llm-service.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: multi-agent-platform
{{- end }}

{{/*
Selector labels
*/}}
{{- define "llm-service.selectorLabels" -}}
app.kubernetes.io/name: {{ include "llm-service.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "llm-service.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "llm-service.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Generate database connection string
*/}}
{{- define "llm-service.databaseUrl" -}}
{{- if .Values.database.enabled }}
{{- printf "%s://%s:%s@%s:%d/%s" .Values.database.type .Values.database.username .Values.database.password .Values.database.host (.Values.database.port | int) .Values.database.database }}
{{- end }}
{{- end }}

{{/*
Generate Redis connection string
*/}}
{{- define "llm-service.redisUrl" -}}
{{- if .Values.redis.enabled }}
{{- if .Values.redis.auth.enabled }}
{{- printf "redis://:%s@%s-redis-master:%d" .Values.redis.auth.password (include "llm-service.fullname" .) (6379 | int) }}
{{- else }}
{{- printf "redis://%s-redis-master:%d" (include "llm-service.fullname" .) (6379 | int) }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Generate Jaeger configuration
*/}}
{{- define "llm-service.jaegerConfig" -}}
{{- if .Values.monitoring.jaeger.enabled }}
{{- printf "%s:%d" .Values.monitoring.jaeger.agent.host (.Values.monitoring.jaeger.agent.port | int) }}
{{- end }}
{{- end }}

{{/*
Create image pull secret
*/}}
{{- define "llm-service.imagePullSecrets" -}}
{{- if .Values.image.imagePullSecrets }}
imagePullSecrets:
{{- range .Values.image.imagePullSecrets }}
  - name: {{ . }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Generate environment-specific configuration
*/}}
{{- define "llm-service.envConfig" -}}
{{- if eq .Values.global.environment "production" }}
RUST_LOG: "info"
PYTHONOPTIMIZE: "1"
{{- else if eq .Values.global.environment "development" }}
RUST_LOG: "debug"
PYTHONDONTWRITEBYTECODE: "1"
{{- else }}
RUST_LOG: "warn"
{{- end }}
{{- end }}

{{/*
Validate required values
*/}}
{{- define "llm-service.validateValues" -}}
{{- if and .Values.database.enabled (not .Values.database.password) }}
{{- fail "Database password is required when database is enabled" }}
{{- end }}
{{- if and .Values.redis.enabled .Values.redis.auth.enabled (not .Values.redis.auth.password) }}
{{- fail "Redis password is required when Redis auth is enabled" }}
{{- end }}
{{- if and .Values.autoscaling.enabled (le (.Values.autoscaling.minReplicas | int) 0) }}
{{- fail "Autoscaling minReplicas must be greater than 0" }}
{{- end }}
{{- if and .Values.autoscaling.enabled (le (.Values.autoscaling.maxReplicas | int) (.Values.autoscaling.minReplicas | int)) }}
{{- fail "Autoscaling maxReplicas must be greater than minReplicas" }}
{{- end }}
{{- end }}
