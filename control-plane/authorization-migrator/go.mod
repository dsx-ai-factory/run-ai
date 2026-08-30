module github.com/run-ai/runai/control-plane/authorization-migrator

go 1.26.3

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/Nerzal/gocloak/v13 v13.9.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/onsi/ginkgo/v2 v2.32.1
	github.com/onsi/gomega v1.43.0
	github.com/run-ai/runai/control-plane/authorization v0.0.0
	github.com/run-ai/runai/runai-common-packages/config v1.2.8
	github.com/run-ai/runai/runai-common-packages/fips v1.0.0
	github.com/run-ai/runai/runai-common-packages/logging v1.0.13
	github.com/run-ai/runai/runai-common-packages/metrics v1.2.4
	github.com/run-ai/runai/runai-common-packages/runaidb v1.0.11
	gorm.io/driver/postgres v1.6.2
	gorm.io/gorm v1.31.2
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3
)

require (
	filippo.io/edwards25519 v1.1.1 // indirect
	github.com/ClickHouse/ch-go v0.73.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.47.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Rican7/retry v0.3.1 // indirect
	github.com/ThreeDotsLabs/watermill v1.3.5 // indirect
	github.com/ThreeDotsLabs/watermill-redisstream v1.3.0 // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/aws/smithy-go v1.27.8 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.3 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/golang-migrate/migrate/v4 v4.19.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.11.2 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/lithammer/shortuuid/v4 v4.2.0 // indirect
	github.com/maypok86/otter/v2 v2.2.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.48.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.27 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.64.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/redis/go-redis/v9 v9.14.0 // indirect
	github.com/run-ai/runai/control-plane/audit-service/pkg/client v1.0.7 // indirect
	github.com/run-ai/runai/control-plane/authorization/pkg/client v0.0.0 // indirect
	github.com/run-ai/runai/runai-application-packages/client-utils v1.0.1 // indirect
	github.com/run-ai/runai/runai-application-packages/jsonutils v1.0.2 // indirect
	github.com/run-ai/runai/runai-common-packages/cache/v2 v2.4.3 // indirect
	github.com/run-ai/runai/runai-common-packages/runaictx v1.1.5 // indirect
	github.com/run-ai/runai/runai-common-packages/runaihttp v1.1.5 // indirect
	github.com/run-ai/runai/runai-common-packages/tracing v1.0.25 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/kafka-go v0.4.48 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/stretchr/testify v1.12.1
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/grpc v1.82.1 // indirect
	gorm.io/driver/clickhouse v0.7.0 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/plugin/opentelemetry v0.1.15 // indirect
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-chi/chi/v5 v5.2.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/run-ai/runai/runai-application-packages/dbmodels v1.0.1 // indirect
	github.com/run-ai/runai/runai-application-packages/queryapi v1.0.2 // indirect
	github.com/run-ai/runai/runai-common-packages/runaipubsub v1.6.3 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	golang.org/x/crypto v0.53.0 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/run-ai/runai/control-plane/authorization v0.0.0 => ../authorization

replace github.com/run-ai/runai/control-plane/authorization/pkg/client => ../../control-plane/authorization/pkg/client

replace github.com/run-ai/runai/control-plane/cluster-service/pkg/client => ../../control-plane/cluster-service/pkg/client

replace github.com/run-ai/runai/runai-clients/workloads-client => ../../runai-clients/workloads-client

replace github.com/run-ai/runai/runai-clients/datavolumes-service-client => ../../runai-clients/datavolumes-service-client

replace github.com/run-ai/runai/runai-clients/org-unit-service-client => ../../runai-clients/org-unit-service-client

replace github.com/run-ai/runai/runai-cluster/common/cluster-capabilities => ../../runai-cluster/common/cluster-capabilities
