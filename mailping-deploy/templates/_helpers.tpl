{{/*
Expand the name of the chart.
*/}}
{{- define "mailping-deploy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "mailping-deploy.fullname" -}}
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
{{- define "mailping-deploy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "mailping-deploy.labels" -}}
helm.sh/chart: {{ include "mailping-deploy.chart" . }}
{{ include "mailping-deploy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "mailping-deploy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mailping-deploy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "mailping-deploy.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "mailping-deploy.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the PostgreSQL database URL
*/}}
{{- define "mailping-deploy.databaseUrl" -}}
{{- if .Values.postgresql.external.enabled }}
{{- printf "postgres://%s:%s@%s:%s/%s?sslmode=disable" .Values.postgresql.external.username .Values.postgresql.external.password .Values.postgresql.external.host (.Values.postgresql.external.port | toString) .Values.postgresql.external.database }}
{{- else }}
{{- printf "postgres://%s:%s@%s-postgresql.%s.svc.cluster.local:5432/%s?sslmode=disable" .Values.postgresql.auth.username .Values.postgresql.auth.password .Release.Name .Release.Namespace .Values.postgresql.auth.database }}
{{- end }}
{{- end }}