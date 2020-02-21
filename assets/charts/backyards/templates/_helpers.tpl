{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "backyards.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "prometheus.name" -}}
{{- printf "%s-prometheus" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubestatemetrics.name" -}}
{{- printf "%s-kubestatemetrics" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubeletdiscovery.name" -}}
{{- printf "%s-kubelet-discovery" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "grafana.name" -}}
{{- printf "%s-grafana" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tracing.name" -}}
{{- printf "%s-tracing" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "auditsink.name" -}}
{{- printf "%s-auditsink" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ingressgateway.name" -}}
{{- printf "%s-ingressgateway" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "als.name" -}}
{{- printf "%s-als" (include "backyards.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "backyards.fullname" -}}
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

{{/*
Create a default fully qualified app name for the Prometheus component
*/}}
{{- define "prometheus.fullname" -}}
{{- printf "%s-prometheus" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the kube-state-metrics component
*/}}
{{- define "kubestatemetrics.fullname" -}}
{{- printf "%s-kubestatemetrics" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the kubelet-discovery component
*/}}
{{- define "kubeletdiscovery.fullname" -}}
{{- printf "%s-kubelet-discovery" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the Grafana component
*/}}
{{- define "grafana.fullname" -}}
{{- printf "%s-grafana" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the tracing component
*/}}
{{- define "tracing.fullname" -}}
{{- printf "%s-tracing" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the AuditSink component
*/}}
{{- define "auditsink.fullname" -}}
{{- printf "%s-auditsink" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the Ingress gateway component
*/}}
{{- define "ingressgateway.fullname" -}}
{{- printf "%s-ingressgateway" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name for the Ingress gateway component
*/}}
{{- define "als.fullname" -}}
{{- printf "%s-als" (include "backyards.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "backyards.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Call nested templates.
Source: https://stackoverflow.com/a/52024583/3027614
*/}}
{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end -}}


