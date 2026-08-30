{{- /*
  Helper templates for the secret-generator hook.
  Each component defines a named template that returns its secret
  configurations as JSON ({"items": [...]}).
*/ -}}

{{- /* ============================================ */ -}}
{{- /*        COMPONENT SECRET CONFIGURATIONS      */ -}}
{{- /* ============================================ */ -}}

{{- /*
  PostgreSQL secret configuration.
  Always generated -- even when bundled postgres is disabled (postgresql.enabled=false),
  external postgres credentials are passed via global.postgresql.auth.* and must be
  stored in the generated secret so that other components can consume them.

  When bundled postgres is disabled, password keys are marked as "externalOnly" so that the
  secret generator refuses to fall back to random generation -- random passwords would
  not match the external database. The generator will error at runtime if no explicit
  value is provided (via chartValue, existingSecret, or legacy secret migration).
*/ -}}
{{- define "secret-generator.postgres-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}

{{- $genName := printf "%s-gen-postgresql" $prefix }}
{{- $existingSecret := .Values.global.postgresql.auth.existingSecret }}

{{- $userPasswordKey := .Values.global.postgresql.auth.secretKeys.userPasswordKey | default "password" }}
{{- $adminPasswordKey := .Values.global.postgresql.auth.secretKeys.adminPasswordKey | default "postgres-password" }}

{{- $userKey := dict "name" $userPasswordKey "type" "password" "length" 24 }}
{{- if .Values.global.postgresql.auth.password }}
  {{- $_ := set $userKey "chartValue" .Values.global.postgresql.auth.password }}
{{- end }}
{{- if not .Values.postgresql.enabled }}
  {{- $_ := set $userKey "externalOnly" true }}
{{- end }}

{{- $adminKey := dict "name" $adminPasswordKey "type" "password" "length" 24 }}
{{- if .Values.global.postgresql.auth.postgresPassword }}
  {{- $_ := set $adminKey "chartValue" .Values.global.postgresql.auth.postgresPassword }}
{{- end }}
{{- if not .Values.postgresql.enabled }}
  {{- $_ := set $adminKey "externalOnly" true }}
{{- end }}

{{- $postgresKeys := list $userKey $adminKey }}
{{- $postgresDict := dict
  "name" $genName
  "legacySecretName" "runai-backend-postgresql"
  "keys" $postgresKeys
}}

{{- if $existingSecret }}
  {{- $_ := set $postgresDict "existingSecretName" $existingSecret }}
{{- end }}

{{- $configs = append $configs $postgresDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}

{{- /*
  Keycloak secret configuration.
  Includes DB config and admin password.
*/ -}}
{{- define "secret-generator.keycloak-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}

{{- $kcAdminPwKey := dict "name" "KEYCLOAK_ADMIN_PASSWORD" "type" "password" "length" 24 }}
{{- if .Values.keycloakx.adminPassword }}
  {{- $_ := set $kcAdminPwKey "chartValue" .Values.keycloakx.adminPassword }}
{{- end }}

{{- $genKcName := printf "%s-gen-keycloakx" $prefix }}
{{- $keycloakDict := dict
  "name" $genKcName
  "legacySecretName" "runai-backend-keycloakx"
  "keys" (list
    $kcAdminPwKey
  )
}}
{{- if .Values.keycloakx.existingSecret }}
  {{- $_ := set $keycloakDict "existingSecretName" .Values.keycloakx.existingSecret }}
{{- end }}

{{- $configs = append $configs $keycloakDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}



{{- /*
  Tenants Manager ADMIN_PASSWORD configuration.
  Reconciles ADMIN_PASSWORD into runai-backend-gen-tenants-manager.
  This is the primary spec for the Secret: it carries legacySecretName so that
  upgrades migrate the existing ADMIN_PASSWORD from runai-backend-tenants-manager
  instead of rotating to a fresh random value.

  When tenantsManager.config.existingSecret is set, existingSecretName defers
  this spec entirely — consistent with the postgres/keycloak pattern.
  Complexity validation for adminPassword chartValue lives in
  templates/tenants-manager/secret.yaml.
*/ -}}
{{- define "secret-generator.tenants-manager-admin-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}

{{- $adminKey := dict
  "name" "ADMIN_PASSWORD"
  "type" "password"
  "length" 24
  "chartValue" (.Values.tenantsManager.config.adminPassword | default "")
}}

{{- $adminDict := dict
  "name" (printf "%s-gen-tenants-manager" $prefix)
  "legacySecretName" (printf "%s-tenants-manager" $prefix)
  "keys" (list $adminKey)
}}

{{- if .Values.tenantsManager.config.existingSecret }}
  {{- $_ := set $adminDict "existingSecretName" .Values.tenantsManager.config.existingSecret }}
{{- end }}

{{- $configs = append $configs $adminDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Tenants Manager TENANT_MNG_PASSWORD configuration.
  Reconciles TENANT_MNG_PASSWORD into the same runai-backend-gen-tenants-manager
  Secret as tenants-manager-admin-configs (mirrors the cache/pubsub split on
  gen-cache-pubsub). No legacySecretName: legacy migration for this key is
  covered by the admin spec above.

  When global.management.existingSecret is set, existingSecretName defers
  this spec entirely.
*/ -}}
{{- define "secret-generator.tenants-manager-mgmt-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}

{{- $mgmtPwKey := dict
  "name" "TENANT_MNG_PASSWORD"
  "type" "password"
  "length" 24
  "chartValue" (.Values.global.management.password | default "")
}}

{{- $mgmtDict := dict
  "name" (printf "%s-gen-tenants-manager" $prefix)
  "keys" (list $mgmtPwKey)
}}

{{- if .Values.global.management.existingSecret }}
  {{- $_ := set $mgmtDict "existingSecretName" .Values.global.management.existingSecret }}
{{- end }}

{{- $configs = append $configs $mgmtDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  NATS AUTH_TOKEN configuration.
  Source of truth for the NATS auth token; cache and pubsub passwords
  alias it via copyFrom in their own sections (see secret-generator.cache-configs
  and secret-generator.pubsub-auth-configs).

    - nats.enabled=true  -> 32-char alphanumeric. Legacy nats.conf migration
      (extracts the token from runai-backend-nats) is handled by the generator.
    - nats.enabled=false -> static empty string. Suits no-auth external
      cache/pubsub topologies; authenticated externals should override via
      global.cache.auth.existingSecret / global.pubsub.auth.existingSecret.

  When nats.extraEnvVarsSecret is overridden to a user-managed Secret, this
  spec defers entirely (existingSecretName). The user's Secret must contain
  AUTH_TOKEN, CACHE_PASSWORD, and PUBSUB_PASSWORD (the cache and pubsub
  sections cascade to the same secret unless their own existingSecret is
  set).
*/ -}}
{{- define "secret-generator.nats-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}
{{- $genName := printf "%s-gen-cache-pubsub" $prefix }}

{{- $authTokenKey := dict "name" "AUTH_TOKEN" "type" "alphanumeric" "length" 32 }}
{{- if not .Values.nats.enabled }}
  {{- $authTokenKey = dict "name" "AUTH_TOKEN" "type" "static" "chartValue" "" }}
{{- end }}

{{- $natsDict := dict
  "name" $genName
  "legacySecretName" "runai-backend-nats"
  "existingSecretName" (.Values.nats.extraEnvVarsSecret | default "")
  "keys" (list $authTokenKey)
}}
{{- $configs = append $configs $natsDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Cache CACHE_PASSWORD configuration.
  Targets the same generated Secret as nats-configs (gen-cache-pubsub) and
  reconciles a single key alongside AUTH_TOKEN.

  Resolution (per the secret-generator's chartValue -> copyFrom chain):
    1. chartValue from global.cache.auth.password when non-empty (wins).
    2. copyFrom AUTH_TOKEN when nats.enabled=true (default).
    3. static empty when nats.enabled=false.

  existingSecretName cascade:
    1. global.cache.auth.existingSecret (cache-specific override; consumers
       route through control-plane.cache.password.secret.* to the same Secret,
       so the generator skips adding CACHE_PASSWORD here).
    2. nats.extraEnvVarsSecret when set to a user-managed Secret (cache reads
       AUTH_TOKEN/CACHE_PASSWORD from the same user-owned Secret as NATS).
    3. otherwise unset; generator reconciles CACHE_PASSWORD into gen-cache-pubsub.
*/ -}}
{{- define "secret-generator.cache-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}
{{- $genName := printf "%s-gen-cache-pubsub" $prefix }}

{{- $cachePassword := .Values.global.cache.auth.password | default "" }}
{{- $cachePasswordKey := .Values.global.cache.auth.secretKeys.passwordKey | default "cache-password" }}
{{- $cachePwKey := dict "name" $cachePasswordKey "copyFrom" "AUTH_TOKEN" "chartValue" $cachePassword }}
{{- if not .Values.nats.enabled }}
  {{- $cachePwKey = dict "name" $cachePasswordKey "type" "static" "chartValue" $cachePassword }}
{{- end }}

{{- $cacheDict := dict
  "name" $genName
  "existingSecretName" (.Values.global.cache.auth.existingSecret | default "")
  "keys" (list $cachePwKey)
}}
{{- $configs = append $configs $cacheDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Pubsub PUBSUB_PASSWORD configuration.
  Mirrors secret-generator.cache-configs, keyed off global.pubsub.auth.*
  instead of global.cache.auth.*.
*/ -}}
{{- define "secret-generator.pubsub-auth-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}
{{- $genName := printf "%s-gen-cache-pubsub" $prefix }}

{{- $pubsubPassword := .Values.global.pubsub.auth.password | default "" }}
{{- $pubsubPasswordKey := .Values.global.pubsub.auth.secretKeys.passwordKey | default "pubsub-password" }}
{{- $pubsubUserKey := .Values.global.pubsub.auth.secretKeys.userKey | default "pubsub-user" }}
{{- $pubsubUser := .Values.global.pubsub.auth.user }}
{{- $pubsubPwKey := dict "name" $pubsubPasswordKey "copyFrom" "AUTH_TOKEN" "chartValue" $pubsubPassword }}
{{- if not .Values.nats.enabled }}
  {{- $pubsubPwKey = dict "name" $pubsubPasswordKey "type" "static" "chartValue" $pubsubPassword }}
{{- end }}

{{- $pubsubDict := dict
  "name" $genName
  "existingSecretName" (.Values.global.pubsub.auth.existingSecret | default "")
  "keys" (list
    $pubsubPwKey
    (dict "name" $pubsubUserKey "type" "static" "chartValue" $pubsubUser)
  )
}}
{{- $configs = append $configs $pubsubDict }}

{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Notifications Proxy DB encryption key configuration.
  Generates a 32-char alphanumeric key once; migrates from the legacy
  Helm-managed secret if it exists.
*/ -}}
{{- define "secret-generator.notifications-proxy-encryption-key-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}

{{- $encKey := dict "name" "DB_ENCRYPTION_KEY" "type" "alphanumeric" "length" 32 }}
{{- if .Values.notificationsProxy.config.db.encryptionKey }}
  {{- $_ := set $encKey "chartValue" .Values.notificationsProxy.config.db.encryptionKey }}
{{- end }}

{{- $configs = append $configs (dict
  "name" (printf "%s-gen-notifications-proxy-encryption-key" $prefix)
  "legacySecretName" (printf "%s-notifications-proxy-encryption-key" $prefix)
  "keys" (list $encKey)
) }}

{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Telemetry service (formerly diagnostics-service) CUSTOMER_ID configuration.
  CUSTOMER_ID is generated once (24-char alphanumeric) and preserved across
  upgrades and ArgoCD syncs. Migrates the existing CUSTOMER_ID from the
  legacy Helm-managed runai-backend-diagnostics-service Secret (where it
  coexisted with DB_USER). DB_USER stays in that legacy Secret as a
  chart-managed config value; only CUSTOMER_ID moves here.

  Both Secret names deliberately keep the pre-rename "diagnostics-service"
  spelling: they carry the tenant's existing CUSTOMER_ID, and renaming them
  would generate a new identity on upgrade.
*/ -}}
{{- define "secret-generator.telemetry-service-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}
{{- if eq (include "control-plane.telemetry.enabled" .) "true" }}
{{- $customerIdKey := dict "name" "CUSTOMER_ID" "type" "alphanumeric" "length" 24 }}
{{- $configs = append $configs (dict
  "name" (printf "%s-gen-diagnostics-service" $prefix)
  "legacySecretName" (printf "%s-diagnostics-service" $prefix)
  "keys" (list $customerIdKey)
) }}
{{- end }}
{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Telemetry service secret rename: gen-diagnostics-service -> gen-telemetry-service.
  Migrates CUSTOMER_ID into the new telemetry-named Secret once
  secret-generator.telemetry-service-configs above has ensured
  runai-backend-gen-diagnostics-service exists (preserved, migrated, or freshly
  created) -- this entry always has a source to migrate from, regardless of
  whether the upgrade started pre-generator or already on gen-diagnostics-service.

  The old runai-backend-gen-diagnostics-service Secret is intentionally left in
  place (not deleted): the generator has no hop beyond a single legacySecretName,
  so it remains the durable source of truth for CUSTOMER_ID.
*/ -}}
{{- define "secret-generator.telemetry-service-rename-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}
{{- if eq (include "control-plane.telemetry.enabled" .) "true" }}
{{- $customerIdKey := dict "name" "CUSTOMER_ID" "type" "alphanumeric" "length" 24 }}
{{- $configs = append $configs (dict
  "name" (printf "%s-gen-telemetry-service" $prefix)
  "legacySecretName" (printf "%s-gen-diagnostics-service" $prefix)
  "keys" (list $customerIdKey)
) }}
{{- end }}
{{- dict "items" $configs | toJson -}}
{{- end -}}


{{- /*
  Notifications service ENCRYPT_DECRYPT_KEY configuration.
  The service encrypts sensitive sink credentials with ENCRYPT_DECRYPT_KEY;
  losing it makes every stored sink credential unreadable. It is generated
  once (32-char alphanumeric) and preserved across upgrades and ArgoCD syncs.

  On first upgrade the existing ENCRYPT_DECRYPT_KEY is migrated from the legacy
  standalone Secret so the key stays stable when notifications-service is folded
  into control-plane-chart.

  The legacy name is derived from the RELEASE NAME, not fullnameOverride:
  "{{ .Release.Name }}-notifications-service". Unlike the other control-plane
  services (whose legacy Secrets are fullnameOverride-based, e.g.
  runai-backend-diagnostics-service), notifications-service was a subchart whose
  Secret was named "{Release.Name}-notifications-service". On SaaS pools the
  release name is NOT runai-backend (e.g. saas-pool1 -> saas-pool1-notifications-service),
  so a fullnameOverride-based default would MISS the legacy Secret and regenerate
  the key, making stored sink credentials unreadable. Because this generator runs
  in the same release, .Release.Name reconstructs the exact legacy name on every
  environment. 
  Sink credentials stay in the chart-managed Secret; only ENCRYPT_DECRYPT_KEY
  moves here.
*/ -}}
{{- define "secret-generator.notifications-service-configs" -}}
{{- $prefix := .Values.fullnameOverride }}
{{- $configs := list }}
{{- if .Values.notificationsService.enabled }}
{{- $encKey := dict "name" "ENCRYPT_DECRYPT_KEY" "type" "alphanumeric" "length" 32 }}
{{- $legacyName := printf "%s-notifications-service" .Release.Name }}
{{- $configs = append $configs (dict
  "name" (printf "%s-gen-notifications-service" $prefix)
  "legacySecretName" $legacyName
  "keys" (list $encKey)
) }}
{{- end }}
{{- dict "items" $configs | toJson -}}
{{- end -}}
