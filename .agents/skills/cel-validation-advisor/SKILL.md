---
name: cel-validation-advisor
description: Use when adding or registering any Kubernetes webhook (validating, mutating, or conversion), or writing, editing, or explicitly asking for admission validation, CRD XValidation rules, ValidatingAdmissionPolicy, webhook matchConditions, or CEL expressions in the Run:ai monorepo.
---

## TL;DR (the whole thing in 3 lines)

- Whenever adding a Kubernetes webhook, assess **CEL** before implementing it and offer a recommendation once, even if the conclusion is to keep the webhook. Also assess new or changed admission validation.
- Use CRD CEL for schema-owned constraints; VAP for built-in resources, request context, authorization, or named parameters; keep webhooks for arbitrary lookups and plugins.
- Every CEL feature you use must be **GA in Kubernetes ≤ 1.30** (chart floor `kubeVersion: '>=1.30.0-0'`). Advise during unrelated work; implement CEL when the user requests it.

## Load this skill

This file is the portable Markdown prompt: copy or attach the whole file, including the examples, to any coding agent. No agent-specific tools are required. Repo paths below are relative to the Run:ai checkout; in another repository, locate equivalent types and generators first.

| Agent | Setup from the repository root | Verify / invoke |
|---|---|---|
| Codex | Open the checkout containing `.agents/skills/cel-validation-advisor/SKILL.md`; Codex discovers this shared directory natively, including from subdirectories of the repository. No `.codex/skills/` copy is needed. | Start Codex in that checkout and ask it to use `cel-validation-advisor`; in CLI/IDE, run `/skills` to confirm discovery or invoke `$cel-validation-advisor` with the validation task. |
| Claude Code | The committed `.claude/skills/cel-validation-advisor` symlink points to `../../.agents/skills/cel-validation-advisor`. If distributing just this file, copy it to `.claude/skills/cel-validation-advisor/SKILL.md`. | Start `claude`; invoke `/cel-validation-advisor` with the validation task. |
| Gemini CLI | Keep this file at `.agents/skills/cel-validation-advisor/SKILL.md`; current Gemini CLI discovers that workspace alias. For versions without that alias, copy it to `.gemini/skills/cel-validation-advisor/SKILL.md`. | Start `gemini`, run `/skills reload` then `/skills list`, and ask it to use `cel-validation-advisor`. |
| Cursor | Open the checkout containing `.agents/skills/cel-validation-advisor/SKILL.md`. For versions without shared-directory discovery, copy it to `.cursor/skills/cel-validation-advisor/SKILL.md`. | Open Customize → Skills, confirm discovery, and ask Agent to use `cel-validation-advisor`. |
| GitHub Copilot in VS Code | Open the checkout containing `.agents/skills/cel-validation-advisor/SKILL.md`; current VS Code supports this shared project-skill directory. | In Chat, run `/skills` to confirm discovery, then invoke `/cel-validation-advisor` with the validation task. |
| Other agents or clients without shared skill discovery | Attach this Markdown file as context, or paste its contents with the request below. | Ask the agent to read the file before proposing validation changes. |

Keep one maintained copy; use the fallback copy only if discovery fails. If the client lacks skill support entirely, attaching the prompt still works, without automatic activation.

```text
Use cel-validation-advisor to implement the following validation: <requirement>.
Target: <resource kind, API version, and create/update/delete operations>.
Kubernetes minimum: 1.30. Preserve current accepted callers and bypasses.
Read the relevant Go types, generated schema, and existing validator first.
```

Setup references: [Codex](https://learn.chatgpt.com/docs/build-skills#where-codex-loads-local-skills), [Claude Code](https://code.claude.com/docs/en/skills), [Gemini CLI](https://geminicli.com/docs/cli/skills/), [Cursor](https://cursor.com/docs/skills), [GitHub Copilot in VS Code](https://code.visualstudio.com/docs/agent-customization/agent-skills).

## Guide map

- Use **DETECT → CLASSIFY → GATE** for advisory decisions.
- Use **Paired CRD examples** for existence, cross-field, transition, messages, lists, and strings.
- Use **VAP examples** for `oldObject`, ConfigMap parameters, and typed CRD parameters.
- Use **Run:ai examples** for workload fields and template filtering.
- Use **Best practices**, **Quick reference**, and **Limitations** while implementing.

## Behavior

1. For an explicit CEL question or implementation request, use the relevant guide sections directly. Whenever the task adds a Kubernetes webhook, run DETECT → CLASSIFY → GATE before implementing it; do not wait for validation code or an explicit request to use this skill. For other admission work, run the three steps silently.
2. For every new webhook, report the CEL assessment once: use the **Alert block** for an applicable opportunity, or briefly explain why the webhook must stay when CEL is unsuitable. For other admission work, print the **Alert block** only if a step yields an alert. Then continue the user's original task unchanged.
3. If the user already asked to implement CEL, proceed within that request using **Patterns**, **Rollout** and **Tests**. Otherwise offer the recommendation without changing the implementation. `go CEL` is one way to accept, not a required password.
4. Never repeat the assessment for the same webhook or validation in one session just because both its handler and registration are edited.

## Step 1: DETECT

Fire when the task requests a new Kubernetes webhook, or the code or configuration being written or edited matches a signal. Registration alone is enough, including Go, YAML, Helm templates, or generated configuration. Stay silent otherwise.

| Signal in task, code, or configuration | Meaning |
|---|---|
| request to add a Kubernetes webhook, even before its handler exists | assess CEL before implementation |
| `admission.Handler`, `CustomValidator`, `CustomDefaulter`, `ValidateCreate`/`ValidateUpdate`/`ValidateDelete`, `Default`, `admission.Denied(` in webhook code | new or changed admission webhook logic |
| `ServiceAccountValidator`, `HasPrefix(username, "system:serviceaccount:")`, `request.UserInfo`, `AdmissionRequest.UserInfo` | caller identity check |
| new entry or rule in a `ValidatingWebhookConfiguration` or `MutatingWebhookConfiguration`, or new webhook server registration (including init-ca `config_maps.go`) | registering an admission webhook, even without handler changes |
| new CRD conversion webhook (`spec.conversion.strategy: Webhook`) | assess and explain why conversion still needs a webhook |
| Go `if` checks on a CR: required field, mutually exclusive fields, immutable field, unique list items | object-shape check written in Go |
| `+kubebuilder:validation:XValidation`, `x-kubernetes-validations`, `admissionv1.ValidatingAdmissionPolicy{`, any CEL string | already CEL → run **Step 3** only |

**Silent during unrelated work** (explicit questions about these areas still get an answer):

- REST input validation in `control-plane/` services. It is not Kubernetes admission.
- Checks inside `Reconcile` that only set status or conditions.
- Existing webhook code that is read or reviewed but not changed.
- Non-Kubernetes webhooks, such as GitHub or payment callbacks.

## Step 2: CLASSIFY

Classify the complete check, including its dependencies. A username check plus an arbitrary lookup is not a username-only check. For a mutating webhook, assess any separate validation and invocation filtering, but keep mutation/defaulting in the webhook at the 1.30 floor; `MutatingAdmissionPolicy` is GA only in 1.36. Do not suggest validating CEL as a replacement for mutation or CRD version conversion.

| The check needs | Verdict | Concrete criterion / repo pattern |
|---|---|---|
| fields or transitions of a CRD whose schema we own | **CRD-CEL** `XValidation` | `Application.HelmSpec` chart source XOR; transition recipes below |
| fields of a built-in resource or a CRD whose schema we cannot change | **VAP** | Deployment replica limit; built-ins cannot acquire CRD markers |
| caller username or request context | **VAP** | application-controller identity policy; preserve the intended set of writers |
| an authorization decision for the caller | **VAP** with `authorizer` | supported authorization checks, not arbitrary Kubernetes reads |
| a known parameter object | **VAP** with `paramKind` + binding `paramRef` | WorkloadRunner policy uses a named ServiceAccount and compares `request.userInfo.uid` to `params.metadata.uid` |
| arbitrary `Get`/`List`, custom UID resolution, external APIs, or plugin execution | **webhook stays** | a `ServiceAccountValidator` callback may resolve data that cannot be represented by a fixed parameter |
| filter whether an admission webhook is called | webhook **`matchConditions`** | express the intended invocation conditions for a new webhook, or preserve the same checks when moving an established bypass |
| mutation or defaulting | **mutating webhook stays** at the 1.30 floor | validating CEL cannot change objects |
| CRD version conversion | **conversion webhook stays** | admission CEL does not convert between API versions |

For label-scoped VAP enforcement, put the selector on `Binding.Spec.MatchResources`. A user-editable opt-in label is not an authorization boundary.

**CEL has bounded inputs:** the admitted object, previous object, request context, and (for VAP) declared parameters and the authorizer. It does not offer arbitrary resource queries.

## Step 3: GATE

Run:ai floor: Kubernetes **1.30** (`runai-chart/Chart.yaml`, installer `minimumKubernetesVersion`, operator `isVAPSupported`). A feature is allowed only if it is GA at or below 1.30.

| Feature | GA | Verdict |
|---|---|---|
| `ValidatingAdmissionPolicy` v1: `variables`, `matchConditions`, `paramRef`, `messageExpression`, `auditAnnotations`, `authorizer` | 1.30 | ✓ allow |
| Webhook `matchConditions` | 1.30 | ✓ allow |
| CRD `x-kubernetes-validations`: `oldSelf`, `messageExpression`, `fieldPath`, `reason` | 1.29 | ✓ allow |
| CEL libraries: lists, regex, URL, authorizer, strings v1 | ≤ 1.28 | ✓ allow |
| CEL `quantity` library, optional types | 1.29 | ✓ allow |
| `optionalOldSelf`, CRD validation ratcheting (`CRDValidationRatcheting`) | 1.33 | ✗ alert |
| CEL `ip` / `cidr` libraries | 1.31 | ✗ alert |
| CEL `format` library | 1.32 | ✗ alert |
| CEL two-variable comprehensions | 1.33 | ✗ alert |
| CEL `semver` library | 1.34 | ✗ alert |
| `MutatingAdmissionPolicy` | 1.36 | ✗ alert, keep the mutating webhook |

Also alert when an expression can exceed the CEL cost budget: iterating a list without `maxItems`, a map without `maxProperties`, or a string without `maxLength`, including nested collections.

Sources: [CEL availability](https://kubernetes.io/docs/reference/using-api/cel/), [VAP](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/), [active feature gates](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/) and [graduated feature gates](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates-removed/). Re-verify these when the chart floor changes. Library introduction versions and API feature graduation are separate checks.

## Patterns (when CEL implementation is requested)

### VAP

1. Build one `admissionpolicy.Definition` per target resource: typed `Policy`, typed `Binding` whose `PolicyName` matches, optional `Enabled` callback.
2. Hand it to `admissionpolicy.NewResources` (resource-based operands) or `admissionpolicy.NewResourcesOld` (legacy `OperandOld` operands). The owning operand is the service that owns the webhook today. No new operand, no init-ca ConfigMap, no shared VAP, no `paramRef` to `RunaiConfig`.
3. For a username-based allowlist, identity is the complete username: `request.userInfo.username in variables.allowedServiceAccounts`, values shaped `system:serviceaccount:<installation-namespace>:<effective-service-account>`. Derive namespace and ServiceAccount name from `RunaiConfig`, including NMC overrides. Do not replace UID-based checks with usernames. The existing WorkloadRunner policy intentionally uses a ServiceAccount parameter UID; arbitrary custom UID resolvers remain in the webhook. Inventory all currently accepted writers before tightening a prefix/substring check to an exact allowlist, and make any behavior change explicit.
4. Both objects carry the label `run.ai/created-by: runai-operator`. Do not use `app`; reviewers reject it as too common.
5. `failurePolicy: Fail` unless the replaced webhook was `Ignore`.

### CRD-CEL

1. Add `// +kubebuilder:validation:XValidation:rule="...",message="..."` on the Go type or field in `*_types.go`. Escape reserved words: `self.__namespace__`.
2. Regenerate the CRD and update every shipped copy, including `runai-cluster/cluster/runai-crd-upgrader/crds/`.
3. Remove only the equivalent branch from the webhook. Keep the webhook if it still runs other checks.
4. If the webhook being replaced is disabled in some installations, adding a CRD rule enforces new behavior there. Say so in the alert and require an audit of stored objects before enabling.

## Paired CRD examples

These are **NEW teaching examples**, not shipped Run:ai APIs. Every validation pattern below appears in both the Go markers and the equivalent YAML `spec` schema. Use the owner's generator for real CRDs; do not hand-edit generated files. `XValidation` markers generate CRD rules, not VAP or webhook resources.

| Pattern | Go marker / YAML location | Example behavior |
|---|---|---|
| Field existence + `startsWith` | `ExampleSpec` / `spec.x-kubernetes-validations` | absent image allowed; a supplied image needs the example registry prefix |
| Cross-field validation | `ExampleSpec` / `spec.x-kubernetes-validations` | minimum cannot exceed maximum |
| Dynamic `messageExpression` | `ExampleSpec` / `spec.x-kubernetes-validations` | an invalid image reports its actual value |
| Transition with `oldSelf` | `StableID` / `spec.properties.stableID` | required identifier may be set on create but not changed on update |
| List `all` + string `matches` | `Members` / `spec.properties.members` | each name follows the pattern |
| List `map` + `exists` | `Members` / `spec.properties.members` | projected names include `worker` |

Go type fragment (place in a CRD API package and use `ExampleSpec` as its required `spec` field):

```go
// +kubebuilder:validation:XValidation:rule="!has(self.image) || self.image.startsWith('registry.example.com/')",message="image must use registry.example.com/",messageExpression="has(self.image) ? 'image ' + self.image + ' must use registry.example.com/' : 'image must use registry.example.com/'"
// +kubebuilder:validation:XValidation:rule="self.minReplicas <= self.maxReplicas",message="minReplicas must not exceed maxReplicas"
// +kubebuilder:validation:XValidation:rule="self.maxReplicas <= 10",message="maxReplicas must not exceed 10"
type ExampleSpec struct {
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Image *string `json:"image,omitempty"`
	// +kubebuilder:validation:Minimum=0
	MinReplicas int32 `json:"minReplicas"`
	// +kubebuilder:validation:Minimum=0
	MaxReplicas int32 `json:"maxReplicas"`
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="stableID is immutable"
	StableID string `json:"stableID"`
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:rule="self.all(m, m.name.matches('^[a-z][a-z0-9-]*$'))",message="member names must be lowercase identifiers"
	// +kubebuilder:validation:XValidation:rule="self.map(m, m.name).exists(n, n == 'worker')",message="a worker member is required"
	Members []ExampleMember `json:"members"`
}

type ExampleMember struct {
	// +kubebuilder:validation:MaxLength=32
	Name string `json:"name"`
}
```

Equivalent raw YAML: a complete CRD wrapping the matching `spec` schema. Generator metadata and descriptions are omitted; constraints, required fields, types, and rule placement match the Go fragment.

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: celrecipes.examples.run.ai
spec:
  group: examples.run.ai
  scope: Namespaced
  names:
    plural: celrecipes
    singular: celrecipe
    kind: CELRecipe
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required: [spec]
          properties:
            spec:
              type: object
              required: [minReplicas, maxReplicas, stableID, members]
              x-kubernetes-validations:
                - rule: "!has(self.image) || self.image.startsWith('registry.example.com/')"
                  message: image must use registry.example.com/
                  messageExpression: "has(self.image) ? 'image ' + self.image + ' must use registry.example.com/' : 'image must use registry.example.com/'"
                - rule: "self.minReplicas <= self.maxReplicas"
                  message: minReplicas must not exceed maxReplicas
                - rule: "self.maxReplicas <= 10"
                  message: maxReplicas must not exceed 10
              properties:
                image:
                  type: string
                  maxLength: 256
                minReplicas:
                  type: integer
                  format: int32
                  minimum: 0
                maxReplicas:
                  type: integer
                  format: int32
                  minimum: 0
                stableID:
                  type: string
                  maxLength: 64
                  x-kubernetes-validations:
                    - rule: self == oldSelf
                      message: stableID is immutable
                members:
                  type: array
                  maxItems: 16
                  items:
                    type: object
                    required: [name]
                    properties:
                      name:
                        type: string
                        maxLength: 32
                  x-kubernetes-validations:
                    - rule: "self.all(m, m.name.matches('^[a-z][a-z0-9-]*$'))"
                      message: member names must be lowercase identifiers
                    - rule: "self.map(m, m.name).exists(n, n == 'worker')"
                      message: a worker member is required
```

The CRD message uses the bounded image string so Kubernetes 1.30 can estimate concatenation cost. Numeric-to-string concatenation in a CRD `messageExpression` can fail its static cost check on this version; validate it on the target API server instead of assuming that a working VAP message is also accepted in a CRD.

A valid instance has `minReplicas: 1`, `maxReplicas: 3`, `stableID: batch-a`, and `members: [{name: worker}]`; image is optional. Reject `minReplicas: 4` with maximum 3, maximum 11, an empty member list, uppercase names, or an update changing `stableID`.

`oldSelf` is available without `optionalOldSelf` on Kubernetes 1.30. A transition rule normally runs only when both old and new values exist. The required field above therefore cannot be removed. To freeze the presence of an **optional** field, put the rule on its persistent parent:

```go
// +kubebuilder:validation:XValidation:rule="has(self.labelKey) == has(oldSelf.labelKey) && (!has(self.labelKey) || self.labelKey == oldSelf.labelKey)",message="labelKey presence and value are immutable"
type OptionalIdentitySpec struct {
	// +optional
	// +kubebuilder:validation:MaxLength=64
	LabelKey *string `json:"labelKey,omitempty"`
}
```

Equivalent schema fragment for that parent (make the parent required in the containing resource):

```yaml
type: object
properties:
  labelKey:
    type: string
    maxLength: 64
x-kubernetes-validations:
  - rule: "has(self.labelKey) == has(oldSelf.labelKey) && (!has(self.labelKey) || self.labelKey == oldSelf.labelKey)"
    message: labelKey presence and value are immutable
```

A cross-field transition can also use `oldSelf` on a parent object; it does not inherently require VAP. Transitions within lists require correlatable map-list entries, not atomic/set-list ancestry.

## VAP examples

These **NEW teaching policies** use built-in Deployments to make the examples independent of Run:ai installation. Save each complete YAML block separately and try it only in a disposable environment. They enforce `Deny` to make the expected outcomes observable; use the staged **Rollout** below for production migrations. Each binding limits enforcement to a labeled test namespace. Create that namespace and label it `cel-example: enabled` before testing. These examples cover the main Deployment resource; none covers the separate `deployments/scale` subresource.

### Transition with `oldObject`

This update-only policy rejects replica decreases, including a dynamic message. `oldObject` is null on CREATE, so a policy covering CREATE as well must guard the old-object access. Deployment replicas are defaulted by Kubernetes before validation. This is an example constraint, not a recommendation to prohibit scaling down in production.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: example-no-scale-down
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
      - apiGroups: [apps]
        apiVersions: [v1]
        operations: [UPDATE]
        resources: [deployments]
  validations:
    - expression: object.spec.replicas >= oldObject.spec.replicas
      message: replicas must not decrease
      messageExpression: "'replicas cannot decrease from ' + string(oldObject.spec.replicas) + ' to ' + string(object.spec.replicas)"
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: example-no-scale-down
spec:
  policyName: example-no-scale-down
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        cel-example: enabled
```

Updating replicas from 2 to 3 succeeds; from 3 to 2 fails. The rule targets normal Deployment writes, not `deployments/scale`; add and test a separate Scale-aware rule when scale-subresource enforcement is required.

### ConfigMap parameters

ConfigMap values are strings. Validate the key and numeric representation before conversion; this example accepts a maximum from 1 through 99. The bounded regular expression prevents conversion overflow. The ConfigMap is a **parameter**, not a container for policy YAML or an init-ca installation mechanism.

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-replica-limit
  namespace: default
data:
  maxReplicas: "5"
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: example-configmap-limit
spec:
  failurePolicy: Fail
  paramKind:
    apiVersion: v1
    kind: ConfigMap
  matchConstraints:
    resourceRules:
      - apiGroups: [apps]
        apiVersions: [v1]
        operations: [CREATE, UPDATE]
        resources: [deployments]
  validations:
    - expression: >-
        params != null && has(params.data) && 'maxReplicas' in params.data &&
        params.data['maxReplicas'].matches('^[1-9][0-9]?$') &&
        object.spec.replicas <= int(params.data['maxReplicas'])
      message: replicas exceed the limit or maxReplicas is missing or invalid
      messageExpression: >-
        params != null && has(params.data) && 'maxReplicas' in params.data &&
        params.data['maxReplicas'].matches('^[1-9][0-9]?$')
        ? 'replicas must not exceed ' + params.data['maxReplicas']
        : 'maxReplicas must be an integer from 1 to 99'
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: example-configmap-limit
spec:
  policyName: example-configmap-limit
  paramRef:
    name: example-replica-limit
    namespace: default
    parameterNotFoundAction: Deny
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        cel-example: enabled
```

### Custom CRD parameters

Use a typed parameter for structured policy configuration. Install the CRD first and wait for its `Established` condition before creating its instance or policy; applying all documents once can race API discovery. The parameter schema validates the limit independently of the Deployment policy.

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: replicabudgets.examples.run.ai
spec:
  group: examples.run.ai
  scope: Namespaced
  names:
    plural: replicabudgets
    singular: replicabudget
    kind: ReplicaBudget
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required: [spec]
          properties:
            spec:
              type: object
              required: [maxReplicas]
              properties:
                maxReplicas:
                  type: integer
                  minimum: 1
                  maximum: 99
---
apiVersion: examples.run.ai/v1alpha1
kind: ReplicaBudget
metadata:
  name: example-replica-limit
  namespace: default
spec:
  maxReplicas: 5
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: example-custom-limit
spec:
  failurePolicy: Fail
  paramKind:
    apiVersion: examples.run.ai/v1alpha1
    kind: ReplicaBudget
  matchConstraints:
    resourceRules:
      - apiGroups: [apps]
        apiVersions: [v1]
        operations: [CREATE, UPDATE]
        resources: [deployments]
  validations:
    - expression: params != null && object.spec.replicas <= params.spec.maxReplicas
      message: replicas exceed the configured limit
      messageExpression: "params != null ? 'replicas must not exceed ' + string(params.spec.maxReplicas) : 'ReplicaBudget is required'"
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: example-custom-limit
spec:
  policyName: example-custom-limit
  paramRef:
    name: example-replica-limit
    namespace: default
    parameterNotFoundAction: Deny
  validationActions: [Deny]
  matchResources:
    namespaceSelector:
      matchLabels:
        cel-example: enabled
```

For either parameter example, replicas 5 should pass and 6 should fail. Missing parameters must deny after admission caches observe the deletion; invalid ConfigMap values must also deny. Test each policy separately so another binding cannot mask a missing check. Restrict who can edit the namespace label and parameter objects, and verify parameter read authorization for intended callers. An explicit `paramRef.namespace` chooses that namespace; omitting it for a namespaced parameter uses the admitted resource's namespace. `parameterNotFoundAction` must be set explicitly; `Deny` uses the policy's `failurePolicy`, and these examples pair it with `Fail` and binding action `Deny`.

## Run:ai examples

### Workload field validation

**EXISTS:** `IdentityFields` declares `usage` (`Template` or `Submit`, default `Submit`) and optional `baseWorkload`; TrainingWorkload v2alpha1 uses these shared fields. **NEW example constraint:** a template may not itself name a base workload. This illustrates the CEL translation of the design document; it is not a claim that this branch already enforces that rule. Confirm the product requirement and applicable workload kinds before adding it to a shared type.

Go marker to place on the owning spec or shared field-group type, after confirming its generated schema contains both fields:

```go
// +kubebuilder:validation:XValidation:rule="!has(self.usage) || self.usage != 'Template' || !has(self.baseWorkload) || self.baseWorkload == ''",message="templates must not reference a baseWorkload"
```

Equivalent YAML fragment under that object's schema (merge with the existing properties):

```yaml
x-kubernetes-validations:
  - rule: "!has(self.usage) || self.usage != 'Template' || !has(self.baseWorkload) || self.baseWorkload == ''"
    message: templates must not reference a baseWorkload
```

| Input | Result of the example rule |
|---|---|
| `usage: Template`, no `baseWorkload` | allow |
| `usage: Template`, `baseWorkload: ''` | allow |
| `usage: Template`, `baseWorkload: parent` | reject |
| `usage: Submit`, `baseWorkload: parent` | allow |
| usage absent | allow; existing schema defaults it to `Submit` |

Source locations: `sdk/apis/workload_fields/field_groups/fields_identity.go`, `workload-controller/pkg/api/trainingworkload/v2alpha1/types.go`, and `workload-controller/pkg/api/common/constants.go`, all under `runai-cluster/cluster/`.

### Webhook `matchConditions`: template skipping

A false match condition skips the **whole webhook**; it does not reject the request. All conditions must be true to call it. If none is false and evaluation errors, the webhook's `failurePolicy` decides whether to reject (`Fail`) or skip (`Ignore`).

**EXISTS:** init-ca registers workload ownership validators for UPDATE/DELETE. **NEW example:** filter a separate spec-only validator whose complete handler already bypasses templates. Do not add this to ownership/authorization validation just because the resource is a template. This branch does not establish that such a spec-only bypass exists; inspect the actual handler in the target version before applying it.

Append this Go fragment to the verified existing `admissionv1.ValidatingWebhook`; retain its service/TLS settings, operations, other conditions, and failure policy:

```go
webhook.MatchConditions = append(webhook.MatchConditions, admissionv1.MatchCondition{
	Name: "skip-template-spec-validation",
	Expression: "!(request.operation in ['CREATE', 'UPDATE']) || object == null || !has(object.spec) || !has(object.spec.usage) || object.spec.usage != 'Template'",
})
```

Equivalent YAML fragment inside that same webhook entry (append the item; do not replace the existing condition list):

```yaml
matchConditions:
  - name: skip-template-spec-validation
    expression: >-
      !(request.operation in ['CREATE', 'UPDATE']) || object == null ||
      !has(object.spec) || !has(object.spec.usage) || object.spec.usage != 'Template'
```

| Request | Match result | Effect |
|---|---|---|
| CREATE/UPDATE with new `usage: Template` | false | skip this spec-only webhook |
| CREATE/UPDATE with `usage: Submit` or usage absent | true | preserve the webhook call |
| UPDATE from Template to Submit | true | validate the submitted workload |
| UPDATE from Submit to Template | false | only safe if the full handler already permits this bypass |
| DELETE or another operation | true | preserve existing operation matching and deletion checks |

The condition only narrows the existing rules; it does not add CREATE if the webhook is registered for UPDATE/DELETE. Do not add broad `system:` username bypasses. For a handler whose checks depend on `oldObject`, preserve those checks instead of copying this new-object-only predicate.

## Best practices

1. **Match expression scope to schema scope.** `self` is the node carrying the CRD rule. Use `object`/`oldObject` in VAP or webhook filters. Required fields and defaults determine whether a direct access is safe.
2. **Guard optional values.** Use `has()` for each optional ancestor before dereferencing children. For a map key containing `/` or `-`, test membership (`'example.com/key' in labels`) before indexing. In CRD validation, a null value is treated as absent for `has()`.
3. **Preserve absence and transitions.** Field-level `oldSelf` does not freeze optional-field presence; use a parent rule for add/remove prevention. Guard `oldObject` on CREATE and `object` on DELETE. Do not require `optionalOldSelf` for ordinary 1.30 transitions.
4. **Bound cost.** Prefer schema `required`, `enum`, range and length constraints for simple checks. Bound list `maxItems`, map `maxProperties`, string `maxLength`, and nested collections used by CEL. Avoid nested scans when a bounded map or set can express the rule.
5. **Keep types explicit.** Convert ConfigMap strings only after validating their range/format. Convert numbers with `string()` for messages. Use a typed parameter CRD when structure matters.
6. **Make failures useful.** A validation expression returns a boolean; `messageExpression` returns a nonempty, single-line string. Supply a static fallback. Do not place credentials or confidential field contents in messages.
7. **Preserve enforcement semantics.** Trace accepted users, custom UIDs, template skips, subresources, operations, and installation flags. A tighter allowlist or newly unconditional CRD rule is a behavior change, not a mechanical migration.
8. **Declare missing-parameter behavior.** Use a matching policy/binding pair with explicit `parameterNotFoundAction`, `failurePolicy`, and validation actions; control who can edit parameters. A valid policy without a binding enforces nothing.
9. **Respect the supported version floor.** Compare API and library versions with the target cluster and controller-gen version. For schema field names that are CEL keywords, use the Kubernetes escape (for example `self.__namespace__`).
10. **Validate generated output and admission.** Regenerate every shipped CRD copy, inspect rule placement and VAP type-check warnings, and test allowed/rejected requests at the minimum supported Kubernetes version. A CEL string comparison alone cannot verify admission behavior.

## CEL quick reference

Variables below refer to the ExampleSpec schema unless a different context is stated. These are expressions, not complete manifests.

| Operation | CEL | Context / result |
|---|---|---|
| Field access | `self.minReplicas` | CRD rule on spec; integer |
| Existence | `has(self.image)` | CRD spec; boolean |
| Optional string | `!has(self.image) \|\| self.image.startsWith('registry.example.com/')` | allow absent; validate present |
| Cross-field | `self.minReplicas <= self.maxReplicas` | boolean |
| Previous value | `self == oldSelf` | CRD field transition |
| Previous object | `oldObject == null \|\| object.spec.replicas >= oldObject.spec.replicas` | VAP on Deployment CREATE/UPDATE only |
| List exists | `self.members.exists(m, m.name == 'worker')` | boolean |
| List all | `self.members.all(m, m.name.matches('^[a-z][a-z0-9-]*$'))` | boolean; true for an empty list |
| List map | `self.members.map(m, m.name)` | list of strings; project before another operation |
| String matching | `self.stableID.matches('^batch-[a-z]+$')` | boolean |
| Conditional | `self.maxReplicas <= 10 ? 'within limit' : 'over limit'` | string, suitable for a message rather than a rule |
| Type checking | `type(self.minReplicas) == int` | boolean; CEL integers are 64-bit |
| Size | `self.members.size() <= 16` | boolean; prefer `maxItems` for this alone |
| Map key | `has(object.metadata.labels) && 'example.com/team' in object.metadata.labels` | VAP CREATE/UPDATE; boolean |
| Authorization | `authorizer.requestResource.check('update').allowed()` | VAP authorization check for the current caller |

## Limitations: when a webhook is still needed

| Scenario | Why validating CEL is insufficient |
|---|---|
| Validate against an external HTTP service or database | expressions cannot issue arbitrary network/database calls |
| List arbitrary resources or follow a dynamic chain of references (for example a parent Department) | `paramRef` supplies declared parameter resources, not arbitrary `Get`/`List` |
| Resolve custom identity UIDs through arbitrary callbacks | a fixed username list is not equivalent to a lookup; only a suitable known parameter can move this check to VAP |
| Execute existing plugin logic that cannot be expressed as a bounded CEL expression | plugin code and arbitrary Go are not callable from CEL |
| Mutate/default an admitted object at the Kubernetes 1.30 floor | CRD validation and VAP only validate; mutation remains a separate mechanism |

An authorizer check and a declared parameter are supported exceptions to the “no external data” shorthand. A webhook also does not make multi-object changes atomic; use the correct controller or storage design if transactional consistency is required.

## Scope and sources

This skill supplies guidance and examples. It does not add migration automation, a CEL unit-test framework, Run:ai CI integration, or MutatingAdmissionPolicy implementation. The Tests section describes validation expected for a later implementation; it does not introduce new testing tooling.

- [RUN-37853](https://runai.atlassian.net/browse/RUN-37853): skill requirements.
- [CEL Validation Webhooks design](https://runai.atlassian.net/wiki/spaces/EN/pages/4490526721/CEL+Validation+Webhooks): background; verify examples against source and supported Kubernetes behavior before reuse.
- [Kubernetes CRD validation rules](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#validation-rules) and [Kubebuilder markers](https://book.kubebuilder.io/reference/markers/crd-validation.html): rule scope, transitions, messages, and generator syntax.
- [Kubernetes VAP](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/): policies, bindings, parameters, and authorization.
- [Webhook match conditions](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#matching-requests-matchconditions): filtering and failure behavior.

## Rollout (VAP only)

```text
1. PR A: add the Definition with validationActions [Audit]. Old webhook stays authoritative.
2. Verify the policy's typeChecking status is clean and Audit shows only intended writers.
3. PR B: flip the binding to [Deny]. Keep the webhook.
4. PR C: after Deny is proven in a live environment, remove the webhook registration,
   handler and its configuration.
   ✗ Never remove the webhook in the same PR that adds the policy.
```

## Tests

| Change | Required | Framework |
|---|---|---|
| VAP | unit test on the `Definition`: names, binding reference, resource rule, operations, selector, failure policy, validation actions, identity semantics (exact usernames or parameter UIDs). Plus an envtest against a real API server: intended writer allowed, unrelated ServiceAccount denied under `Deny`, allowed under `Audit`. | the owning service’s existing framework; the application-controller resource examples use testify in `validating_admission_policy_test.go` and `validating_admission_policy_envtest_test.go` |
| CRD-CEL | regenerate the CRD, assert the upgrader copy matches, envtest create/update of valid and invalid objects. | the owning service's existing framework. Ginkgo where the service uses Ginkgo. |
| Both | ✗ never test only the CEL string. | |

## Alert block

Print exactly this shape, once:

```text
💡 CEL opportunity
What:    <the webhook or validation being added or changed, one line>
Move to: VAP | CRD-CEL | webhook matchConditions
Because: <request context | owned schema | webhook invocation conditions>
Gate:    <CEL feature> GA <version> ≤ 1.30 ✓   (or ✗ and the alternative)
Pattern: <repo file to copy from>
Rollout: Audit → Deny → remove webhook   (replacing an existing webhook with VAP only)
Say "go CEL" to switch, or keep going as is.
```

For a new webhook with no applicable CEL opportunity, use a short assessment instead: `CEL assessment: keep the webhook because <concrete dependency or unsupported operation>.` Do not invent an opportunity or invite a switch to an unsupported feature. For a new validation implemented directly in CEL, there is no existing webhook to remove; follow the requested enforcement semantics and the Tests section.

When Step 3 fails on already-CEL code, print only the `Gate:` line with ✗ and the allowed alternative. Omit the invitation to switch when implementation is already authorized.

## Code references

| Ref | Location | What it is |
|---:|---|---|
| 1 | `runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy/definition.go` | `Definition` type and validation |
| 2 | `runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy/resource.go` | `NewResources` |
| 3 | `runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy/legacy.go` | `NewResourcesOld` for `OperandOld` operands |
| 4 | `runai-cluster/cluster/go-operator/internal/run/operands/application-controller/resources/validating_admission_policy.go` | canonical identity VAP |
| 5 | `runai-cluster/cluster/go-operator/internal/run/operands/application-controller/resources/validating_admission_policy_envtest_test.go` | canonical API-server-backed test |
| 6 | `runai-cluster/cluster/sdk/apis/run/v1alpha1/application_types.go` | canonical `XValidation` marker |
| 7 | `runai-cluster/cluster/common/utils/validator/service_account_validator.go` | legacy identity check being replaced |
| 8 | `runai-cluster/cluster/go-operator/internal/run/operands/init-ca/resources/config_maps.go` | current webhook registrations |
| 9 | `runai-cluster/cluster/go-operator/controllers/run/utils.go` | `minimumKubernetesVersion`, `isVAPSupported` |
| 10 | `runai-cluster/cluster/runai-chart/Chart.yaml` | `kubeVersion` floor |
| 11 | `runai-cluster/cluster/runai-crd-upgrader/crds/` | shipped CRD copies to regenerate |
| 12 | `runai-cluster/cluster/go-operator/internal/run/operands/anyworkload-controller/resources/validating_admission_policy.go` | existing named ServiceAccount parameter and UID comparison |

## 🧠 Remember it as

**Owned schema → CRD-CEL. Request or declared parameters → VAP. Arbitrary lookup → webhook. Check the 1.30 floor.**
