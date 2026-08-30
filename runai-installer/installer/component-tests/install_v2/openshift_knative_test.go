package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serverlessCSV = "serverless-operator.v1.36.1"

	knativeServingALMExamples = `[{"apiVersion":"operator.knative.dev/v1beta1","kind":"KnativeServing","metadata":{"name":"knative-serving","namespace":"knative-serving"},"spec":{"config":{"network":{"ingress-class":"kourier.ingress.networking.knative.dev"}}}}]`
)

// seedOpenShiftKnative resolves the OpenShift Serverless flow on the fake
// cluster: the operator's Subscription installs a Succeeded CSV carrying a
// KnativeServing alm-example, and the KnativeServing CRD is Established.
func seedOpenShiftKnative(k8s *harness.FakeK8sExecutor) {
	if k8s.SeededObjects == nil {
		k8s.SeededObjects = map[client.ObjectKey]*unstructured.Unstructured{}
	}
	k8s.SeededObjects[client.ObjectKey{Namespace: "openshift-serverless", Name: "serverless-operator"}] = olmSubscriptionWithCSV(serverlessCSV)
	k8s.SeededObjects[client.ObjectKey{Namespace: "openshift-serverless", Name: serverlessCSV}] = olmCSVWithExamples(knativeServingALMExamples)
	k8s.SeededObjects[client.ObjectKey{Name: "knativeservings.operator.knative.dev"}] = establishedCRDObject("knativeservings.operator.knative.dev")
}

// TestOpenShiftInstall_KnativeViaServerless verifies Knative installs through the
// OpenShift Serverless operator on OpenShift: the operator (cluster-wide), the
// knative-serving namespace, and a KnativeServing CR built from the operator's
// CSV default with the inference domain injected and the user's customResource
// merged on top — and no upstream Helm operator or fronting Ingress.
func TestOpenShiftInstall_KnativeViaServerless(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftKnative(k8s)
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	wcCert, wcKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.Inference = &installconfig.Inference{
		WildcardDomain: "*.inference.test.runai.local",
		WildcardTLS:    &installconfig.TLS{CertPath: wcCert, KeyPath: wcKey},
	}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{
		Mode: "install",
		OpenShift: &installconfig.DependencyOpenShift{
			CustomResource: map[string]any{"config": map[string]any{"autoscaler": map[string]any{"stable-window": "60s"}}},
		},
	}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("knative-operator"), "on OpenShift Knative must not install the upstream operator via Helm")

	sub := applyCallByKindName(k8s, "Subscription", "serverless-operator")
	require.NotNil(t, sub, "the OpenShift Serverless operator is subscribed via OLM")
	assert.Equal(t, "redhat-operators", subscriptionSource(t, sub))
	assert.Equal(t, "openshift-serverless", sub.Namespace)
	assert.Equal(t, "stable", subscriptionChannel(t, sub), "Serverless channel resolves from the compat matrix")
	assert.Equal(t, serverlessCSV, subscriptionField(t, sub, "startingCSV"), "the Serverless subscription pins the matrix version via startingCSV")
	assert.Equal(t, "Manual", subscriptionField(t, sub, "installPlanApproval"), "default approval strategy is Manual so OLM does not upgrade Serverless in the background")

	og := applyCallByKindName(k8s, "OperatorGroup", "openshift-serverless-og")
	require.NotNil(t, og)
	ogObj := og.Object.(*unstructured.Unstructured)
	_, hasTargets, _ := unstructured.NestedSlice(ogObj.Object, "spec", "targetNamespaces")
	assert.False(t, hasTargets, "the Serverless operator uses a cluster-wide (AllNamespaces) OperatorGroup")

	require.NotNil(t, applyCallByKindName(k8s, "Namespace", "knative-serving"), "the knative-serving namespace is created")
	assert.Nil(t, applyCallByKindName(k8s, "Ingress", "knative-serving"), "no nginx fronting Ingress on OpenShift — Serverless uses Routes")

	ks := applyCallByKindName(k8s, "KnativeServing", "knative-serving")
	require.NotNil(t, ks, "the KnativeServing CR is applied")
	ksObj := ks.Object.(*unstructured.Unstructured)

	domain, foundDomain, _ := unstructured.NestedMap(ksObj.Object, "spec", "config", "domain")
	require.True(t, foundDomain)
	_, hasInferenceDomain := domain["inference.test.runai.local"]
	assert.True(t, hasInferenceDomain, "the inference wildcard serving domain is injected into the KnativeServing CR")

	ingressClass, _, _ := unstructured.NestedString(ksObj.Object, "spec", "config", "network", "ingress-class")
	assert.Equal(t, "kourier.ingress.networking.knative.dev", ingressClass, "the operator's CSV default config is preserved")

	stableWindow, _, _ := unstructured.NestedString(ksObj.Object, "spec", "config", "autoscaler", "stable-window")
	assert.Equal(t, "60s", stableWindow, "the user's customResource override is merged on top")

	// Run:ai's cluster operator only marks Knative available (enabling "standard"
	// inference in the UI) when the required Serving feature flags are enabled, so
	// the OpenShift CR must inject them (config-features).
	features, foundFeatures, _ := unstructured.NestedStringMap(ksObj.Object, "spec", "config", "features")
	require.True(t, foundFeatures, "the Run:ai-required Serving feature flags are injected into the KnativeServing CR")
	assert.Equal(t, "enabled", features["kubernetes.podspec-affinity"], "podspec-affinity must be enabled (required by Run:ai)")
	assert.Equal(t, "enabled", features["kubernetes.podspec-schedulername"], "podspec-schedulername must be enabled (required by Run:ai)")

	evs := readEvents(t, eventsPath)
	for _, step := range []string{
		"install-serverless-operator",
		"create-knative-serving-namespace",
		"install-knative-serving",
	} {
		require.NotNil(t, findEvent(evs, "step_succeeded", step), "step %s must succeed", step)
	}

	summary := featuresSummaryEvent(t, evs)
	assert.EqualValues(t, 1, summary["installed"], "the knative feature")
}

// runOpenShiftKnativeInstall runs a Knative-on-OpenShift install (Serverless
// flow seeded) and returns the emitted events. mutate may tweak the config
// (e.g. drop the inference wildcard certificate) before the run.
func runOpenShiftKnativeInstall(t *testing.T, mutate func(*installconfig.Config)) []map[string]any {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftKnative(k8s)
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	cfg.Prerequisites = &installconfig.Prerequisites{Skip: true}
	wcCert, wcKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.Inference = &installconfig.Inference{
		WildcardDomain: "*.inference.test.runai.local",
		WildcardTLS:    &installconfig.TLS{CertPath: wcCert, KeyPath: wcKey},
	}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
	if mutate != nil {
		mutate(cfg)
	}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return readEvents(t, eventsPath)
}

// TestPreflight_OpenShiftInferenceCert_WarnsWhenProvided verifies that supplying
// an inference wildcard certificate on OpenShift raises a non-blocking preflight
// warning: on OpenShift, Knative serving is exposed through OpenShift Routes (the
// router's certificate), so cluster.inference.wildcardTls is never applied.
func TestPreflight_OpenShiftInferenceCert_WarnsWhenProvided(t *testing.T) {
	evs := runOpenShiftKnativeInstall(t, nil)

	step := findEvent(evs, "step_succeeded", "preflight-openshift-inference-certificate")
	require.NotNil(t, step, "the OpenShift inference-certificate check must run when a wildcard cert is provided on OpenShift with Knative")
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "OpenShift Routes")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "a warning must not block the install")
}

// TestPreflight_OpenShiftInferenceCert_NoWarnWithoutCert verifies the check is
// not assembled when no inference wildcard certificate is provided (the common
// OpenShift case, where the router certificate is expected to serve inference).
func TestPreflight_OpenShiftInferenceCert_NoWarnWithoutCert(t *testing.T) {
	evs := runOpenShiftKnativeInstall(t, func(c *installconfig.Config) {
		// Drop the dedicated wildcard cert; give the cluster cert a SAN that
		// also covers the wildcard so input validation passes without an
		// inference-specific certificate being supplied.
		c.Cluster.Inference.WildcardTLS = nil
		cert, key := harness.GenCert(t, []string{"test.runai.local", "*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
		c.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
	})

	assert.Nil(t, findEvent(evs, "step_succeeded", "preflight-openshift-inference-certificate"),
		"no warning when no inference wildcard certificate is provided on OpenShift")
}
