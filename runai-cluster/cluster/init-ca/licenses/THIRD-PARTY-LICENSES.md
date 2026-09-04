# Third-Party Software Licenses — init-ca

This file inventories the third-party Go modules compiled into the `init-ca`
service and the open-source license under which each is distributed.

It is generated with [`go-licenses`](https://github.com/google/go-licenses):

```
cd runai-cluster/cluster/init-ca
go-licenses report ./...
```

First-party `github.com/dsx-ai-factory/run-ai/...` modules are omitted — they are part of
this repository and are covered by the root [LICENSE](../../../../LICENSE) (Apache-2.0).

## Summary

90 third-party modules:

- **Apache-2.0**: 41
- **MIT**: 26
- **BSD-3-Clause**: 21
- **BSD-2-Clause**: 1
- **ISC**: 1

## Modules

| Module | Version | License | License source |
|--------|---------|---------|----------------|
| `github.com/beorn7/perks/quantile` | v1.0.1 | MIT | https://github.com/beorn7/perks/blob/v1.0.1/LICENSE |
| `github.com/cert-manager/cert-manager/pkg` | v1.21.1 | Apache-2.0 | https://github.com/cert-manager/cert-manager/blob/v1.21.1/LICENSE |
| `github.com/cespare/xxhash/v2` | v2.3.0 | MIT | https://github.com/cespare/xxhash/blob/v2.3.0/LICENSE.txt |
| `github.com/davecgh/go-spew/spew` | d8f796af33cc | ISC | https://github.com/davecgh/go-spew/blob/d8f796af33cc/LICENSE |
| `github.com/emicklei/go-restful/v3` | v3.13.0 | MIT | https://github.com/emicklei/go-restful/blob/v3.13.0/LICENSE |
| `github.com/evanphx/json-patch/v5` | v5.9.11 | BSD-3-Clause | https://github.com/evanphx/json-patch/blob/v5.9.11/v5/LICENSE |
| `github.com/fsnotify/fsnotify` | v1.9.0 | BSD-3-Clause | https://github.com/fsnotify/fsnotify/blob/v1.9.0/LICENSE |
| `github.com/fxamacker/cbor/v2` | v2.9.1 | MIT | https://github.com/fxamacker/cbor/blob/v2.9.1/LICENSE |
| `github.com/gabriel-vasile/mimetype` | v1.4.8 | MIT | https://github.com/gabriel-vasile/mimetype/blob/v1.4.8/LICENSE |
| `github.com/go-logr/logr` | v1.4.3 | Apache-2.0 | https://github.com/go-logr/logr/blob/v1.4.3/LICENSE |
| `github.com/go-logr/zapr` | v1.3.0 | Apache-2.0 | https://github.com/go-logr/zapr/blob/v1.3.0/LICENSE |
| `github.com/go-openapi/jsonpointer` | v0.23.1 | Apache-2.0 | https://github.com/go-openapi/jsonpointer/blob/v0.23.1/LICENSE |
| `github.com/go-openapi/jsonreference` | v0.21.5 | Apache-2.0 | https://github.com/go-openapi/jsonreference/blob/v0.21.5/LICENSE |
| `github.com/go-openapi/swag` | v0.26.0 | Apache-2.0 | https://github.com/go-openapi/swag/blob/v0.26.0/LICENSE |
| `github.com/go-openapi/swag/cmdutils` | cmdutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/cmdutils/v0.26.0/cmdutils/LICENSE |
| `github.com/go-openapi/swag/conv` | conv | Apache-2.0 | https://github.com/go-openapi/swag/blob/conv/v0.26.0/conv/LICENSE |
| `github.com/go-openapi/swag/fileutils` | fileutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/fileutils/v0.26.0/fileutils/LICENSE |
| `github.com/go-openapi/swag/jsonname` | jsonname | Apache-2.0 | https://github.com/go-openapi/swag/blob/jsonname/v0.26.0/jsonname/LICENSE |
| `github.com/go-openapi/swag/jsonutils` | jsonutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/jsonutils/v0.26.0/jsonutils/LICENSE |
| `github.com/go-openapi/swag/loading` | loading | Apache-2.0 | https://github.com/go-openapi/swag/blob/loading/v0.26.0/loading/LICENSE |
| `github.com/go-openapi/swag/mangling` | mangling | Apache-2.0 | https://github.com/go-openapi/swag/blob/mangling/v0.26.0/mangling/LICENSE |
| `github.com/go-openapi/swag/netutils` | netutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/netutils/v0.26.0/netutils/LICENSE |
| `github.com/go-openapi/swag/stringutils` | stringutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/stringutils/v0.26.0/stringutils/LICENSE |
| `github.com/go-openapi/swag/typeutils` | typeutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/typeutils/v0.26.0/typeutils/LICENSE |
| `github.com/go-openapi/swag/yamlutils` | yamlutils | Apache-2.0 | https://github.com/go-openapi/swag/blob/yamlutils/v0.26.0/yamlutils/LICENSE |
| `github.com/go-playground/locales` | v0.14.1 | MIT | https://github.com/go-playground/locales/blob/v0.14.1/LICENSE |
| `github.com/go-playground/universal-translator` | v0.18.1 | MIT | https://github.com/go-playground/universal-translator/blob/v0.18.1/LICENSE |
| `github.com/go-playground/validator/v10` | v10.26.0 | MIT | https://github.com/go-playground/validator/blob/v10.26.0/LICENSE |
| `github.com/go-viper/mapstructure/v2` | v2.4.0 | MIT | https://github.com/go-viper/mapstructure/blob/v2.4.0/LICENSE |
| `github.com/google/gnostic-models` | v0.7.1 | Apache-2.0 | https://github.com/google/gnostic-models/blob/v0.7.1/LICENSE |
| `github.com/google/uuid` | v1.6.0 | BSD-3-Clause | https://github.com/google/uuid/blob/v1.6.0/LICENSE |
| `github.com/json-iterator/go` | v1.1.12 | MIT | https://github.com/json-iterator/go/blob/v1.1.12/LICENSE |
| `github.com/leodido/go-urn` | v1.4.0 | MIT | https://github.com/leodido/go-urn/blob/v1.4.0/LICENSE |
| `github.com/mattn/go-colorable` | v0.1.13 | MIT | https://github.com/mattn/go-colorable/blob/v0.1.13/LICENSE |
| `github.com/mattn/go-isatty` | v0.0.19 | MIT | https://github.com/mattn/go-isatty/blob/v0.0.19/LICENSE |
| `github.com/mitchellh/mapstructure` | v1.5.0 | MIT | https://github.com/mitchellh/mapstructure/blob/v1.5.0/LICENSE |
| `github.com/modern-go/concurrent` | bacd9c7ef1dd | Apache-2.0 | https://github.com/modern-go/concurrent/blob/bacd9c7ef1dd/LICENSE |
| `github.com/modern-go/reflect2` | 35a7c28c31ee | Apache-2.0 | https://github.com/modern-go/reflect2/blob/35a7c28c31ee/LICENSE |
| `github.com/munnerz/goautoneg` | a7dc8b61c822 | BSD-3-Clause | https://github.com/munnerz/goautoneg/blob/a7dc8b61c822/LICENSE |
| `github.com/pelletier/go-toml/v2` | v2.2.3 | MIT | https://github.com/pelletier/go-toml/blob/v2.2.3/LICENSE |
| `github.com/pkg/errors` | v0.9.1 | BSD-2-Clause | https://github.com/pkg/errors/blob/v0.9.1/LICENSE |
| `github.com/pmezard/go-difflib/difflib` | 5d4384ee4fb2 | BSD-3-Clause | https://github.com/pmezard/go-difflib/blob/5d4384ee4fb2/LICENSE |
| `github.com/prometheus/client_golang/internal/github.com/golang/gddo/httputil` | v1.23.2 | BSD-3-Clause | https://github.com/prometheus/client_golang/blob/v1.23.2/internal/github.com/golang/gddo/LICENSE |
| `github.com/prometheus/client_golang/prometheus` | v1.23.2 | Apache-2.0 | https://github.com/prometheus/client_golang/blob/v1.23.2/LICENSE |
| `github.com/prometheus/client_model/go` | v0.6.2 | Apache-2.0 | https://github.com/prometheus/client_model/blob/v0.6.2/LICENSE |
| `github.com/prometheus/common` | v0.67.5 | Apache-2.0 | https://github.com/prometheus/common/blob/v0.67.5/LICENSE |
| `github.com/rs/zerolog` | v1.31.0 | MIT | https://github.com/rs/zerolog/blob/v1.31.0/LICENSE |
| `github.com/sagikazarmark/locafero` | v0.7.0 | MIT | https://github.com/sagikazarmark/locafero/blob/v0.7.0/LICENSE |
| `github.com/sourcegraph/conc` | v0.3.0 | MIT | https://github.com/sourcegraph/conc/blob/v0.3.0/LICENSE |
| `github.com/spf13/afero` | v1.12.0 | Apache-2.0 | https://github.com/spf13/afero/blob/v1.12.0/LICENSE.txt |
| `github.com/spf13/cast` | v1.7.1 | MIT | https://github.com/spf13/cast/blob/v1.7.1/LICENSE |
| `github.com/spf13/pflag` | v1.0.10 | BSD-3-Clause | https://github.com/spf13/pflag/blob/v1.0.10/LICENSE |
| `github.com/spf13/viper` | v1.20.1 | MIT | https://github.com/spf13/viper/blob/v1.20.1/LICENSE |
| `github.com/subosito/gotenv` | v1.6.0 | MIT | https://github.com/subosito/gotenv/blob/v1.6.0/LICENSE |
| `github.com/x448/float16` | v0.8.4 | MIT | https://github.com/x448/float16/blob/v0.8.4/LICENSE |
| `go.uber.org/multierr` | v1.11.0 | MIT | https://github.com/uber-go/multierr/blob/v1.11.0/LICENSE.txt |
| `go.uber.org/zap` | v1.28.0 | MIT | https://github.com/uber-go/zap/blob/v1.28.0/LICENSE |
| `go.yaml.in/yaml/v2` | v2.4.4 | Apache-2.0 | https://github.com/yaml/go-yaml/blob/v2.4.4/LICENSE |
| `go.yaml.in/yaml/v3` | v3.0.4 | MIT | https://github.com/yaml/go-yaml/blob/v3.0.4/LICENSE |
| `golang.org/x/crypto/sha3` | — | BSD-3-Clause | https://cs.opensource.google/go/x/crypto/+/v0.53.0:LICENSE |
| `golang.org/x/net` | — | BSD-3-Clause | https://cs.opensource.google/go/x/net/+/v0.56.0:LICENSE |
| `golang.org/x/oauth2` | — | BSD-3-Clause | https://cs.opensource.google/go/x/oauth2/+/v0.36.0:LICENSE |
| `golang.org/x/sync/errgroup` | — | BSD-3-Clause | https://cs.opensource.google/go/x/sync/+/v0.22.0:LICENSE |
| `golang.org/x/sys` | — | BSD-3-Clause | https://cs.opensource.google/go/x/sys/+/v0.46.0:LICENSE |
| `golang.org/x/term` | — | BSD-3-Clause | https://cs.opensource.google/go/x/term/+/v0.44.0:LICENSE |
| `golang.org/x/text` | — | BSD-3-Clause | https://cs.opensource.google/go/x/text/+/v0.40.0:LICENSE |
| `golang.org/x/time/rate` | — | BSD-3-Clause | https://cs.opensource.google/go/x/time/+/v0.15.0:LICENSE |
| `gomodules.xyz/jsonpatch/v2` | v2.5.0 | Apache-2.0 | https://github.com/gomodules/jsonpatch/blob/v2.5.0/v2/LICENSE |
| `google.golang.org/protobuf` | f2248ac996af | BSD-3-Clause | https://github.com/protocolbuffers/protobuf-go/blob/f2248ac996af/LICENSE |
| `gopkg.in/evanphx/json-patch.v4` | v4.13.0 | BSD-3-Clause | https://github.com/evanphx/json-patch/blob/v4.13.0/LICENSE |
| `gopkg.in/inf.v0` | v0.9.1 | BSD-3-Clause | https://github.com/go-inf/inf/blob/v0.9.1/LICENSE |
| `gopkg.in/yaml.v3` | v3.0.1 | MIT | https://github.com/go-yaml/yaml/blob/v3.0.1/LICENSE |
| `k8s.io/api` | v0.36.4 | Apache-2.0 | https://github.com/kubernetes/api/blob/v0.36.4/LICENSE |
| `k8s.io/apiextensions-apiserver/pkg/apis/apiextensions` | v0.36.2 | Apache-2.0 | https://github.com/kubernetes/apiextensions-apiserver/blob/v0.36.2/LICENSE |
| `k8s.io/apimachinery/pkg` | v0.36.4 | Apache-2.0 | https://github.com/kubernetes/apimachinery/blob/v0.36.4/LICENSE |
| `k8s.io/apimachinery/third_party/forked/golang` | v0.36.4 | BSD-3-Clause | https://github.com/kubernetes/apimachinery/blob/v0.36.4/third_party/forked/golang/LICENSE |
| `k8s.io/client-go` | v0.36.4 | Apache-2.0 | https://github.com/kubernetes/client-go/blob/v0.36.4/LICENSE |
| `k8s.io/klog/v2` | v2.140.0 | Apache-2.0 | https://github.com/kubernetes/klog/blob/v2.140.0/LICENSE |
| `k8s.io/kube-openapi/pkg` | 927ab1f70cd6 | Apache-2.0 | https://github.com/kubernetes/kube-openapi/blob/927ab1f70cd6/LICENSE |
| `k8s.io/kube-openapi/pkg/internal/third_party/go-json-experiment/json` | 927ab1f70cd6 | BSD-3-Clause | https://github.com/kubernetes/kube-openapi/blob/927ab1f70cd6/pkg/internal/third_party/go-json-experiment/json/LICENSE |
| `k8s.io/kube-openapi/pkg/validation/spec` | 927ab1f70cd6 | Apache-2.0 | https://github.com/kubernetes/kube-openapi/blob/927ab1f70cd6/pkg/validation/spec/LICENSE |
| `k8s.io/utils` | be93311217bd | Apache-2.0 | https://github.com/kubernetes/utils/blob/be93311217bd/LICENSE |
| `k8s.io/utils/internal/third_party/forked/golang` | be93311217bd | BSD-3-Clause | https://github.com/kubernetes/utils/blob/be93311217bd/internal/third_party/forked/golang/LICENSE |
| `k8s.io/utils/third_party/forked/golang/btree` | be93311217bd | Apache-2.0 | https://github.com/kubernetes/utils/blob/be93311217bd/third_party/forked/golang/btree/LICENSE |
| `sigs.k8s.io/controller-runtime` | v0.24.1 | Apache-2.0 | https://github.com/kubernetes-sigs/controller-runtime/blob/v0.24.1/LICENSE |
| `sigs.k8s.io/gateway-api/apis/v1` | v1.6.0 | Apache-2.0 | https://github.com/kubernetes-sigs/gateway-api/blob/v1.6.0/LICENSE |
| `sigs.k8s.io/json` | 2d320260d730 | Apache-2.0 | https://github.com/kubernetes-sigs/json/blob/2d320260d730/LICENSE |
| `sigs.k8s.io/randfill` | v1.0.0 | Apache-2.0 | https://github.com/kubernetes-sigs/randfill/blob/v1.0.0/LICENSE |
| `sigs.k8s.io/structured-merge-diff/v6` | v6.4.0 | Apache-2.0 | https://github.com/kubernetes-sigs/structured-merge-diff/blob/v6.4.0/LICENSE |
| `sigs.k8s.io/yaml` | v1.6.0 | Apache-2.0 | https://github.com/kubernetes-sigs/yaml/blob/v1.6.0/LICENSE |
