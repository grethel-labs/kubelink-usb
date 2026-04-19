{{- define "kubelink-usb.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubelink-usb.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name (include "kubelink-usb.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "kubelink-usb.controllerServiceAccountName" -}}
{{- if .Values.serviceAccount.controller.name -}}
{{- .Values.serviceAccount.controller.name -}}
{{- else -}}
{{- printf "%s-controller" (include "kubelink-usb.fullname" .) -}}
{{- end -}}
{{- end -}}

{{- define "kubelink-usb.agentServiceAccountName" -}}
{{- if .Values.serviceAccount.agent.name -}}
{{- .Values.serviceAccount.agent.name -}}
{{- else -}}
{{- printf "%s-agent" (include "kubelink-usb.fullname" .) -}}
{{- end -}}
{{- end -}}
