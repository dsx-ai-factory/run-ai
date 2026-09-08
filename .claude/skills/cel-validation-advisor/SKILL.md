---
name: cel-validation-advisor
description: Advisory skill that fires automatically while writing or editing Kubernetes admission validation in the Run:ai monorepo. Use when touching validating webhooks, admission handlers (admission.Handler, CustomValidator, ValidateCreate/Update/Delete), ServiceAccountValidator or request.UserInfo identity checks, ValidatingWebhookConfiguration rules in init-ca, Go field checks on CRD objects (required, mutually exclusive, immutable, unique), kubebuilder XValidation markers, x-kubernetes-validations, ValidatingAdmissionPolicy definitions, or any CEL expression. Tells the user when the validation can move to CEL (VAP or CRD-CEL), which repo pattern to copy, and whether the CEL features used are GA in Kubernetes 1.30, the Run:ai minimum. Never rewrites code on its own.
allowed-tools: Bash, Read, Grep, Glob
---

## TL;DR (the whole thing in 3 lines)

- You are writing admission-time validation for a Kubernetes object. Before continuing, decide whether it belongs in **CEL** instead of a webhook.
- Caller identity → `ValidatingAdmissionPolicy` (VAP). Object shape only → CRD CEL (`XValidation`). Anything that reads other objects, runs plugins, or mutates → the webhook stays.
- Every CEL feature you use must be **GA in Kubernetes ≤ 1.30** (chart floor `kubeVersion: '>=1.30.0-0'`). Alert the user; do not switch the implementation unless they say so.

## Behavior

1. Run the three steps below silently.
2. If a step yields an alert, print the **Alert block** once, then continue the user's original task unchanged.
3. Switch to CEL only when the user answers `go CEL`. Then follow **Patterns**, **Rollout** and **Tests**.
4. Never post the alert twice for the same validation in one session.

## Step 1: DETECT

Fire when the code being written or edited matches a signal. Stay silent otherwise.

| Signal in code | Meaning |
|---|---|
| `admission.Handler`, `CustomValidator`, `ValidateCreate`/`ValidateUpdate`/`ValidateDelete`, `admission.Denied(` | new or changed webhook logic |
| `ServiceAccountValidator`, `HasPrefix(username, "system:serviceaccount:")`, `request.UserInfo`, `AdmissionRequest.UserInfo` | caller identity check |
| new rule in a `ValidatingWebhookConfiguration` (init-ca `config_maps.go`) | registering a webhook |
| Go `if` checks on a CR: required field, mutually exclusive fields, immutable field, unique list items | object-shape check written in Go |
| `+kubebuilder:validation:XValidation`, `x-kubernetes-validations`, `admissionv1.ValidatingAdmissionPolicy{`, any CEL string | already CEL → run **Step 3** only |

**Explicitly silent** (never alert):

- REST input validation in `control-plane/` services. It is not Kubernetes admission.
- Checks inside `Reconcile` that only set status or conditions.
- Existing webhook code that is read or reviewed but not changed.
- Mutating webhooks and defaulting. `MutatingAdmissionPolicy` is GA only in 1.36.

## Step 2: CLASSIFY

Walk the table top to bottom. The first matching row is the verdict.

| The check needs | Verdict | Copy from |
|---|---|---|
| caller identity only | **VAP**, exact username | `application-controller/resources/validating_admission_policy.go` |
| caller identity, label-scoped objects | **VAP** + `objectSelector` on the binding | same file; selector goes on `Binding.Spec.MatchResources` |
| fields of the admitted object only | **CRD-CEL** `XValidation` | `sdk/apis/run/v1alpha1/application_types.go` (`chartRef` xor `chartUrl`) |
| old vs new value of a field | **CRD-CEL** with `oldSelf` | `sdk/apis/kai/v1alpha1/nodepool_types.go` (immutable `labelKey`) |
| an authorization check on the caller (SubjectAccessReview) | **VAP** with the `authorizer` CEL variable | none in repo yet; see Kubernetes VAP docs |
| one well-known parameter object | **VAP** with `paramRef`, but prefer embedding exact usernames; never `paramRef` to `RunaiConfig` | anyworkload-controller WorkloadRunner VAP (ServiceAccount param) |
| arbitrary other objects, `Get`/`List`, custom lookups | **webhook stays** | project-controller parent Department lookup |
| plugin framework or arbitrary Go | **webhook stays** | dromos WorkloadRunner plugin validators |
| mutation or defaulting | **webhook stays** | none |

Remember it as: **CEL sees only the request and the object. If the check needs anything else, it is a webhook.**

## Step 3: GATE

Run:ai floor: Kubernetes **1.30** (`runai-chart/Chart.yaml`, installer `minimumKubernetesVersion`, operator `isVAPSupported`). A feature is allowed only if it is GA at or below 1.30.

| Feature | GA | Verdict |
|---|---|---|
| `ValidatingAdmissionPolicy` v1: `variables`, `matchConditions`, `paramRef`, `messageExpression`, `auditAnnotations`, `authorizer` | 1.30 | ✓ allow |
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

Source: Kubernetes docs, `reference/using-api/cel`, `reference/access-authn-authz/validating-admission-policy`, `reference/command-line-tools-reference/feature-gates` (active gates) and `feature-gates-removed` (graduated gates). Re-verify these when the chart floor changes.

## Patterns (when the user says `go CEL`)

### VAP

1. Build one `admissionpolicy.Definition` per target resource: typed `Policy`, typed `Binding` whose `PolicyName` matches, optional `Enabled` callback.
2. Hand it to `admissionpolicy.NewResources` (resource-based operands) or `admissionpolicy.NewResourcesOld` (legacy `OperandOld` operands). The owning operand is the service that owns the webhook today. No new operand, no init-ca ConfigMap, no shared VAP, no `paramRef` to `RunaiConfig`.
3. Identity is the complete username: `request.userInfo.username in variables.allowedServiceAccounts`, values shaped `system:serviceaccount:<installation-namespace>:<effective-service-account>`. Derive namespace and ServiceAccount name from `RunaiConfig`, including NMC overrides. No prefix, no substring, no UID comparison.
4. Both objects carry the label `run.ai/created-by: runai-operator`. Do not use `app`; reviewers reject it as too common.
5. `failurePolicy: Fail` unless the replaced webhook was `Ignore`.

### CRD-CEL

1. Add `// +kubebuilder:validation:XValidation:rule="...",message="..."` on the Go type or field in `*_types.go`. Escape reserved words: `self.__namespace__`.
2. Regenerate the CRD and update every shipped copy, including `runai-cluster/cluster/runai-crd-upgrader/crds/`.
3. Remove only the equivalent branch from the webhook. Keep the webhook if it still runs other checks.
4. If the webhook being replaced is disabled in some installations, adding a CRD rule enforces new behavior there. Say so in the alert and require an audit of stored objects before enabling.

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
| VAP | unit test on the `Definition`: names, binding reference, resource rule, operations, selector, failure policy, validation actions, exact usernames. Plus an envtest against a real API server: intended writer allowed, unrelated ServiceAccount denied under `Deny`, allowed under `Audit`. | testify, as in `application-controller/resources/validating_admission_policy_test.go` and `validating_admission_policy_envtest_test.go` |
| CRD-CEL | regenerate the CRD, assert the upgrader copy matches, envtest create/update of valid and invalid objects. | the owning service's existing framework. Ginkgo where the service uses Ginkgo. |
| Both | ✗ never test only the CEL string. | |

## Alert block

Print exactly this shape, once:

```text
💡 CEL opportunity
What:    <the validation being written, one line>
Move to: VAP | CRD-CEL          because: <caller identity | object shape only>
Gate:    <CEL feature> GA <version> ≤ 1.30 ✓   (or ✗ and the alternative)
Pattern: <repo file to copy from>
Rollout: Audit → Deny → remove webhook   (VAP only)
Say "go CEL" to switch, or keep going as is.
```

When Step 3 fails on already-CEL code, print only the `Gate:` line with ✗ and the allowed alternative.

## Code references

| Ref | Location | What it is |
|---:|---|---|
| 1 | `runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy/definition.go` | `Definition` type and validation |
| 2 | `runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy/resource.go` | `NewResources` |
| 3 | `runai-cluster/cluster/go-operator/pkg/operands/admissionpolicy/legacy.go` | `NewResourcesOld` for `OperandOld` operands |
| 4 | `runai-cluster/cluster/go-operator/internal/run/operands/application-controller/resources/validating_admission_policy.go` | canonical identity VAP |
| 5 | `runai-cluster/cluster/go-operator/internal/run/operands/application-controller/resources/validating_admission_policy_envtest_test.go` | canonical API-server-backed test |
| 6 | `runai-cluster/cluster/sdk/apis/run/v1alpha1/application_types.go` | canonical `XValidation` marker |
| 7 | `runai-cluster/cluster/sdk/apis/kai/v1alpha1/nodepool_types.go` | canonical `oldSelf` immutability rule |
| 8 | `runai-cluster/cluster/common/utils/validator/service_account_validator.go` | legacy identity check being replaced |
| 9 | `runai-cluster/cluster/go-operator/internal/run/operands/init-ca/resources/config_maps.go` | current webhook registrations |
| 10 | `runai-cluster/cluster/go-operator/controllers/run/utils.go` | `minimumKubernetesVersion`, `isVAPSupported` |
| 11 | `runai-cluster/cluster/runai-chart/Chart.yaml` | `kubeVersion` floor |
| 12 | `runai-cluster/cluster/runai-crd-upgrader/crds/` | shipped CRD copies to regenerate |

## 🧠 Remember it as

**Who is asking → VAP. What does it look like → CRD-CEL. Needs anything else → webhook. Nothing newer than 1.30.**
