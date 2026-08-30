package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const almExamplesAnnotation = "alm-examples"

const (
	olmAPIVersionV1       = "operators.coreos.com/v1"
	olmAPIVersionV1alpha1 = "operators.coreos.com/v1alpha1"

	kindOperatorGroup         = "OperatorGroup"
	kindSubscription          = "Subscription"
	kindClusterServiceVersion = "ClusterServiceVersion"
	kindInstallPlan           = "InstallPlan"

	csvPhaseSucceeded = "Succeeded"

	installPlanApprovalAutomatic = "Automatic"
)

// csvReadyPollInterval is how long WaitForCSV waits between polls. It is a var
// (not a const) so tests can shorten it.
var csvReadyPollInterval = 5 * time.Second

type OLMCatalog struct {
	Source          string
	SourceNamespace string
}

// OperatorGroupObject builds an OperatorGroup. With targetNamespaces it scopes
// the operator to those namespaces (own/single-namespace install modes); with an
// empty list it omits the field, yielding a cluster-wide (AllNamespaces) group
// for operators that only support that mode (e.g. cert-manager).
func OperatorGroupObject(name, namespace string, targetNamespaces []string) *unstructured.Unstructured {
	og := &unstructured.Unstructured{}
	og.SetAPIVersion(olmAPIVersionV1)
	og.SetKind(kindOperatorGroup)
	og.SetName(name)
	og.SetNamespace(namespace)
	if len(targetNamespaces) > 0 {
		targets := make([]any, len(targetNamespaces))
		for i, ns := range targetNamespaces {
			targets[i] = ns
		}
		og.Object["spec"] = map[string]any{"targetNamespaces": targets}
	}

	return og
}

// SubscriptionObject builds an OLM Subscription. installPlanApproval selects the
// upgrade strategy ("Manual" pins the operator to startingCSV and blocks later
// upgrades; "Automatic" lets OLM track the channel head); an empty value
// defaults to Automatic to preserve OLM's own default.
func SubscriptionObject(name, namespace, pkg, channel, startingCSV, installPlanApproval string, catalog OLMCatalog, env map[string]string) *unstructured.Unstructured {
	if installPlanApproval == "" {
		installPlanApproval = installPlanApprovalAutomatic
	}
	spec := map[string]any{
		"name":                pkg,
		"channel":             channel,
		"source":              catalog.Source,
		"sourceNamespace":     catalog.SourceNamespace,
		"installPlanApproval": installPlanApproval,
	}
	if startingCSV != "" {
		spec["startingCSV"] = startingCSV
	}
	if len(env) > 0 {
		spec["config"] = map[string]any{"env": envVarList(env)}
	}

	sub := &unstructured.Unstructured{}
	sub.SetAPIVersion(olmAPIVersionV1alpha1)
	sub.SetKind(kindSubscription)
	sub.SetName(name)
	sub.SetNamespace(namespace)
	sub.Object["spec"] = spec

	return sub
}

// envVarList renders an env map as the OLM Subscription spec.config.env list
// ([{name, value}]), sorted by name so the rendered Subscription is stable.
func envVarList(env map[string]string) []any {
	names := make([]string, 0, len(env))
	for name := range env {
		names = append(names, name)
	}
	sort.Strings(names)

	out := make([]any, 0, len(env))
	for _, name := range names {
		out = append(out, map[string]any{"name": name, "value": env[name]})
	}

	return out
}

func subscriptionRef(name, namespace string) *unstructured.Unstructured {
	sub := &unstructured.Unstructured{}
	sub.SetAPIVersion(olmAPIVersionV1alpha1)
	sub.SetKind(kindSubscription)
	sub.SetName(name)
	sub.SetNamespace(namespace)

	return sub
}

func csvRef(name, namespace string) *unstructured.Unstructured {
	csv := &unstructured.Unstructured{}
	csv.SetAPIVersion(olmAPIVersionV1alpha1)
	csv.SetKind(kindClusterServiceVersion)
	csv.SetName(name)
	csv.SetNamespace(namespace)

	return csv
}

type WaitForCSV struct {
	Namespace        string
	SubscriptionName string
}

func (a WaitForCSV) Kind() string { return "olm.wait_csv" }

func (a WaitForCSV) Args() map[string]any {
	return map[string]any{"namespace": a.Namespace, "subscription": a.SubscriptionName}
}

func (a WaitForCSV) Describe() string {
	return fmt.Sprintf("wait for subscription %s/%s to install a ClusterServiceVersion that reaches Succeeded", a.Namespace, a.SubscriptionName)
}

func (a WaitForCSV) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	for {
		ready, err := a.succeeded(ctx, deps)
		if err != nil {
			return orchestrator2.ActionResult{}, err
		}
		if ready {
			return orchestrator2.ActionResult{}, nil
		}
		select {
		case <-ctx.Done():
			return orchestrator2.ActionResult{}, fmt.Errorf("waiting for subscription %s/%s CSV to succeed: %w", a.Namespace, a.SubscriptionName, ctx.Err())
		case <-time.After(csvReadyPollInterval):
		}
	}
}

func (a WaitForCSV) succeeded(ctx context.Context, deps orchestrator2.Deps) (bool, error) {
	sub := subscriptionRef(a.SubscriptionName, a.Namespace)
	err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: a.Namespace, Name: a.SubscriptionName}, sub)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	installedCSV, found, err := unstructured.NestedString(sub.Object, "status", "installedCSV")
	if err != nil {
		return false, err
	}
	if !found || installedCSV == "" {
		// Fail fast on an unresolvable subscription (e.g. a wrong package name or
		// channel, or a package missing from the catalog/air-gap mirror) instead
		// of polling until the step times out.
		if reason, message, failed := subscriptionResolutionFailed(sub); failed {
			return false, fmt.Errorf("subscription %s/%s cannot be resolved (%s): %s", a.Namespace, a.SubscriptionName, reason, message)
		}

		// Under a Manual approval strategy OLM creates the initial InstallPlan
		// unapproved and waits, so the operator would never install. Approve
		// that first plan here so the pinned (startingCSV) version installs;
		// later upgrade plans are left pending, which is the point of Manual.
		// Automatic plans are already approved, so this is a no-op for them.
		if err := a.approvePendingInstallPlan(ctx, deps, sub); err != nil {
			return false, err
		}

		return false, nil
	}

	csv := csvRef(installedCSV, a.Namespace)
	err = deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: a.Namespace, Name: installedCSV}, csv)
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	phase, _, err := unstructured.NestedString(csv.Object, "status", "phase")
	if err != nil {
		return false, err
	}

	return phase == csvPhaseSucceeded, nil
}

// subscriptionResolutionFailed reports an OLM ResolutionFailed condition on a
// subscription (set when the package/channel cannot be found in any catalog),
// so the wait can abort with the operator's own message rather than time out.
func subscriptionResolutionFailed(sub *unstructured.Unstructured) (reason, message string, failed bool) {
	conditions, found, err := unstructured.NestedSlice(sub.Object, "status", "conditions")
	if err != nil || !found {
		return "", "", false
	}
	for _, c := range conditions {
		cond, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if cond["type"] == "ResolutionFailed" && cond["status"] == "True" {
			reason, _ = cond["reason"].(string)
			message, _ = cond["message"].(string)

			return reason, message, true
		}
	}

	return "", "", false
}

// approvePendingInstallPlan approves the InstallPlan a Manual-approval
// Subscription is waiting on, so the initial (pinned) install can proceed. It is
// a no-op when the subscription has no InstallPlan yet, when the plan is already
// approved (the Automatic case), or when the plan has gone away.
func (a WaitForCSV) approvePendingInstallPlan(ctx context.Context, deps orchestrator2.Deps, sub *unstructured.Unstructured) error {
	name, namespace := installPlanRef(sub)
	if name == "" {
		return nil
	}
	if namespace == "" {
		namespace = a.Namespace
	}

	ip := installPlanRefObject(name, namespace)
	err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, ip)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	approved, _, err := unstructured.NestedBool(ip.Object, "spec", "approved")
	if err != nil {
		return err
	}
	if approved {
		return nil
	}

	patch := client.RawPatch(types.MergePatchType, []byte(`{"spec":{"approved":true}}`))

	return deps.K8sExecutor.Patch(ctx, ip, patch)
}

// installPlanRef reads the name/namespace of the InstallPlan a Subscription
// currently points at (status.installPlanRef).
func installPlanRef(sub *unstructured.Unstructured) (name, namespace string) {
	ref, found, err := unstructured.NestedMap(sub.Object, "status", "installPlanRef")
	if err != nil || !found {
		return "", ""
	}
	name, _ = ref["name"].(string)
	namespace, _ = ref["namespace"].(string)

	return name, namespace
}

func installPlanRefObject(name, namespace string) *unstructured.Unstructured {
	ip := &unstructured.Unstructured{}
	ip.SetAPIVersion(olmAPIVersionV1alpha1)
	ip.SetKind(kindInstallPlan)
	ip.SetName(name)
	ip.SetNamespace(namespace)

	return ip
}

// MergeCRSpec deep-merges a user-supplied override onto a base custom-resource
// spec, returning a fresh map (the inputs are never mutated). It is the
// OpenShift equivalent of merging custom Helm values, letting users customize an
// OLM operator's primary CR (ClusterPolicy, NicClusterPolicy, ...). Nested maps
// merge recursively; any other value (scalars, slices) is replaced wholesale by
// the override.
func MergeCRSpec(base, override map[string]any) map[string]any {
	out := make(map[string]any, len(base))
	for k, v := range base {
		out[k] = v
	}
	for k, ov := range override {
		if om, ok := ov.(map[string]any); ok {
			if bm, ok := out[k].(map[string]any); ok {
				out[k] = MergeCRSpec(bm, om)
				continue
			}
			out[k] = MergeCRSpec(map[string]any{}, om)

			continue
		}
		out[k] = ov
	}

	return out
}

// ApplyCSVExampleCR applies an operator's primary custom resource using the
// default spec the operator ships in its ClusterServiceVersion's alm-examples
// annotation (the same default the OpenShift console pre-fills), with the user's
// override merged on top. This avoids hand-authoring a full CR — whose required
// fields and image defaults are operator-version-specific — and fixes the
// "spec.<section>: Required value" rejection an empty CR would hit.
type ApplyCSVExampleCR struct {
	CSVNamespace     string
	SubscriptionName string
	CRAPIVersion     string
	CRKind           string
	CRName           string
	// CRNamespace is empty for cluster-scoped resources (e.g. ClusterPolicy).
	CRNamespace  string
	SpecOverride map[string]any
}

func (a ApplyCSVExampleCR) Kind() string { return "olm.apply_csv_example_cr" }

func (a ApplyCSVExampleCR) Args() map[string]any {
	return map[string]any{
		"csvNamespace": a.CSVNamespace,
		"subscription": a.SubscriptionName,
		"crKind":       a.CRKind,
		"crName":       a.CRName,
	}
}

func (a ApplyCSVExampleCR) Describe() string {
	return fmt.Sprintf("apply %s %q from the operator's CSV default (alm-examples), merged with any override", a.CRKind, a.CRName)
}

func (a ApplyCSVExampleCR) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	sub := subscriptionRef(a.SubscriptionName, a.CSVNamespace)
	if err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: a.CSVNamespace, Name: a.SubscriptionName}, sub); err != nil {
		return orchestrator2.ActionResult{}, err
	}

	installedCSV, _, err := unstructured.NestedString(sub.Object, "status", "installedCSV")
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if installedCSV == "" {
		return orchestrator2.ActionResult{}, fmt.Errorf("subscription %s/%s has no installed CSV", a.CSVNamespace, a.SubscriptionName)
	}

	csv := csvRef(installedCSV, a.CSVNamespace)
	if err := deps.K8sExecutor.Get(ctx, client.ObjectKey{Namespace: a.CSVNamespace, Name: installedCSV}, csv); err != nil {
		return orchestrator2.ActionResult{}, err
	}

	exampleSpec, err := exampleSpecForKind(csv, a.CRKind)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}

	cr := &unstructured.Unstructured{}
	cr.SetAPIVersion(a.CRAPIVersion)
	cr.SetKind(a.CRKind)
	cr.SetName(a.CRName)
	if a.CRNamespace != "" {
		cr.SetNamespace(a.CRNamespace)
	}
	cr.Object["spec"] = MergeCRSpec(exampleSpec, a.SpecOverride)

	return orchestrator2.ActionResult{}, deps.K8sExecutor.Apply(ctx, cr)
}

// exampleSpecForKind extracts the spec of the first alm-examples entry whose
// kind matches from a CSV. A missing annotation or kind is an error: we would
// otherwise apply an empty CR and hit the operator's required-field validation.
func exampleSpecForKind(csv *unstructured.Unstructured, kind string) (map[string]any, error) {
	raw, found, err := unstructured.NestedString(csv.Object, "metadata", "annotations", almExamplesAnnotation)
	if err != nil {
		return nil, err
	}
	if !found || raw == "" {
		return nil, fmt.Errorf("CSV %q has no %s annotation to seed the %s default", csv.GetName(), almExamplesAnnotation, kind)
	}

	var examples []map[string]any
	if err := json.Unmarshal([]byte(raw), &examples); err != nil {
		return nil, fmt.Errorf("parsing %s on CSV %q: %w", almExamplesAnnotation, csv.GetName(), err)
	}

	for _, ex := range examples {
		if ex["kind"] != kind {
			continue
		}
		if spec, ok := ex["spec"].(map[string]any); ok {
			return spec, nil
		}

		return map[string]any{}, nil
	}

	return nil, fmt.Errorf("no %s example in CSV %q alm-examples", kind, csv.GetName())
}
