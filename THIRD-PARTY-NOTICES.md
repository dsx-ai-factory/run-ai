# Third-Party Notices

This product, **Run.ai** (`dsx-ai-factory/run-ai`), includes third-party open-source
software. The components below are copied into this repository ("vendored") in
first-party paths. Each is redistributed under its own license, and each vendored
tree retains its upstream `LICENSE` file and per-file copyright headers in place;
this file consolidates that attribution for convenience.

All vendored third-party components are under **permissive** licenses (MIT / Apache
License 2.0). None impose source-disclosure (copyleft) obligations.

If you find a component that is not listed here, the authoritative attribution is
the `LICENSE` file and/or per-file copyright header shipped alongside the code.

---

## Components

| Component | License | Copyright holder | Upstream | In-repo location(s) |
|---|---|---|---|---|
| ozzo-validation | MIT | Qiang Xue | https://github.com/go-ozzo/ozzo-validation | `runai-application-packages/validation/`, `runai-cluster/cluster/workload-controller/third_party/validation/` |
| conjungo | MIT | InVision | https://github.com/InVisionApp/conjungo | `runai-application-packages/jsonutils/conjungo/` |
| gojsonq | MIT | Saddam H. | https://github.com/thedevsaddam/gojsonq | `runai-application-packages/jsonutils/gojsonq/` |
| go-acc | Apache-2.0 | ORY (Aeneas Rekkas) | https://github.com/ory/go-acc | `control-plane/assets-service/third_party/go_acc/` |
| goofys | Apache-2.0 | Ka-Hing Cheung | https://github.com/kahing/goofys | `runai-cluster/cluster/goofys/` |
| KServe | Apache-2.0 | The KServe Authors | https://github.com/kserve/kserve | `control-plane/workloads-and-policies/third_party/kserve/` |
| Kubeflow MPI Operator (API types) | Apache-2.0 | The Kubeflow Authors | https://github.com/kubeflow/mpi-operator | `runai-cluster/cluster/workload-controller/pkg/api/mpijob/v2beta1/` |
| Seldon Core (API types) | Apache-2.0 | The Seldon Authors | https://github.com/SeldonIO/seldon-core | `runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/seldon/` |
| Kubernetes / kubebuilder scaffolding | Apache-2.0 | The Kubernetes Authors | https://github.com/kubernetes-sigs/controller-runtime | `runai-cluster/cluster/project-controller/hack/`, `runai-cluster/cluster/sdk/hack/`, and Kubernetes code-generator boilerplate in generated `zz_generated.deepcopy.go` files |
| Roboto (font) | Apache-2.0 | Google Inc. | https://github.com/googlefonts/roboto | `control-plane/frontend/public/static-cli/fonts/` |

Notes:
- **ozzo-validation** is present in two vendored copies; both are the same upstream project.
- **goofys** additionally contains RunAI-authored extensions (e.g. a GCS backend under
  `internal/backend_gcs.go`) that are derivative works within the Apache-2.0 tree; the
  goofys `LICENSE.md` governs the tree.
- **KServe** and **Seldon Core**: only a subset of upstream API type definitions are
  vendored (under `third_party/` / `thirdpartyapi/`). RunAI-authored type stubs in the
  same `thirdpartyapi/` area (e.g. `nim/`, `nvcf/`, `argo/`, `ray/`) are first-party and
  are **not** third-party components.
- The full license text for each component is retained in its in-repo `LICENSE`
  (or `LICENSE.md` / `LICENSE.txt`) file at the location(s) above. Apache-2.0
  components: see any of those files, or https://www.apache.org/licenses/LICENSE-2.0.

---

## MIT License

The following components are licensed under the MIT License:

- ozzo-validation — Copyright (c) 2016 Qiang Xue
- conjungo — Copyright (c) 2017 InVision
- gojsonq — Copyright (c) 2018 Saddam H. <thedevsaddam@gmail.com>

```
Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

## Apache License 2.0

The following components are licensed under the Apache License, Version 2.0:

- go-acc — Copyright ORY (Aeneas Rekkas)
- goofys — Copyright Ka-Hing Cheung
- KServe — Copyright The KServe Authors
- Kubeflow MPI Operator — Copyright The Kubeflow Authors
- Seldon Core — Copyright The Seldon Authors
- Kubernetes / kubebuilder scaffolding — Copyright The Kubernetes Authors
- Roboto (font) — Copyright Google Inc.

The full text of the Apache License, Version 2.0 is retained in-repo alongside each
component (see the In-repo location(s) column above) and is available at
https://www.apache.org/licenses/LICENSE-2.0. In compliance with Apache-2.0 §4(a)–(d),
the upstream `LICENSE`, `NOTICE` (where present), and per-file copyright/attribution
notices are retained unmodified in each vendored tree.
