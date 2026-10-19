module github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller

go 1.27.2

require (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants v0.0.0-20261019150750-8e418cccab3e
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority v0.0.0-20261019150750-8e418cccab3e
	github.com/dsx-ai-factory/run-ai/runai-common-packages/fips v1.0.3
	github.com/dyson/certman v0.2.1
	github.com/golang/mock v1.6.0
	github.com/magiconair/properties v1.18.12
	github.com/stretchr/testify v1.12.1
	golang.org/x/exp v0.0.0-20261014143641-f0251eb4aa3e
	gomodules.xyz/jsonpatch/v2 v2.5.0
	k8s.io/api v0.35.4
	k8s.io/apimachinery v0.35.4
	k8s.io/client-go v0.35.4
	k8s.io/utils v0.0.0-20260707023825-cf1189d6abe3
	sigs.k8s.io/controller-runtime v0.23.3
	sigs.k8s.io/karpenter v1.14.1
)

replace (
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ../common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra => ../common/dra
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants => ../../common/constants
	github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority => ../../common/priority
)

require (
	github.com/awslabs/operatorpkg v0.0.0-20261006163503-855c7d06c363 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/config v1.3.2 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/logging v1.0.17 // indirect
	github.com/emicklei/go-restful/v3 v3.14.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.6 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/go-openapi/jsonreference v0.21.6 // indirect
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
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rs/zerolog v1.35.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/samber/lo v1.53.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/apiextensions-apiserver v0.35.4 // indirect
	k8s.io/dynamic-resource-allocation v0.35.4 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20251125145642-4e65d59e963e // indirect
	k8s.io/kubernetes v1.35.4 // indirect
	sigs.k8s.io/json v0.0.0-20260909141634-11ed52e25bc5 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)
