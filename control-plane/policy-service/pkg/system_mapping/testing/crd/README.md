# Schema fixtures for system-mapping Layer 4 validation

This directory holds OpenAPI v3 / CRD documents used to validate manifests
produced by Layer 3 of the system-mapping test plan against the same schemas
the Kubernetes API server uses at admission time.

## Contents

- `apps_v1_openapi.json` — OpenAPI v3 spec for the `apps/v1` API group.
- `batch_v1_openapi.json` — OpenAPI v3 spec for the `batch/v1` API group.

These files are downloaded from a **live Kubernetes cluster** (not from static
GitHub files) because K8s generates enum constraints dynamically at runtime.
The static files in the kubernetes/kubernetes repo lack enum constraints —
they're added by kube-openapi based on Go type annotations.

The current schemas were downloaded from a **v1.33** cluster.

Each document is self-contained: every `$ref` in the file resolves to a
schema declared under `components.schemas`, including the cross-group
dependencies (e.g. `core/v1.PodSpec`, `apimachinery/pkg/apis/meta/v1.ObjectMeta`).

## Refreshing

Download from a live cluster (recommended — includes enum constraints):

```sh
kubectl get --raw /openapi/v3/apis/apps/v1 > apps_v1_openapi.json
kubectl get --raw /openapi/v3/apis/batch/v1 > batch_v1_openapi.json
```

Note: The static GitHub files do NOT include enum constraints:
```sh
# DON'T USE - missing enums:
# curl -fsSL https://raw.githubusercontent.com/kubernetes/kubernetes/release-1.33/api/openapi-spec/v3/apis__apps__v1_openapi.json
```

## Adding new workload types

Any type registered with the K8s API server exposes an OpenAPI schema. To add
validation for a new workload type (CRD or built-in):

1. **Install the CRD on a cluster** (skip for built-in types):
   ```sh
   kubectl apply -f <crd.yaml>
   ```

2. **Download the OpenAPI schema**:
   ```sh
   # Built-in types
   kubectl get --raw /openapi/v3/apis/<group>/<version> > <group>_<version>_openapi.json

   # Examples:
   kubectl get --raw /openapi/v3/apis/apps/v1 > apps_v1_openapi.json
   kubectl get --raw /openapi/v3/apis/ray.io/v1 > ray.io_v1_openapi.json
   kubectl get --raw /openapi/v3/apis/run.ai/v1 > run.ai_v1_openapi.json
   ```

3. **Add test config** in `system_mapping_layer3_test.go`:
   ```go
   var rayClusterConfig = gvkConfig{
       gvk:        api.Gvk{Group: "ray.io", Kind: "RayCluster", Version: "v1"},
       inputDir:   "testing/inputs/ray.io_RayCluster_v1",
       schemaPath: "testing/crd/ray.io_v1_openapi.json",
       schemaName: "io.ray.v1.RayCluster",  // check components.schemas in the JSON
       caseNames:  []string{"workload", "pod", "container"},
   }
   ```

4. **Create test inputs** under `testing/inputs/<group>_<Kind>_<version>/`:
   - `skeleton.yaml` — minimal valid manifest
   - `reference.yaml` — comprehensive manifest for Layer 1/2
   - `cases/<name>/policy.yaml` + `golden.yaml` — Layer 3 test cases

**Note**: The CRD must be installed on the cluster to download its OpenAPI
schema. The API server generates schemas dynamically from installed CRDs.

**Alternative for CRDs** (no cluster needed): Extract the schema directly from
the CRD's `spec.versions[].schema.openAPIV3Schema` field and validate using
`k8s.io/apiextensions-apiserver/pkg/apiserver/validation`.
