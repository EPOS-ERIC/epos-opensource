{{- define "epos.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "epos.name" -}}
{{- if .Values.nameOverride -}}
{{- .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "epos.labels" -}}
app.kubernetes.io/name: {{ include "epos.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.Version }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}
{{- end -}}

{{- define "epos.selectorLabels" -}}
app.kubernetes.io/name: {{ include "epos.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "epos.imagePullSecrets" -}}
{{- if .Values.image_pull_secrets.enabled }}
imagePullSecrets:
  - name: epos-registry-secret
{{- end }}
{{- end -}}

{{- define "epos.nodeSelector" -}}
{{- if .Values.nodeSelector }}
nodeSelector:
  {{- .Values.nodeSelector | toYaml | nindent 2 }}
{{- end }}
{{- end -}}

{{- define "epos.tolerations" -}}
{{- if .Values.tolerations }}
tolerations:
  {{- .Values.tolerations | toYaml | nindent 2 }}
{{- end }}
{{- end -}}

{{- define "epos.affinity" -}}
{{- if .Values.affinity }}
affinity:
  {{- .Values.affinity | toYaml | nindent 2 }}
{{- end }}
{{- end -}}

{{- define "epos.securityContext" -}}
{{- if .Values.securityContext }}
securityContext:
  {{- .Values.securityContext | toYaml | nindent 2 }}
{{- end }}
{{- end -}}

{{- define "epos.postgresqlConnectionString" -}}
jdbc:postgresql://{{ default .Values.components.metadata_database.host }}:{{ default .Values.components.metadata_database.port 5432 }}/{{ .Values.components.metadata_database.db_name }}?user={{ .Values.components.metadata_database.user }}&password={{ .Values.components.metadata_database.password }}
{{- end -}}
