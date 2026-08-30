{{- define "common.pod.template.tpl" -}}
{{- $top := first . }}
{{- $pod := index . 1 }}
{{- $serviceAccount := index . 2 }}
{{- $obj := index . 3 }}
{{- $additionalAnnotations := index . 4 }}
metadata:
  {{- with merge (default dict $obj.annotations) $additionalAnnotations }}
  annotations:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  labels:
    {{- include "common.selectorLabels" $top | nindent 4 }}
    {{- include "common.selectorLabelName" $obj | nindent 4 }}
    {{- include "common.labels" $top | nindent 4 }}
  {{- with $pod.podLabels }}
    {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  {{- with $pod.imagePullSecrets | default $top.Values.global.imagePullSecrets}}
  imagePullSecrets:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  serviceAccountName: {{ include "common.serviceAccountName" (list $top $serviceAccount) }}
  securityContext:
    {{- include "common.securityContext.render" (list $top $pod.podSecurityContext) | nindent 4 }}
  containers:
  - {{- include "common.container" (list $top $pod) | nindent 4 }}
  {{- with $pod.nodeSelector }}
  nodeSelector:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with $pod.restartPolicy }}
  restartPolicy:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  affinity:
    nodeAffinity:
      {{- with (($pod.affinity).nodeAffinity | default $top.Values.global.affinity.nodeAffinity) }}
        {{- toYaml . | nindent 6 }}
      {{- end }}
    podAffinity:
      {{- with (($pod.affinity).podAffinity | default $top.Values.global.affinity.podAffinity) }}
        {{- toYaml . | nindent 6 }}
      {{- end }}
    podAntiAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        {{- if $top.Values.global.requireDefaultPodAntiAffinity }}
        - labelSelector:
            matchLabels:
              {{- include "common.selectorLabelName" $obj | nindent 14 }}
          topologyKey: "kubernetes.io/hostname"
        {{- end }}
        {{- with $pod.affinity }}
        {{- with .podAntiAffinity }}
        {{- with .requiredDuringSchedulingIgnoredDuringExecution }}
          {{- toYaml . | nindent 8 }}
        {{- end }}
        {{- end }}
        {{- end }}
        {{- with $top.Values.global.affinity }}
        {{- with .podAntiAffinity }}
        {{- with .requiredDuringSchedulingIgnoredDuringExecution}}
          {{- toYaml . | nindent 8 }}
        {{- end }}
        {{- end }}
        {{- end }}
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchLabels:
                {{- include "common.selectorLabelName" $obj | nindent 16 }}
            topologyKey: "kubernetes.io/hostname"
        {{- with $pod.affinity }}
        {{- with .podAntiAffinity }}
        {{- with .preferredDuringSchedulingIgnoredDuringExecution }}
          {{- toYaml . | nindent 8 }}
        {{- end }}
        {{- end }}
        {{- end }}
        {{- with $top.Values.global.affinity }}
        {{- with .podAntiAffinity }}
        {{- with .preferredDuringSchedulingIgnoredDuringExecution }}
          {{- toYaml . | nindent 8 }}
        {{- end }}
        {{- end }}
        {{- end }}
  {{- if or $pod.tolerations $top.Values.global.tolerations }}
  tolerations:
  {{- with $pod.tolerations }}
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with $top.Values.global.tolerations }}
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- end }}
  {{- with $pod.priorityClassName }}
  priorityClassName: {{ . }}
  {{- end }}
  volumes:
  {{- if ($top.Values.global.customCA).enabled | default false }}
    {{- $top.Values.global.customCA.volumes | toYaml | nindent 4 }}
  {{- end }}
  {{- with $pod.volumes }}
    {{- range . }}
    {{- if .projected }}
    - name: {{ .name }}
      # We have an hidden assumption here that a secret with no items is a service secret and it's should be versioned
      # according to the argoRollout logic. If it has items we assume it's an infra secret
      # like redis or postgres and it's name shouldn't be versioned
      projected:
        sources:
          {{- range .projected.sources }}
          {{- if .secret }}
          - secret:
              {{- if .secret.items }}
              name: {{ tpl .secret.name $top  }}
              items:
                {{- range .secret.items }}
                - key: {{ tpl .key $top }}
                  path: {{ .path }}
                {{- end }}
              {{- else }}
              {{- if and (($top.Values.global.config).argoRollouts) (ne $pod.argoRollouts false) }}
              name: {{ tpl .secret.name $top }}-{{ $top.Chart.Version }}
              {{- else }}
              name: {{ tpl .secret.name $top  }}
              {{- end }}
              {{- end }}
              {{- if hasKey .secret "optional" }}
              optional: {{ .secret.optional }}
              {{- end }}
          {{- end }}
          {{- end }}
    {{- end }}
    {{- if .secret }}
    - name: {{ .name }}
      secret:
        {{- if and (($top.Values.global.config).argoRollouts) (ne $pod.argoRollouts false) }}
        secretName: {{ tpl .secret.secretName $top }}-{{ $top.Chart.Version }}
        {{- else }}
        secretName: {{ tpl .secret.secretName $top  }}
        {{- end }}
      {{- else if .configMap }}
    - name: {{ .name }}
      configMap:
        {{- if and (($top.Values.global.config).argoRollouts) (ne $pod.argoRollouts false) }}
        name: {{ .configMap.name }}-{{ $top.Chart.Version }}
        {{- else }}
        name: {{ .configMap.name }}
        {{- end }}
      {{- end }}
    {{- if hasKey . "emptyDir" }}
    - name: {{ .name }}
      emptyDir: {}
    {{- end }}
  {{- end }}
  {{- end }}
{{- end }}

{{- define "common.pod.template" -}}
{{- include "common.utils.merge" (append . "common.pod.template.tpl") }}
{{- end }}
