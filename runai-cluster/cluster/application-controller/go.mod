module github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/application-controller

go 1.26.8

require (
	github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client v1.0.11
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants v0.0.0-20240519090715-44f7f790e5ee
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-common-packages/fips v1.0.0
	github.com/google/uuid v1.6.0
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/rs/zerolog v1.35.1
	go.uber.org/zap v1.28.0
	helm.sh/helm/v3 v3.21.3
	k8s.io/api v0.36.2
	k8s.io/apiextensions-apiserver v0.36.2
	k8s.io/apimachinery v0.36.2
	k8s.io/client-go v1.5.2
	k8s.io/utils v0.0.0-20260626114624-be93311217bd
	sigs.k8s.io/controller-runtime v0.24.1
)

require (
	cel.dev/expr v0.25.2 // indirect
	cyphar.com/go-pathrs v0.2.5 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v0.0.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/go-openapi/swag/pools v0.27.1 // indirect
	github.com/google/cel-go v0.29.0 // indirect
	github.com/moby/sys/mountinfo v0.7.2 // indirect
	github.com/opencontainers/selinux v1.13.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.67.0 // indirect
	golang.org/x/exp v0.0.0-20260410095643-746e56fc9e2f // indirect
	golang.org/x/mod v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260720211330-0afa2a65878a // indirect
	k8s.io/cloud-provider v0.36.2 // indirect
	k8s.io/component-helpers v0.36.2 // indirect
	k8s.io/controller-manager v0.36.2 // indirect
	k8s.io/cri-api v0.36.2 // indirect
	k8s.io/cri-client v0.36.2 // indirect
	k8s.io/csi-translation-lib v0.36.2 // indirect
	k8s.io/dynamic-resource-allocation v0.36.2 // indirect
	k8s.io/kube-scheduler v0.36.2 // indirect
	k8s.io/kubelet v0.36.2 // indirect
	k8s.io/kubernetes v1.36.2 // indirect
	k8s.io/mount-utils v0.36.2 // indirect
	k8s.io/streaming v0.36.2 // indirect
)

require (
	dario.cat/mergo v1.0.1 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/BurntSushi/toml v1.6.0 // indirect
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/Masterminds/sprig/v3 v3.3.0 // indirect
	github.com/Masterminds/squirrel v1.5.4 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cert-manager/cert-manager v1.20.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/chai2010/gettext-go v1.0.2 // indirect
	github.com/cyphar/filepath-securejoin v0.7.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils v1.0.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware v1.0.1 // indirect
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 v0.0.0 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.2.8 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.13 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.5 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.25 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch v5.9.11+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exponent-io/jsonpath v0.0.0-20210407135951-1de76d718b3f // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-gorp/gorp/v3 v3.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/go-openapi/jsonreference v0.21.5 // indirect
	github.com/go-openapi/swag v0.27.1 // indirect
	github.com/go-openapi/swag/cmdutils v0.27.1 // indirect
	github.com/go-openapi/swag/conv v0.27.1 // indirect
	github.com/go-openapi/swag/fileutils v0.27.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.27.1 // indirect
	github.com/go-openapi/swag/loading v0.27.1 // indirect
	github.com/go-openapi/swag/mangling v0.27.1 // indirect
	github.com/go-openapi/swag/netutils v0.27.1 // indirect
	github.com/go-openapi/swag/stringutils v0.27.1 // indirect
	github.com/go-openapi/swag/typeutils v0.27.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.27.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.30.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260507013755-92041b743c96 // indirect
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674 // indirect
	github.com/gosuri/uitable v0.0.4 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kai-scheduler/KAI-scheduler v0.18.2 // indirect
	github.com/kai-scheduler/kai-resource-management-api v0.1.3
	github.com/klauspost/compress v1.18.6 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.12.3 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/spdystream v0.5.1 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openshift/api v0.0.0-20260605005319-1194f4c62539 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring v0.88.0 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.5 // indirect
	github.com/prometheus/procfs v0.20.1 // indirect
	github.com/redis/go-redis/v9 v9.10.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rubenv/sql-migrate v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/term v0.46.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.5.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260720211330-0afa2a65878a // indirect
	google.golang.org/grpc v1.83.2 // indirect
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.36.2 // indirect
	k8s.io/autoscaler/vertical-pod-autoscaler v1.5.1 // indirect
	k8s.io/cli-runtime v0.36.2 // indirect
	k8s.io/component-base v0.36.2 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20260317180543-43fb72c5454a // indirect
	k8s.io/kubectl v0.36.2 // indirect
	oras.land/oras-go/v2 v2.6.2 // indirect
	sigs.k8s.io/gateway-api v1.5.0 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/kustomize/api v0.21.1 // indirect
	sigs.k8s.io/kustomize/kyaml v0.21.1 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ../common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/getters => ../common/getters
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/personalcreds => ../common/personalcreds
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/runaipublic => ../common/runaipublic
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init => ../common/scheme-init
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils => ../common/utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk => ../sdk
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common => ../workload-common
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/cert-utils => ../../common/cert-utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants => ../../common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/node-pool-utils => ../../common/node-pool-utils
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/profiling => ../../common/profiling
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils => ../../common/utils
)

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../../../control-plane/cluster-service/pkg/client

replace k8s.io/client-go => k8s.io/client-go v0.36.2

replace sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.24.1

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1 => ../dromos/api/v1alpha1

replace k8s.io/cri-streaming => k8s.io/cri-streaming v0.36.2
