module github.com/dsx-ai-factory/run-ai/control-plane/assets-service

go 1.27.2

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/docker/go-units v0.5.0
	github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client v1.0.7
	github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events v1.0.6
	github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils v1.0.5
	github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs v1.0.4
	github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels v1.0.1
	github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils v1.0.5
	github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware v1.0.1
	github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo v1.1.1
	github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils v1.0.2
	github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk v0.0.0-20250907093507-4e419875bd3d
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities v1.0.5
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2 v2.4.8
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.3.1
	github.com/dsx-ai-factory/run-ai/runai-common-packages/fips v1.0.0
	github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb v1.0.2
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.15
	github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics v1.2.6
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.7
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb v1.0.11
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp v1.1.5
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub v1.6.3
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.29
	github.com/ghodss/yaml v1.0.0
	github.com/go-chi/chi/v5 v5.3.2
	github.com/google/uuid v1.6.0
	github.com/iancoleman/strcase v0.3.0
	github.com/jinzhu/copier v0.4.0
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/testify v1.12.1
	go.uber.org/zap v1.28.0
	gorm.io/gorm v1.31.2
	k8s.io/api v0.36.2
	k8s.io/apimachinery v0.36.2
	k8s.io/client-go v1.5.2
	k8s.io/kubernetes v1.35.4
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2
	sigs.k8s.io/controller-runtime v0.23.3
)

require (
	dario.cat/mergo v1.0.2 // indirect
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/ClickHouse/ch-go v0.74.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.49.0 // indirect
	github.com/DATA-DOG/go-sqlmock v1.5.2 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/NVIDIA/k8s-dra-driver-gpu v0.0.0-20260329235847-c657396bd619 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/PaesslerAG/jsonpath v0.1.1 // indirect
	github.com/Rican7/retry v0.3.1 // indirect
	github.com/ThreeDotsLabs/watermill v1.3.5 // indirect
	github.com/ThreeDotsLabs/watermill-redisstream v1.3.0 // indirect
	github.com/andybalholm/brotli v1.2.6 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/smithy-go v1.28.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v0.2.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/docker/go-connections v0.7.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils v1.0.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/validation v1.0.2 // indirect
	github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority v0.0.0 // indirect
	github.com/ebitengine/purego v0.10.1 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.8.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.1 // indirect
	github.com/go-openapi/jsonreference v1.0.2 // indirect
	github.com/go-openapi/swag v0.29.2 // indirect
	github.com/go-openapi/swag/cmdutils v0.29.2 // indirect
	github.com/go-openapi/swag/conv v0.29.2 // indirect
	github.com/go-openapi/swag/fileutils v0.29.2 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/go-openapi/swag/loading v0.29.2 // indirect
	github.com/go-openapi/swag/mangling v0.29.2 // indirect
	github.com/go-openapi/swag/netutils v0.29.2 // indirect
	github.com/go-openapi/swag/pools v0.29.2 // indirect
	github.com/go-openapi/swag/stringutils v0.29.2 // indirect
	github.com/go-openapi/swag/typeutils v0.29.2 // indirect
	github.com/go-openapi/swag/yamlutils v0.29.2 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/go-sql-driver/mysql v1.10.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-migrate/migrate/v4 v4.19.1 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260507013755-92041b743c96 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.32.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kai-scheduler/KAI-scheduler v0.18.2 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/kubeflow/training-operator v1.9.3 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lib/pq v1.11.2 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/lithammer/shortuuid/v4 v4.3.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20260330125221-c963978e514e // indirect
	github.com/magiconair/properties v1.8.10 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/maypok86/otter/v2 v2.3.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.3.3 // indirect
	github.com/moby/moby/api v1.56.1 // indirect
	github.com/moby/moby/client v0.6.0 // indirect
	github.com/moby/patternmatcher v0.6.1 // indirect
	github.com/moby/sys/sequential v0.7.0 // indirect
	github.com/moby/sys/user v0.4.1 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.54.0 // indirect
	github.com/nats-io/nkeys v0.4.16 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/novln/docker-parser v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openshift/api v0.0.0-20260605005319-1194f4c62539 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.88.0 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/redis/go-redis/v9 v9.23.0 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/segmentio/kafka-go v0.4.48 // indirect
	github.com/shirou/gopsutil/v4 v4.26.6 // indirect
	github.com/shopspring/decimal v1.5.0 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/testcontainers/testcontainers-go v0.44.0 // indirect
	github.com/testcontainers/testcontainers-go/modules/postgres v0.43.0 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
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
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/exp v0.0.0-20261014143641-f0251eb4aa3e // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/clickhouse v0.7.0 // indirect
	gorm.io/driver/mysql v1.6.0 // indirect
	gorm.io/driver/postgres v1.6.3 // indirect
	gorm.io/plugin/opentelemetry v0.1.16 // indirect
	k8s.io/apiextensions-apiserver v0.35.4 // indirect
	k8s.io/apiserver v0.35.4 // indirect
	k8s.io/autoscaler/vertical-pod-autoscaler v1.5.1 // indirect
	k8s.io/component-base v0.35.4 // indirect
	k8s.io/controller-manager v0.35.4 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	knative.dev/pkg v0.0.0-20260120122510-4a022ed9999a // indirect
	sigs.k8s.io/gateway-api v1.5.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.22.4

replace github.com/dsx-ai-factory/run-ai/control-plane/policy-service => ../policy-service

replace github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies => ./../workloads-and-policies

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority => ../../runai-cluster/common/priority

// required due to a replace in runai-operator repo
replace (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ../../runai-cluster/cluster/common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions => ../../runai-cluster/cluster/common/extensions
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds => ../../runai-cluster/cluster/common/personalcreds
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/runaipublic => ../../runai-cluster/cluster/common/runaipublic
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils => ../../runai-cluster/cluster/common/utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk => ../../runai-cluster/cluster/sdk
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common => ../../runai-cluster/cluster/workload-common
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller => ../../runai-cluster/cluster/workload-controller
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cert-utils => ../../runai-cluster/common/cert-utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities => ../../runai-cluster/common/cluster-capabilities
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants => ../../runai-cluster/common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/profiling => ../../runai-cluster/common/profiling
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner => ../../runai-cluster/common/top-owner
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils => ../../runai-cluster/common/utils
)

replace github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client => ../../control-plane/authorization/pkg/client

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../../control-plane/cluster-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client => ../../runai-clients/datavolumes-service-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client => ../../runai-clients/k8s-objects-tracker-client

replace github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client => ../../control-plane/org-unit-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client => ../../runai-clients/policy-service-client

replace github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client => ../../control-plane/workloads/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client => ../../runai-clients/workloads-manager-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client => ../../runai-clients/assets-service-client

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/topology => ../../runai-cluster/cluster/common/topology

replace k8s.io/client-go => k8s.io/client-go v0.35.4

replace k8s.io/apimachinery => k8s.io/apimachinery v0.35.4

// workloads-and-policies requires k8s.io/api/apiextensions v0.35.x, but its own
// replace pins (to the 1.34 line) are ignored across module boundaries. Cap them
// here so client-go v0.34.2 keeps finding k8s.io/api/storagemigration/v1alpha1
// (removed from k8s.io/api in v0.36).
replace (
	k8s.io/api => k8s.io/api v0.35.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.35.4
)

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 => ../../runai-cluster/cluster/dromos/api/v1alpha1
