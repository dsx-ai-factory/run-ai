module github.com/dsx-ai-factory/run-ai/runai-mcp-server

go 1.27.2

require (
	github.com/Masterminds/semver/v3 v3.5.0
	github.com/MicahParks/keyfunc/v3 v3.8.2
	github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client v1.1.1
	github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client v1.0.7
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v1.0.1
	github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client v1.0.21
	github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client v0.0.1
	github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils v1.0.8
	github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics v1.2.7
	github.com/go-chi/chi/v5 v5.5.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/jsonschema-go v0.4.3
	github.com/google/uuid v1.6.0
	github.com/modelcontextprotocol/go-sdk v1.8.0
	github.com/rs/zerolog v1.35.1
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.12.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/MicahParks/jwkset v0.11.3 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils v1.0.3 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware v1.0.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.3.2 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.17 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.9 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.30 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/encoding v0.5.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client => ../control-plane/authorization/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client => ../runai-clients/assets-service-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client => ../runai-clients/catalog-service-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client => ../runai-clients/datavolumes-service-client

replace github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client => ../control-plane/org-unit-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client => ../runai-clients/policy-service-client

replace github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client => ../control-plane/workloads/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client => ../runai-clients/workloads-manager-client

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../control-plane/cluster-service/pkg/client
