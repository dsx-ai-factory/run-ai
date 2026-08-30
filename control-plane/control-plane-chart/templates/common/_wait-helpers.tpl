{{/*
Generic init container that waits for a TCP service to be reachable.
Input: dict with ctx (Helm root context), containerName, svcHost, svcPort, svcDisplayName
*/}}
{{- define "_awaitingServiceReadiness" -}}
{{- $waitImage := .ctx.Values.global.waitImage -}}
- name: {{ .containerName }}
  image: {{ printf "%s/%s:%s" $waitImage.registry $waitImage.repository ($waitImage.tag | toString) }}
  command: ['bash', '-c', 'until (exec 3<>/dev/tcp/{{ .svcHost }}/{{ .svcPort }}) 2>/dev/null; do echo "Waiting for {{ .svcDisplayName }} service"; sleep 2; done;']
  imagePullPolicy: {{ $waitImage.pullPolicy }}
  resources:
    limits:
      cpu: 100m
      memory: 100Mi
    requests:
      cpu: 50m
      memory: 50Mi
  securityContext:
    allowPrivilegeEscalation: false
    runAsNonRoot: true
    readOnlyRootFilesystem: true
    {{- if ne .ctx.Values.global.config.kubernetesDistribution "openshift" }}
    runAsUser: 65532
    {{- end }}
    seccompProfile:
      type: RuntimeDefault
    capabilities:
      drop:
      - ALL
{{- end -}}

{{/* Notification Service */}}
{{- define "awaitingNotificationServiceReadiness" -}}
{{- include "awaitingNotificationServiceReadinessFromMap" (first .) -}}
{{- end -}}

{{- define "awaitingNotificationServiceReadinessFromMap" -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" .
  "containerName" "check-notification-service-readiness"
  "svcHost" (printf "%s-%s" .Values.fullnameOverride .Values.notificationsService.name)
  "svcPort" (.Values.notificationsService.service.configPort | toString)
  "svcDisplayName" .Values.notificationsService.name) -}}
{{- end -}}

{{/* Temp Backend Settings */}}
{{- define "awaitingTempBackendSettingsReadiness" -}}
{{- $top := first . -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" $top
  "containerName" "check-temp-backend-settings-readiness"
  "svcHost" (printf "%s-%s" $top.Values.fullnameOverride $top.Values.tempBackendSettings.name)
  "svcPort" ($top.Values.tempBackendSettings.service.port | toString)
  "svcDisplayName" $top.Values.tempBackendSettings.name) -}}
{{- end -}}

{{/* Identity Manager */}}
{{- define "awaitingIdentityManagerReadiness" -}}
{{- include "awaitingIdentityManagerReadinessFromMap" (first .) -}}
{{- end -}}

{{- define "awaitingIdentityManagerReadinessFromMap" -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" .
  "containerName" "check-identity-manager-readiness"
  "svcHost" (printf "%s-%s" .Values.fullnameOverride .Values.identityManager.name)
  "svcPort" (.Values.identityManager.service.port | toString)
  "svcDisplayName" .Values.identityManager.name) -}}
{{- end -}}

{{/* Authorization Service */}}
{{- define "awaitingAuthorizationServiceReadiness" -}}
{{- include "awaitingAuthorizationServiceReadinessFromMap" (first .) -}}
{{- end -}}

{{- define "awaitingAuthorizationServiceReadinessFromMap" -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" .
  "containerName" "check-authorization-service-readiness"
  "svcHost" (printf "%s-%s" .Values.fullnameOverride .Values.authorization.name)
  "svcPort" (.Values.authorization.service.port | toString)
  "svcDisplayName" .Values.authorization.name) -}}
{{- end -}}

{{/* Metrics Service */}}
{{- define "awaitingMetricsServiceReadiness" -}}
{{- $top := first . -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" $top
  "containerName" "check-metrics-service-readiness"
  "svcHost" (printf "%s-%s" $top.Values.fullnameOverride $top.Values.metricsService.name)
  "svcPort" ($top.Values.metricsService.service.port | toString)
  "svcDisplayName" $top.Values.metricsService.name) -}}
{{- end -}}

{{/* Tenants Manager */}}
{{- define "awaitingTenantsManagerReadiness" -}}
{{- $top := first . -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" $top
  "containerName" "check-tenants-manager-readiness"
  "svcHost" (printf "%s-%s" $top.Values.fullnameOverride $top.Values.tenantsManager.name)
  "svcPort" ($top.Values.tenantsManager.service.port | toString)
  "svcDisplayName" $top.Values.tenantsManager.name) -}}
{{- end -}}

{{/* Org Unit Service */}}
{{- define "awaitingOrgUnitServiceReadiness" -}}
{{- $top := first . -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" $top
  "containerName" "check-org-unit-service-readiness"
  "svcHost" (printf "%s-%s" $top.Values.fullnameOverride $top.Values.orgUnitService.name)
  "svcPort" ($top.Values.orgUnitService.service.port | toString)
  "svcDisplayName" $top.Values.orgUnitService.name) -}}
{{- end -}}

{{/* Cluster Service */}}
{{- define "awaitingClusterServiceReadiness" -}}
{{- $top := first . -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" $top
  "containerName" "check-cluster-service-readiness"
  "svcHost" (printf "%s-%s" $top.Values.fullnameOverride $top.Values.clusterService.name)
  "svcPort" ($top.Values.clusterService.service.port | toString)
  "svcDisplayName" $top.Values.clusterService.name) -}}
{{- end -}}

{{/* Workloads Service */}}
{{- define "awaitingWorkloadsServiceReadiness" -}}
{{- $top := first . -}}
{{- include "_awaitingServiceReadiness" (dict
  "ctx" $top
  "containerName" "check-workloads-service-readiness"
  "svcHost" (printf "%s-%s" $top.Values.fullnameOverride $top.Values.workloads.name)
  "svcPort" ($top.Values.workloads.service.port | toString)
  "svcDisplayName" $top.Values.workloads.name) -}}
{{- end -}}

{{/* Postgres — kept separate due to psql-specific logic, volume mounts, and env vars */}}
{{- define "awaitingPostgresReadinessContainer" -}}
{{- $top := first . -}}
{{- include "_awaitingPostgresReadinessContainer" (dict "ctx" $top) -}}
{{- end -}}

{{- define "awaitingPostgresReadinessContainerFromMap" -}}
{{- include "_awaitingPostgresReadinessContainer" (dict "ctx" .) -}}
{{- end -}}

{{- define "_awaitingPostgresReadinessContainer" -}}
{{- $top := .ctx }}
{{- $pgAuth := $top.Values.global.postgresql.auth }}
- name: check-postgres-readiness
  image: docker.io/postgres:16-alpine
  env:
    - name: PGSSLROOTCERT
      value: /etc/ssl/certs/runai-ca.pem
    - name: PG_CONNECTION_STRING
      value: {{ printf "postgres://%s@%s:%.0f/%s?sslmode=%s" $pgAuth.username $pgAuth.host $pgAuth.port $pgAuth.database $pgAuth.dbSslMode }}
  command: ['sh',
            '-c',
            'trap "exit 0" TERM; export PGPASSWORD="$(cat /etc/secrets/DB_PASSWORD)" && until psql "$PG_CONNECTION_STRING" -c "SELECT 1" -q --no-align --tuples-only 2>/dev/null; do echo "Waiting for {{ $pgAuth.host }} service"; sleep 2; done; echo "Successfully connected to the PostgreSQL"']
  imagePullPolicy: IfNotPresent
  volumeMounts:
  {{- if ($top.Values.global.customCA).enabled | default false }}
    {{- $top.Values.global.customCA.volumeMounts | toYaml | nindent 4 }}
  {{- end }}
    - mountPath: /etc/secrets
      name: db-password
  resources:
    limits:
      cpu: 100m
      memory: 100Mi
    requests:
      cpu: 50m
      memory: 50Mi
  securityContext:
    allowPrivilegeEscalation: false
    runAsNonRoot: true
    {{- /* postgres:16-alpine defaults to root, so vanilla k8s needs an explicit
    non-root uid; OpenShift restricted-v2 rejects uid 70 and injects a
    namespace-range uid when none is set. */}}
    {{- if ne $top.Values.global.config.kubernetesDistribution "openshift" }}
    runAsUser: 70
    {{- end }}
    seccompProfile:
      type: RuntimeDefault
    capabilities:
      drop:
      - ALL
{{- end -}}
