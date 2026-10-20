module github.com/dsx-ai-factory/run-ai/control-plane/runai-cli

go 1.27.2

require (
	github.com/charmbracelet/bubbles v0.18.0
	github.com/charmbracelet/bubbletea v0.26.4
	github.com/charmbracelet/lipgloss v0.11.0
	github.com/deckarep/golang-set/v2 v2.6.0
	github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client v1.0.11
	github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils v1.0.5
	github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/cli-exposer-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/cluster-api-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client v0.0.0
	github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client v0.0.0-00010101000000-000000000000
	github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang/mock v1.6.0
	github.com/hashicorp/go-version v1.9.0
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/onsi/ginkgo/v2 v2.27.2
	github.com/onsi/gomega v1.38.2
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c
	github.com/prometheus/client_golang v1.25.0
	github.com/samber/lo v1.51.0
	github.com/siderolabs/go-kubeconfig v0.1.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.58.0
	golang.org/x/exp v0.0.0-20250103183323-7d7fa50e5329
	golang.org/x/oauth2 v0.37.0
	golang.org/x/sync v0.24.0
	golang.org/x/term v0.47.0
	golang.org/x/text v0.43.0
	gopkg.in/yaml.v3 v3.0.1
	helm.sh/helm/v3 v3.20.2
	k8s.io/api v0.35.4
	k8s.io/apimachinery v0.35.4
	k8s.io/client-go v0.35.4
	k8s.io/utils v0.0.0-20260507154919-ff6756f316d2
	modernc.org/mathutil v1.5.0
	nhooyr.io/websocket v1.8.11
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/atotto/clipboard v0.1.4 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.9.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/charmbracelet/x/ansi v0.1.2 // indirect
	github.com/charmbracelet/x/input v0.1.0 // indirect
	github.com/charmbracelet/x/term v0.1.1 // indirect
	github.com/charmbracelet/x/windows v0.1.0 // indirect
	github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx v1.1.7 // indirect
	github.com/dsx-ai-factory/run-ai/runai-common-packages/tracing v1.0.29 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/erikgeiser/coninput v0.0.0-20211004153227-1c3628e74d0f // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.22.1 // indirect
	github.com/go-openapi/jsonreference v0.21.2 // indirect
	github.com/go-openapi/swag v0.25.1 // indirect
	github.com/go-openapi/swag/cmdutils v0.25.1 // indirect
	github.com/go-openapi/swag/conv v0.25.1 // indirect
	github.com/go-openapi/swag/fileutils v0.25.1 // indirect
	github.com/go-openapi/swag/jsonname v0.25.1 // indirect
	github.com/go-openapi/swag/jsonutils v0.25.1 // indirect
	github.com/go-openapi/swag/loading v0.25.1 // indirect
	github.com/go-openapi/swag/mangling v0.25.1 // indirect
	github.com/go-openapi/swag/netutils v0.25.1 // indirect
	github.com/go-openapi/swag/stringutils v0.25.1 // indirect
	github.com/go-openapi/swag/typeutils v0.25.1 // indirect
	github.com/go-openapi/swag/yamlutils v0.25.1 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/go-playground/validator/v10 v10.30.5 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250820193118-f64d9cf942d6 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/muesli/ansi v0.0.0-20230316100256-276c6243b2f6 // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.72.1 // indirect
	github.com/prometheus/procfs v0.22.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/sahilm/fuzzy v0.1.1-0.20230530133925-c48e322e2a8f // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.11 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/arch v0.3.0 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/time v0.14.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250910181357-589584f1c912 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

replace github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants => ../../runai-cluster/cluster/common/constants

replace github.com/dsx-ai-factory/run-ai/runai-clients/cli-exposer-client => ../../runai-clients/cli-exposer-client

replace github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client => ../../control-plane/cluster-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client => ../../runai-clients/assets-service-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client => ../../runai-clients/k8s-objects-tracker-client

replace github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client => ../../control-plane/org-unit-service/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client => ../../runai-clients/policy-service-client

replace github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client => ../../control-plane/workloads/pkg/client

replace github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client => ../../runai-clients/workloads-manager-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/cluster-api-client => ../../runai-clients/cluster-api-client

replace github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client => ../../runai-clients/datavolumes-service-client
