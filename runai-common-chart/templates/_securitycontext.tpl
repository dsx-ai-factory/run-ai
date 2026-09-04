{{/*
Render a securityContext block, adapting it for OpenShift.

On OpenShift the namespace-assigned SCC (restricted-v2) injects a uid/gid/
fsGroup from the namespace's allowed range and REJECTS pods that request an
explicit runAsUser/runAsGroup/fsGroup outside that range - the ReplicaSet then
never creates a pod (e.g. traefik hardcoding runAsUser
65532). So on OpenShift we drop those three fields and let the platform assign
valid ids. On every other distribution the securityContext is emitted
unchanged, so rendered output is byte-identical to before.

Keyed on the same signal used elsewhere in the chart:
global.config.kubernetesDistribution == "openshift".

Usage:
  securityContext:
    {{- include "common.securityContext.render" (list $top $securityContext) | nindent 4 }}
*/}}
{{- define "common.securityContext.render" -}}
{{- $top := index . 0 -}}
{{- $sc := index . 1 -}}
{{- if eq ((($top.Values).global).config).kubernetesDistribution "openshift" -}}
{{- $sc = omit ($sc | default dict) "runAsUser" "runAsGroup" "fsGroup" -}}
{{- end -}}
{{- toYaml $sc -}}
{{- end -}}
