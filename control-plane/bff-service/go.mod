module github.com/dsx-ai-factory/run-ai/control-plane/bff-service

go 1.27.2

require (
	github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client v1.0.12
	github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client v1.0.3
	github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client v1.0.7
	github.com/dsx-ai-factory/run-ai/control-plane/temp-backend-settings/pkg/client v1.0.6
	github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client v1.0.8
	github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2 v2.4.8
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.3.2
	github.com/dsx-ai-factory/run-ai/runai-common-packages/fips v1.0.3
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.17
	github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics v1.2.7
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.9
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb v1.0.19
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp v1.1.11
	github.com/dsx-ai-factory/run-ai/runai-common-packages/setting v1.2.5
	github.com/go-chi/chi/v5 v5.5.0
	github.com/google/uuid v1.6.0
	github.com/stretchr/testify v1.12.1
	gorm.io/gorm v1.31.2
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/ClickHouse/ch-go v0.74.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.49.0 // indirect
	github.com/Rican7/retry v0.3.1 // indirect
	github.com/ThreeDotsLabs/watermill v1.5.3 // indirect
	github.com/ThreeDotsLabs/watermill-redisstream v1.4.5 // indirect
	github.com/andybalholm/brotli v1.2.6 // indirect
	github.com/aws/smithy-go v1.28.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils v1.0.3 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub v1.6.7 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.30 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.8.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-migrate/migrate/v4 v4.20.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.32.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/lithammer/shortuuid/v4 v4.3.0 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/maypok86/otter/v2 v2.3.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.54.0 // indirect
	github.com/nats-io/nkeys v0.4.16 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/kafka-go v0.4.51 // indirect
	github.com/shopspring/decimal v1.6.0 // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.1 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/exp v0.0.0-20261014143641-f0251eb4aa3e // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	gorm.io/driver/clickhouse v0.7.0 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.3 // indirect
	gorm.io/plugin/opentelemetry v0.1.16 // indirect
)

replace github.com/dsx-ai-factory/run-ai/control-plane/bff-service/pkg/client => ./pkg/client

replace github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client => ../../control-plane/authorization/pkg/client
