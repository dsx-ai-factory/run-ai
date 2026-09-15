module github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api

go 1.26.6

require (
	github.com/MicahParks/keyfunc v1.9.0
	github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client v1.0.7
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk v0.0.0-20250907093507-4e419875bd3d
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/auth v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2 v2.4.3
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.2.8
	github.com/dsx-ai-factory/run-ai/runai-common-packages/fips v1.0.0
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.13
	github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics v1.2.4
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.5
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp v1.1.5
	github.com/go-chi/chi/v5 v5.2.2
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/kubeflow/training-operator v1.9.3
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee
	github.com/onsi/ginkgo/v2 v2.28.1
	github.com/onsi/gomega v1.41.0
	github.com/rs/zerolog v1.35.1
	golang.org/x/sync v0.20.0
	gorm.io/gorm v1.25.7
	k8s.io/api v0.35.4
	k8s.io/apimachinery v0.35.4
	k8s.io/client-go v1.5.2
	k8s.io/kubectl v0.35.4
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2
	nhooyr.io/websocket v1.8.11
	sigs.k8s.io/controller-runtime v0.23.3
)

require (
	cel.dev/expr v0.25.1 // indirect
	cyphar.com/go-pathrs v0.2.1 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cyphar/filepath-securejoin v0.6.0 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/cel-go v0.26.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/opencontainers/selinux v1.13.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.61.0 // indirect
	golang.org/x/mod v0.35.0 // indirect
	k8s.io/apiserver v0.35.4 // indirect
	k8s.io/cloud-provider v0.35.4 // indirect
	k8s.io/component-base v0.35.4 // indirect
	k8s.io/component-helpers v0.35.4 // indirect
	k8s.io/controller-manager v0.35.4 // indirect
	k8s.io/cri-api v0.35.4 // indirect
	k8s.io/cri-client v0.35.4 // indirect
	k8s.io/csi-translation-lib v0.35.4 // indirect
	k8s.io/dynamic-resource-allocation v0.35.4 // indirect
	k8s.io/kube-scheduler v0.35.4 // indirect
	k8s.io/kubelet v0.35.4 // indirect
	k8s.io/kubernetes v1.35.4 // indirect
	k8s.io/mount-utils v0.35.4 // indirect
)

require (
	github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client v0.0.0
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/kai-scheduler/KAI-scheduler v0.16.9 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
)

require (
	github.com/Rican7/retry v0.3.1 // indirect
	github.com/ThreeDotsLabs/watermill v1.3.5 // indirect
	github.com/ThreeDotsLabs/watermill-redisstream v1.3.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/smithy-go v1.27.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cenkalti/backoff/v3 v3.2.2 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cert-manager/cert-manager v1.20.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-oidc/v3 v3.11.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils v1.0.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs v1.0.4 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils v1.0.2 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common v0.0.0-00010101000000-000000000000 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/auth v0.0.0-20250428075130-be7291f11588 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub v1.6.3 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.25 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.4 // indirect
	github.com/go-openapi/jsonreference v0.21.4 // indirect
	github.com/go-openapi/swag v0.25.1 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.1 // indirect
	github.com/go-openapi/swag/conv v0.25.1 // indirect
	github.com/go-openapi/swag/fileutils v0.25.1 // indirect
	github.com/go-openapi/swag/jsonname v0.25.4 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.1 // indirect
	github.com/go-openapi/swag/loading v0.25.1 // indirect
	github.com/go-openapi/swag/mangling v0.25.1 // indirect
	github.com/go-openapi/swag/netutils v0.25.1 // indirect
	github.com/go-openapi/swag/stringutils v0.25.1 // indirect
	github.com/go-openapi/swag/typeutils v0.25.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-containerregistry v0.20.5 // indirect
	github.com/google/pprof v0.0.0-20260115054156-294ebfa9ad83 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.10.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kai-scheduler/kai-resource-management-api v0.1.2 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lithammer/shortuuid/v3 v3.0.7 // indirect
	github.com/lithammer/shortuuid/v4 v4.2.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/maypok86/otter/v2 v2.2.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/moby/spdystream v0.5.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/nats-io/nats.go v1.48.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/novln/docker-parser v1.0.0 // indirect
	github.com/oklog/ulid v1.3.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/openshift/api v0.0.0-20250901120840-a638ff2e96fb // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.88.0 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/redis/go-redis/v9 v9.14.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/kafka-go v0.4.48 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.64.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260427160629-7cedc36a6bc4 // indirect
	google.golang.org/grpc v1.82.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.35.4 // indirect
	k8s.io/autoscaler/vertical-pod-autoscaler v1.5.1 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260127142750-a19766b6e2d4 // indirect
	knative.dev/networking v0.0.0-20260120131110-a7cdca238a0d // indirect
	knative.dev/pkg v0.0.0-20260120122510-4a022ed9999a // indirect
	knative.dev/serving v0.48.1 // indirect
	sigs.k8s.io/gateway-api v1.5.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ../../cluster/common/constants

	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions => ../common/extensions
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters => ../../cluster/common/getters
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/runaipublic => ../common/runaipublic
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/topology => ../common/topology

	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils => ../common/utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator => ../go-operator

	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/researcher-service => ../researcher-service

	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk => ../sdk

	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common => ../workload-common

	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller => ../workload-controller

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/auth => ../../common/auth

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cert-utils => ../../common/cert-utils

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/clients => ../../common/clients

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants => ../../common/constants

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority => ../../common/priority

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/profiling => ../../common/profiling

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner => ../../common/top-owner

	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils => ../../common/utils
)

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../../../control-plane/cluster-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client => ../../../control-plane/authorization/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client => ../../../runai-clients/workloads-client

replace k8s.io/client-go => k8s.io/client-go v0.35.4

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.23.3

replace knative.dev/serving => knative.dev/serving v0.44.0

replace knative.dev/pkg => knative.dev/pkg v0.0.0-20250424013628-d5e74d29daa3

replace knative.dev/networking => knative.dev/networking v0.0.0-20250117155906-67d1c274ba6a

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 => ../dromos/api/v1alpha1

replace github.com/dsx-ai-factory/run-ai/runai-common-packages/auth => ./../../common/runai-auth
