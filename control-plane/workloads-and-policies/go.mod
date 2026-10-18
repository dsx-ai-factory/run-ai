module github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies

go 1.27.1

require (
	github.com/NVIDIA/k8s-nim-operator v0.0.0-20260601160319-679627a97be7
	github.com/PaesslerAG/jsonpath v0.1.1
	github.com/distribution/reference v0.6.0
	github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils v1.0.5
	github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs v1.0.4
	github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels v1.0.1
	github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils v1.0.5
	github.com/dsx-ai-factory/run-ai/runai-application-packages/scopeinfo v1.1.1
	github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils v1.0.2
	github.com/dsx-ai-factory/run-ai/runai-application-packages/validation v1.0.2
	github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk v0.0.0-20250907093507-4e419875bd3d
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants v0.0.0-20251016112722-024bf1e7a8f1
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.15
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.6
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp v1.1.5
	github.com/ghodss/yaml v1.0.0
	github.com/google/uuid v1.6.0
	github.com/itchyny/gojq v0.12.19
	github.com/kubeflow/training-operator v1.9.3
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/stretchr/testify v1.12.1
	golang.org/x/exp v0.0.0-20260908205506-85c1c2202aba
	gorm.io/gorm v1.31.2
	k8s.io/api v0.36.2
	k8s.io/apimachinery v0.36.2
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2
	knative.dev/pkg v0.0.0-20260120122510-4a022ed9999a
	knative.dev/serving v0.48.1
	sigs.k8s.io/lws v0.7.0
)

require (
	cel.dev/expr v0.25.2 // indirect
	cloud.google.com/go v0.121.6 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.5.3 // indirect
	cloud.google.com/go/monitoring v1.24.3 // indirect
	cloud.google.com/go/storage v1.56.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.19.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.2 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.5.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/detectors/gcp v1.34.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric v0.53.0 // indirect
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/internal/resourcemapping v0.53.0 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/NVIDIA/k8s-dra-driver-gpu v0.0.0-20260329235847-c657396bd619 // indirect
	github.com/PaesslerAG/gval v1.0.0 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aws/aws-sdk-go v1.55.6 // indirect
	github.com/aws/smithy-go v1.28.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cncf/xds/go v0.0.0-20260202195803-dba9d589def2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils v1.0.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2 v2.4.6 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.3.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics v1.2.5 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.27 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/envoyproxy/go-control-plane/envoy v1.37.0 // indirect
	github.com/envoyproxy/protoc-gen-validate v1.3.3 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/expr-lang/expr v1.17.6 // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.15 // indirect
	github.com/go-chi/chi/v5 v5.3.2 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v1.0.1 // indirect
	github.com/go-openapi/jsonreference v1.0.1 // indirect
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
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/glog v1.2.5 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/cel-go v0.29.0 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-containerregistry v0.20.5 // indirect
	github.com/google/pprof v0.0.0-20260507013755-92041b743c96 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.15 // indirect
	github.com/googleapis/gax-go/v2 v2.22.0 // indirect
	github.com/googleapis/google-cloud-go-testing v0.0.0-20210719221736-1c9a4c676720 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.31.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.8 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jmespath/go-jmespath v0.4.1-0.20220621161143-b0104c826a24 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kai-scheduler/KAI-scheduler v0.18.2 // indirect
	github.com/kedacore/keda/v2 v2.18.0 // indirect
	github.com/klauspost/compress v1.20.1 // indirect
	github.com/kserve/kserve v0.17.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lithammer/shortuuid/v4 v4.3.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/maypok86/otter/v2 v2.3.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.54.0 // indirect
	github.com/nats-io/nkeys v0.4.16 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/novln/docker-parser v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.88.0 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.71.0 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/redis/go-redis/v9 v9.22.0 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/spiffe/go-spiffe/v2 v2.8.1 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/detectors/gcp v1.44.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.71.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.71.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.47.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk v1.47.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.1 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/api v0.278.0 // indirect
	google.golang.org/genproto v0.0.0-20260319201613-d00831a3d3e7 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260928230214-8a89bd6388cc // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260928230214-8a89bd6388cc // indirect
	google.golang.org/grpc v1.84.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/go-playground/validator.v9 v9.31.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiextensions-apiserver v0.35.4 // indirect
	k8s.io/apiserver v0.35.4 // indirect
	k8s.io/autoscaler/vertical-pod-autoscaler v1.5.1 // indirect
	k8s.io/client-go v1.5.2 // indirect
	k8s.io/component-base v0.35.4 // indirect
	k8s.io/dynamic-resource-allocation v0.35.4 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	knative.dev/networking v0.0.0-20260120131110-a7cdca238a0d // indirect
	sigs.k8s.io/controller-runtime v0.23.3 // indirect
	sigs.k8s.io/gateway-api v1.5.0 // indirect
	sigs.k8s.io/gateway-api-inference-extension v1.3.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace (
	// Force all dependencies to use k8s.io v0.34.1 to avoid type incompatibilities
	k8s.io/api => k8s.io/api v0.35.4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.35.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.35.4
	k8s.io/apiserver => k8s.io/apiserver v0.35.4
	k8s.io/client-go => k8s.io/client-go v0.35.4
	k8s.io/component-base => k8s.io/component-base v0.35.4
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.35.4
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.22.4
)

replace github.com/btnguyen2k/consu.git => ./third_party/consu/checksum

// kserve upstream does not compile against k8s.io v0.35 (PodSpec struct conversion adds
// WorkloadRef); use a locally vendored+patched copy of kserve v0.19.0 that builds against
// the pinned k8s version. See third_party/kserve.
replace github.com/kserve/kserve => ./third_party/kserve

// required due to a replace in runai-operator repo
replace (
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants => ./../../runai-cluster/common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/top-owner => ./../../runai-cluster/common/top-owner
)

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ./../../runai-cluster/cluster/common/constants

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds => ./../../runai-cluster/cluster/common/personalcreds

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller => ./../../runai-cluster/cluster/workload-controller

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common => ./../../runai-cluster/cluster/workload-common

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk => ./../../runai-cluster/cluster/sdk

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils => ./../../runai-cluster/cluster/common/utils

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions => ./../../runai-cluster/cluster/common/extensions

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/runaipublic => ./../../runai-cluster/cluster/common/runaipublic

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/topology => ./../../runai-cluster/cluster/common/topology

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/cert-utils => ./../../runai-cluster/common/cert-utils

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/profiling => ./../../runai-cluster/common/profiling

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils => ./../../runai-cluster/common/utils

replace github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client => ../../runai-clients/assets-service-client

replace github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client => ../../control-plane/authorization/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client => ../../runai-clients/k8s-objects-tracker-client

replace github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client => ../../control-plane/org-unit-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client => ../../runai-clients/policy-service-client

replace github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client => ../../control-plane/workloads/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client => ../../runai-clients/workloads-manager-client

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../../control-plane/cluster-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities => ../../runai-cluster/common/cluster-capabilities

replace github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority => ../../runai-cluster/common/priority

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 => ../../runai-cluster/cluster/dromos/api/v1alpha1
