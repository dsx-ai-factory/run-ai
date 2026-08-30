# Workload Controller Notes

This controller is responsible for validating and controlling workload and policies,
which we use to create resources such as runaijob.

Policies are set by the administrator, and contain default and rules. Workloads are set of values that the researchers want to 
use for creating the runaijob and other resources. A workload can inherit values from another workload.

## Directories
Project structure:

- config
    - crd: The CRD definitions, auto generated
    - samples: sample yaml for each CRD.
- hack: 
    - script for testing webhooks.
- pkg: 
    - api: The go structures representing the CRDs + webhook.
      - field_types: the building blocks of the CRD types
      - common: shared definitiona for any of our supported CRDs
      - jobs: common definition for both training + interactive jobs
      - interactivepolicy/workload - interactive jobs + webhook
      - samplepolicy/workload - a sample CRD, an example for you to debug and play with
    - builders: Generate the resources based on workload + policy.  
      - common: shared code for any of our supported builders
      - factory: builder constructors per builder type
      - jobs: builders for training and interactive jobs
      - sample: sample builder, an example for you to debug and play with 
    - controller: the reconcilers
    - enforcer: the merge and enforce logic.
    - utils: constants and common functions.
- third_party: Packages that we use for the code.
    - conjungo: For doing merge.
    - gojsonq: For JSON queries.
    - validation: Various validation functions.

## Makefiles

Necessary make when changing the schema:
```
make generate install all
```
Run linter on the code using `.golangcli.yml` as config:
```
make lint
```
Note that install really installs the CRD in k8s so it's mainly for local development machine.

Prior to generating schema, you need to install controller-gen by running
```
go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.11.1
```
And create symbolic link from bin/controller-gen (under workload-controller) to the controller-gen executable
(use which command to find out the path)

## Linter Commands
Arrange the import section for the project
```
goimports -local "github.com/GoogleContainerTools/skaffold" -w .
```

## Running Locally

Just run the controller's main using:
```
go build ./...
```
Possible command line options:
```
--zap-log-level <level>
```
The crd/hack directory contain python script to simulate injectioin of webhooks.

# Debug Levels

* V(7): detailed per-field merging info
* V(6): overall merge info
* V(5): resources creation info

# Adding New Fields

There are three kind of fields:
- A primitive field: containing string, number, int or bool value.
- An itemized field: containing a set of items, like volumes, pvcs, etc. 

## Primitive Fields

Add the field to the following struct with the relevant type and a JSON tag:
```
type WorkloadFields struct {
	...
	Something       *StringField `json:"something,omitempty"`
	...
}
```
If there are limitations that run:ai wants to impose on this field, add them to runai_rules.go.
Example:
```
runaiRulesFlds := &WorkloadFields{
		
    Name: &StringField{
		Rules: &StringRules{
    		Format: PtrMe(rules.FormatKubeName),
	    },
    },
    ...
```

## Itemized Fields
There are two type of itemized fields:
- Itemized fields whose value is primitive. Example: environment variable, whose value is a string.
- Itemized fields whose value is complex. Example: volume, whose value has multiple fields: source path, target path and read only.

Example for adding itemized field with primitive value:
```
type EnvVarField struct {
	Rules     *MapRules                       `json:"rules,omitempty"`
	ItemRules *StringRules                    `json:"itemRules,omitempty"`
	KeyRules  *StringRules                    `json:"keyRules,omitempty"`
	Value     *map[string]EnvVarsItemAndRules `json:"items,omitempty"`
}

type EnvVarsItemAndRules struct {
	Rules *MapItemRules `json:"rules,omitempty"`
	Deleted *bool   `json:"deleted,omitempty"`
	Value   *string `json:"value,omitempty"`
}
```
Notes:
1) KeyRules is required only if there are rules to apply on the "key" of each item (the environment 
   variable name in this case). Typically the key is just a correlation string without any rules,
   and in this case KeyRules can be omitted.
   
2) Rules (in both structures) and deletion flag are mandatory. Keep them the same as here.

3) Value and ItemRules depend on the type of value (string VS StringRules), 
   bool VS BoolRules, etc). Keep them in the same format for the validation mechanism
   to work properly.
   
Complex items are a bit more complicated:

```
type VolumesField struct {
	Rules     *MapRules                      `json:"rules,omitempty"`
	ItemRules *VolumeItemRules               `json:"itemRules,omitempty"`
	Value     *map[string]VolumeItemAndRules `json:"items,omitempty"`
}

type VolumeItemAndRules struct {
	Rules *MapItemRules    `json:"rules,omitempty"`
	Value *VolumeItemValue `json:"value,omitempty"`
	// +kubebuilder:default:=false
	Deleted *bool `json:"deleted,omitempty"`
}

type VolumeItemRules struct {
	SourcePath *StringRules `json:"sourcePath,omitempty"`
	TargetPath *StringRules `json:"targetPath,omitempty"`
	ReadOnly   *BoolRules   `json:"readOnly,omitempty"`
	NfsServer  *StringRules `json:"nfsServer,omitempty"`
}

type VolumeItemValue struct {

	SourcePath *string `json:"sourcePath,omitempty"`
	// +kubebuilder:validation:Required
	TargetPath *string `json:"targetPath"`

	ReadOnly *bool `json:"readOnly,omitempty"`

	NfsServer *string `json:"nfsServer,omitempty"`
}
```
The major differences are the VolumeItemRules and VolumeItemValue, 
which specify values and rules for the members of the field. Keep the same
general structure intact, as the validation mechanism assumes this structure.

## Validation Tags - **_Important_** 

Complex fields must be either required or have default value.
Note the VolumeItemValue above for example. 

# Pod Spec Conventions

## hostNetwork and dnsPolicy

When `hostNetwork: true` is set on a workload, the pod spec builder must also set
`dnsPolicy: ClusterFirstWithHostNet`. Without it, Kubernetes falls back to using the
node's DNS resolver instead of kube-dns, and Kubernetes service names won't resolve.

This applies to all builders that set `hostNetwork` on a pod spec:
- `pkg/builders/distributed/common/distributed_pod_spec.go` (distributed training)
- `pkg/builders/jobs/runaijob/pods.go` (training, workspace)

Reference: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy

## restartPolicy defaults

When the user doesn't specify `restartPolicy`, `pods.go` (`DetermineRestartPolicy`) defaults to:

- **Single-container pods → `Never`**
- **Multi-container pods → `OnFailure`** — without this, a sidecar (S3 sync, OAuth proxy) that
  isn't ready but hasn't failed leaves the pod stuck in init forever. `OnFailure` lets Kubernetes
  retry up to the backoff limit.

Does **not** apply to distributed workloads — there the framework (PyTorchJob, MPIJob, etc.)
controls restart behavior, so `distributed_pod_spec.go` only sets it when explicitly provided.

# History

The controller was created using following sample commands:
```
kubebuilder init --domain ""
kubebuilder create api --group run.ai --version v1 --kind WorkloadPolicy
kubebuilder create webhook --group run.ai --version v1 --kind WorkloadPolicy --programmatic-validation --validation
```
The clientset was actually a reuse of Daniel's agent controller with manual adaptation to our CRD.
