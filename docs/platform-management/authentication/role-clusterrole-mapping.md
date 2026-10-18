# Role to ClusterRole Mapping

NVIDIA Run:ai maps each predefined role to Kubernetes ClusterRoles. Users who [authenticate to the cluster](../../installation-and-configuration/configuration/cluster-authentication.md) with their NVIDIA Run:ai identity and access it with `kubectl` receive these Kubernetes permissions in addition to the permissions that the role grants in the NVIDIA Run:ai platform. Use this page to see which ClusterRoles each role maps to and what each ClusterRole grants. For example, share it with your security team when reviewing access.

<Note>

**Note**

This page covers the predefined roles listed in [Roles](roles.md). It does not cover legacy roles or roles used internally by the platform.

</Note>

## How Roles Map to ClusterRoles

When an access rule assigns a role to a subject, NVIDIA Run:ai binds the subject to up to five ClusterRoles. In the table, `<role>` is the name of the role, for example `ai-practitioner`.

| ClusterRole | Bound | Applies to |
| --- | --- | --- |
| `runai-<role>` | Cluster-wide, with a ClusterRoleBinding | Every access rule, regardless of scope |
| `runai-<role>-ns` | In each project namespace in the scope of the access rule, with a RoleBinding | Every access rule |
| `runai-<role>-runai` | In the `runai` namespace, with a RoleBinding | Every access rule |
| `runai-<role>-runai-cs` | In the `runai` namespace, with a RoleBinding | Access rules scoped to the cluster or the account. This ClusterRole exists only for the System administrator and Data and storage administrator roles. |
| `runai-system-administrator-logs` | In the `runai-backend`, `runai-reservation`, `training-operator`, `knative`, `gpu-operator`, and `nim-operator` namespaces, with a RoleBinding | System administrator only |

The `runai-<role>-ns` ClusterRoles are aggregated ClusterRoles. Their rules combine the rules listed on this page with one additional ClusterRole, `runai-additional-workload-gvks-read` or `runai-additional-workload-gvks-crud`. NVIDIA Run:ai creates this additional ClusterRole at runtime, and it grants read or full access to every [external workload type](../../workloads-in-nvidia-run-ai/workload-types/extending-workload-support/defining-a-karta.md) registered in the cluster.

<Note>

**Note**

The `runai-system-administrator-logs` RoleBindings are created only in namespaces that already exist when NVIDIA Run:ai is installed or upgraded. If you install a component such as the GPU Operator after NVIDIA Run:ai, run a Helm upgrade of the NVIDIA Run:ai cluster chart to add the binding for the new namespace. The bindings are also not created when the chart is rendered with `helm template`, for example by GitOps tools such as Argo CD.

</Note>

## ClusterRoles per Role

| Role | ClusterRoles |
| --- | --- |
| System administrator | `runai-system-administrator`, `runai-system-administrator-ns`, `runai-system-administrator-runai`, `runai-system-administrator-runai-cs`, `runai-system-administrator-logs` |
| Department administrator | `runai-department-administrator`, `runai-department-administrator-ns`, `runai-department-administrator-runai` |
| Department viewer | `runai-department-viewer`, `runai-department-viewer-ns`, `runai-department-viewer-runai` |
| Project administrator | `runai-project-administrator`, `runai-project-administrator-ns`, `runai-project-administrator-runai` |
| AI practitioner | `runai-ai-practitioner`, `runai-ai-practitioner-ns`, `runai-ai-practitioner-runai` |
| Data and storage administrator | `runai-data-and-storage-administrator`, `runai-data-and-storage-administrator-ns`, `runai-data-and-storage-administrator-runai`, `runai-data-and-storage-administrator-runai-cs` |
| Editor | `runai-editor`, `runai-editor-ns`, `runai-editor-runai` |
| Viewer | `runai-viewer`, `runai-viewer-ns`, `runai-viewer-runai` |

## Permissions per ClusterRole

Expand a role to see the API groups, resources, and verbs that each of its ClusterRoles grants.

To view the permissions on your own cluster, run:

```bash
kubectl describe clusterrole runai-<role>
```

<Accordion title="System administrator">

**`runai-system-administrator`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `projects`, `nodepools`, `accessrules`, `departments`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |
| `core` | `namespaces` | `get`, `list`, `watch` |

**`runai-system-administrator-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core`, `extensions`, `networking.k8s.io`, `route.openshift.io` | `configmaps`, `services`, `ingresses`, `routes`, `routes/custom-host` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `secrets` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

This ClusterRole also aggregates `runai-additional-workload-gvks-crud`.

**`runai-system-administrator-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `configmaps` (names: `runai-cli-index`) | `update`, `patch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |
| `core` | `pods` | `list`, `get` |
| `core` | `pods/log` | `get` |

**`runai-system-administrator-runai-cs`**

Bound with a RoleBinding in the `runai` namespace, only for access rules scoped to the cluster or the account.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `secrets` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

**`runai-system-administrator-logs`**

Bound with a RoleBinding in the `runai-backend`, `runai-reservation`, `training-operator`, `knative`, `gpu-operator`, and `nim-operator` namespaces. Each binding is created only if the namespace existed when NVIDIA Run:ai was installed or upgraded.

Used by the `runai diagnostics collect-logs` command, which collects pod logs from these namespaces and from the `runai` namespace for NVIDIA Run:ai support.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `pods` | `list`, `get` |
| `core` | `pods/log` | `get` |

</Accordion>

<Accordion title="Department administrator">

**`runai-department-administrator`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `accessrules`, `nodepools`, `projects`, `departments`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-department-administrator-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core`, `extensions`, `networking.k8s.io`, `route.openshift.io` | `configmaps`, `services`, `ingresses`, `routes`, `routes/custom-host` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `secrets` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

This ClusterRole also aggregates `runai-additional-workload-gvks-crud`.

**`runai-department-administrator-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `configmaps` (names: `runai-cli-index`) | `update`, `patch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

</Accordion>

<Accordion title="Department viewer">

**`runai-department-viewer`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `nodepools`, `projects`, `departments`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-department-viewer-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies` | `get`, `list`, `watch` |

This ClusterRole also aggregates `runai-additional-workload-gvks-read`.

**`runai-department-viewer-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

</Accordion>

<Accordion title="Project administrator">

**`runai-project-administrator`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `accessrules`, `nodepools`, `projects`, `departments`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-project-administrator-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core`, `extensions`, `networking.k8s.io`, `route.openshift.io` | `configmaps`, `services`, `ingresses`, `routes`, `routes/custom-host` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `secrets` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

This ClusterRole also aggregates `runai-additional-workload-gvks-crud`.

**`runai-project-administrator-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `configmaps` (names: `runai-cli-index`) | `update`, `patch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

</Accordion>

<Accordion title="AI practitioner">

**`runai-ai-practitioner`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `projects`, `nodepools`, `departments`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-ai-practitioner-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core`, `extensions`, `networking.k8s.io`, `route.openshift.io` | `configmaps`, `services`, `ingresses`, `routes`, `routes/custom-host` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch` |
| `run.ai` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies` | `get`, `list`, `watch` |
| `core` | `secrets` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

This ClusterRole also aggregates `runai-additional-workload-gvks-crud`.

**`runai-ai-practitioner-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `configmaps` (names: `runai-cli-index`) | `update`, `patch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

</Accordion>

<Accordion title="Data and storage administrator">

**`runai-data-and-storage-administrator`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `departments`, `projects`, `datavolumes` | `get`, `list` |
| `kai.resources` | `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-data-and-storage-administrator-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies` | `get`, `list`, `watch` |
| `core` | `secrets` | `get`, `list`, `watch` |

This ClusterRole also aggregates `runai-additional-workload-gvks-read`.

**`runai-data-and-storage-administrator-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

**`runai-data-and-storage-administrator-runai-cs`**

Bound with a RoleBinding in the `runai` namespace, only for access rules scoped to the cluster or the account.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

</Accordion>

<Accordion title="Editor">

**`runai-editor`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `projects`, `departments`, `nodepools`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-editor-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core`, `extensions`, `networking.k8s.io`, `route.openshift.io` | `configmaps`, `services`, `ingresses`, `routes`, `routes/custom-host` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies` | `get`, `list`, `watch` |
| `core` | `secrets` | `get`, `list`, `watch`, `create`, `update`, `patch`, `delete` |

This ClusterRole also aggregates `runai-additional-workload-gvks-crud`.

**`runai-editor-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `core` | `configmaps` (names: `runai-cli-index`) | `update`, `patch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

</Accordion>

<Accordion title="Viewer">

**`runai-viewer`**

Bound with a ClusterRoleBinding, cluster-wide, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai` | `projects`, `nodepools`, `departments`, `datavolumes` | `get`, `list` |
| `kai.resources` | `nodepools`, `projects`, `departments` | `get`, `list` |
| `core`, `run.ai`, `storage.k8s.io` | `namespaces`, `nodes`, `services/proxy`, `projects`, `nodepools`, `storageclasses` | `get`, `list`, `watch` |

**`runai-viewer-ns`**

Bound with a RoleBinding in each project namespace in the scope of the access rule, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai`, `scheduling.run.ai`, `batch`, `kubeflow.org`, `apps`, `workspace.devfile.io`, `ray.io`, `machinelearning.seldon.io`, `kubevirt.io`, `amlarc.azureml.com`, `serving.knative.dev`, `knative.dev`, `argoproj.io`, `tekton.dev`, `core` | `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads`, `jobs`, `mpijobs`, `tfjobs`, `pytorchjobs`, `xgboostjobs`, `statefulsets`, `replicasets`, `deployments`, `devworkspaces`, `rayclusters`, `rayjobs`, `rayservices`, `notebooks`, `seldondeployments`, `virtualmachines`, `virtualmachineinstances`, `amljobs`, `revisions`, `configurations`, `services`, `cronjobs`, `scheduledworkflows`, `workflows`, `pipelineruns`, `taskruns`, `pods`, `pods/log`, `pods/exec`, `pods/attach`, `pods/portforward` | `get`, `list`, `watch` |
| `run.ai`, `scheduling.run.ai` | `runaijobs`, `podgroups` | `get`, `list`, `watch` |
| `core`, `discovery.k8s.io`, `route.openshift.io`, `extensions`, `networking.k8s.io` | `endpoints`, `events`, `endpointslices`, `routes`, `configmaps`, `services`, `ingresses`, `routes/custom-host` | `get`, `list`, `watch` |
| `core` | `pods/log` | `watch` |
| `core` | `persistentvolumeclaims` | `get`, `list`, `watch` |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies` | `get`, `list`, `watch` |
| `core` | `secrets` | `get`, `list`, `watch` |

This ClusterRole also aggregates `runai-additional-workload-gvks-read`.

**`runai-viewer-runai`**

Bound with a RoleBinding in the `runai` namespace, for every access rule.

| API groups | Resources | Verbs |
| --- | --- | --- |
| `run.ai` | `interactivepolicies`, `trainingpolicies`, `inferencepolicies`, `distributedinferencepolicies`, `distributedpolicies`, `interactiveworkloads`, `trainingworkloads`, `inferenceworkloads`, `distributedinferenceworkloads`, `distributedinferenceworkloads/status`, `distributedinferenceworkloads/scale`, `distributedworkloads` | `get`, `list`, `watch` |

</Accordion>
