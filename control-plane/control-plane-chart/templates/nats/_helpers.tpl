{{/*
NATS server configuration.

We provide this explicitly because the subchart's auto-generated config
wraps values with Helm's `| quote`, which produces: token: "$AUTH_TOKEN"
NATS server does NOT expand env vars inside quoted strings.
This config uses unquoted env var references so NATS expands them at runtime.
Env var AUTH_TOKEN is injected via extraEnvVarsSecret.
*/}}
{{- define "control-plane.nats.configuration" -}}
server_name: $NATS_SERVER_NAME
listen: 0.0.0.0:{{ .Values.containerPorts.client }}
http: 0.0.0.0:{{ .Values.containerPorts.monitoring }}
# Authorization for client connections
authorization {
  token: $AUTH_TOKEN
  timeout: 1
}
# Logging options
debug: false
trace: false
logtime: false
# Pid file
pid_file: "/tmp/nats-server.pid"
# Clustering definition
cluster {
  name: "nats"
  listen: 0.0.0.0:{{ .Values.containerPorts.cluster }}
  # Routes are actively solicited and connected to from this server.
  {{- $fullname := include "common.names.fullname" . }}
  {{- $clusterPort := .Values.service.ports.cluster }}
  routes = [
    {{- range $podIndex := until (int .Values.replicaCount) }}
    nats://{{ $fullname }}-{{ $podIndex }}.{{ $fullname }}-headless:{{ $clusterPort }}
    {{- end }}
  ]
}
# JetStream configuration
jetstream: enabled
jetstream {
  store_dir: /data
  max_memory_store: {{ .Values.jetstream.maxMemory }}
  max_file_store: {{ .Values.persistence.size }}
}
{{- end -}}
