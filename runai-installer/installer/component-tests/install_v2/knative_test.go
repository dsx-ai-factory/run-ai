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

// knativeServingCRD seeds the operator's KnativeServing CRD. established
// controls whether it carries an Established=True condition, so tests can
// exercise the wait-for-CRD gate in the serving step (the fake operator never
// actually reconciles a CRD into the cluster).
func knativeServingCRD(established bool) *unstructured.Unstructured {
	status := map[string]any{}
	if established {
		status["conditions"] = []any{map[string]any{"type": "Established", "status": "True"}}
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "CustomResourceDefinition",
		"metadata":   map[string]any{"name": "knativeservings.operator.knative.dev"},
		"status":     status,
	}}
}

// knativeOperatorDeployment seeds the operator's controller Deployment. available
// controls whether it carries an Available=True condition, so tests can exercise
// the wait-for-operator-readiness gate. Seeding it also makes the operator
// install step detect it as already present (the fake Helm installer does not
// materialize the Deployment), so tests that need the operator install to run
// must leave it unseeded.
func knativeOperatorDeployment(available bool) *unstructured.Unstructured {
	status := map[string]any{}
	if available {
		status["conditions"] = []any{map[string]any{"type": "Available", "status": "True"}}
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata":   map[string]any{"name": "knative-operator", "namespace": "knative-operator"},
		"status":     status,
	}}
}

// seedKnativeReady returns the seed map for a run where both gates pass: the
// operator CRD is established and the operator Deployment is available.
func seedKnativeReady() map[client.ObjectKey]*unstructured.Unstructured {
	return map[client.ObjectKey]*unstructured.Unstructured{
		{Name: "knativeservings.operator.knative.dev"}:           knativeServingCRD(true),
		{Namespace: "knative-operator", Name: "knative-operator"}: knativeOperatorDeployment(true),
	}
}

// applyCallByKind returns the recorded apply call for the given kind, or nil.
func applyCallByKind(k8s *harness.FakeK8sExecutor, kind string) *harness.K8sCall {
	for i := range k8s.CallsByVerb("apply") {
		if c := k8s.CallsByVerb("apply")[i]; c.Kind == kind {
			return &c
		}
	}
	return nil
}

// runKnativeInstall runs a connected install with Knative enabled, an inference
// wildcard domain/cert, and the operator CRD seeded as established. mutate may
// further tweak the config (e.g. set an ingress class).
func runKnativeInstall(t *testing.T, mutate func(*installconfig.Config)) (*harness.FakeHelmInstaller, *harness.FakeManifestProvider, *harness.FakeK8sExecutor, []map[string]any) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	k8s := &harness.FakeK8sExecutor{SeededObjects: seedKnativeReady()}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	wcCert, wcKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.Inference = &installconfig.Inference{
		WildcardDomain: "*.inference.test.runai.local",
		WildcardTLS:    &installconfig.TLS{CertPath: wcCert, KeyPath: wcKey},
	}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
	mutate(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return helm, manifests, k8s, readEvents(t, eventsPath)
}

// TestConnectedInstall_Knative verifies the full Knative Serving install: the
// serving namespace, the wildcard TLS secret, the KnativeServing CR + fronting
// Ingress (gated on the operator CRD and operator readiness), and the
// serving-HPA manifest, all grouped under one "knative" prerequisite. The
// operator is seeded ready (the fake Helm installer does not materialize its
// Deployment), so the operator install step is detected as already present.
func TestConnectedInstall_Knative(t *testing.T) {
	_, manifests, k8s, evs := runKnativeInstall(t, func(*installconfig.Config) {})

	// The serving-HPA manifest is fetched from the matching upstream release.
	fetched := manifests.CallsByOp("fetch")
	require.Len(t, fetched, 1)
	assert.Contains(t, fetched[0].URL, "serving-hpa.yaml")
	assert.Contains(t, fetched[0].URL, "knative-v1.18.2")

	applied := appliedNames(k8s)
	assert.True(t, applied["runai-cluster-inference-tls"], "the wildcard TLS secret is applied")

	// The serving step gates the CR on the operator CRD being established and
	// the operator Deployment being available.
	assert.Positive(t, getCount(k8s, "knativeservings.operator.knative.dev"), "the serving step waits on the operator CRD")

	// The KnativeServing CR carries the expected config: the serving domain is
	// the wildcard with its "*." stripped, Kourier is the internal ingress, the
	// external scheme is https, and scale-to-zero is on.
	cr := applyCallByKind(k8s, "KnativeServing")
	require.NotNil(t, cr, "the KnativeServing CR is applied")
	assert.Equal(t, "knative-serving", cr.Namespace)
	cu := cr.Object.(*unstructured.Unstructured)
	domain, found, _ := unstructured.NestedMap(cu.Object, "spec", "config", "domain")
	require.True(t, found)
	_, hasStrippedDomain := domain["inference.test.runai.local"]
	assert.True(t, hasStrippedDomain, "serving domain is the wildcard with its leading *. stripped")
	ingressClass, _, _ := unstructured.NestedString(cu.Object, "spec", "config", "network", "ingress-class")
	assert.Equal(t, "kourier.ingress.networking.knative.dev", ingressClass)
	scheme, _, _ := unstructured.NestedString(cu.Object, "spec", "config", "network", "default-external-scheme")
	assert.Equal(t, "https", scheme)
	scaleToZero, _, _ := unstructured.NestedString(cu.Object, "spec", "config", "autoscaler", "enable-scale-to-zero")
	assert.Equal(t, "true", scaleToZero)

	// The Run:ai-required Serving feature flags must land under spec.config.features
	// (the operator writes them into the config-features ConfigMap that Run:ai's
	// cluster operator reads to mark Knative available).
	servingFeatures, foundFeatures, _ := unstructured.NestedStringMap(cu.Object, "spec", "config", "features")
	require.True(t, foundFeatures, "the Run:ai-required Serving feature flags are injected into the KnativeServing CR")
	assert.Equal(t, "enabled", servingFeatures["kubernetes.podspec-affinity"], "podspec-affinity must be enabled (required by Run:ai)")
	assert.Equal(t, "enabled", servingFeatures["kubernetes.podspec-schedulername"], "podspec-schedulername must be enabled (required by Run:ai)")

	// The fronting Ingress terminates the wildcard TLS secret and routes to the
	// Kourier gateway, with the default nginx ingress class.
	ingress := applyCallByKind(k8s, "Ingress")
	require.NotNil(t, ingress, "the fronting Ingress is applied")
	iu := ingress.Object.(*unstructured.Unstructured)
	className, _, _ := unstructured.NestedString(iu.Object, "spec", "ingressClassName")
	assert.Equal(t, "nginx", className, "the fronting Ingress defaults to the nginx class")
	tls, _, _ := unstructured.NestedSlice(iu.Object, "spec", "tls")
	require.Len(t, tls, 1)
	secretName, _, _ := unstructured.NestedString(tls[0].(map[string]any), "secretName")
	assert.Equal(t, "runai-cluster-inference-tls", secretName)
	host, _, _ := unstructured.NestedString(firstRule(t, iu), "host")
	assert.Equal(t, "*.inference.test.runai.local", host)
	assert.Equal(t, "kourier", ingressBackendService(t, iu), "the Ingress routes to the Kourier gateway")

	for _, step := range []string{
		"create-knative-serving-namespace",
		"install-knative-serving-tls",
		"install-knative-serving",
		"install-knative-serving-hpa",
	} {
		require.NotNil(t, findEvent(evs, "step_succeeded", step), "step %s must succeed", step)
	}
	// The operator is seeded ready, so its install step is detected as present.
	opSkipped := findEvent(evs, "step_skipped", "install-knative-operator")
	require.NotNil(t, opSkipped)
	assert.Equal(t, "already_installed", opSkipped["skipReason"])

	prereq := prereqSummaryEvent(t, evs)
	assert.EqualValues(t, 3, prereq["installed"], "prometheus + gpu + network")
	assert.Equal(t, "installed", prereq["outcome"])

	features := featuresSummaryEvent(t, evs)
	assert.EqualValues(t, 1, features["installed"], "knative")
	assert.Equal(t, "installed", features["outcome"])
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// getCount returns how many Get calls targeted the named object (any namespace).
func getCount(k8s *harness.FakeK8sExecutor, name string) int {
	n := 0
	for _, c := range k8s.CallsByVerb("get") {
		if c.Name == name {
			n++
		}
	}
	return n
}

// firstRule returns the first Ingress rule map.
func firstRule(t *testing.T, ingress *unstructured.Unstructured) map[string]any {
	t.Helper()
	rules, found, _ := unstructured.NestedSlice(ingress.Object, "spec", "rules")
	require.True(t, found)
	require.NotEmpty(t, rules)
	return rules[0].(map[string]any)
}

// ingressBackendService returns the backend service name of the first path of
// the first rule.
func ingressBackendService(t *testing.T, ingress *unstructured.Unstructured) string {
	t.Helper()
	paths, found, _ := unstructured.NestedSlice(firstRule(t, ingress), "http", "paths")
	require.True(t, found)
	require.NotEmpty(t, paths)
	name, _, _ := unstructured.NestedString(paths[0].(map[string]any), "backend", "service", "name")
	return name
}

// TestConnectedInstall_Knative_CustomIngressClass verifies the fronting Ingress
// honours the class from externalConnection.ingress.className.
func TestConnectedInstall_Knative_CustomIngressClass(t *testing.T) {
	_, _, k8s, _ := runKnativeInstall(t, func(c *installconfig.Config) {
		c.ExternalConnection = &installconfig.ExternalConnection{
			Type:    "ingress",
			Ingress: &installconfig.IngressConnection{ClassName: "contour"},
		}
	})

	ingress := applyCallByKind(k8s, "Ingress")
	require.NotNil(t, ingress)
	u := ingress.Object.(*unstructured.Unstructured)
	className, _, _ := unstructured.NestedString(u.Object, "spec", "ingressClassName")
	assert.Equal(t, "contour", className)
}

// TestConnectedInstall_Knative_GatewayAPI verifies that under Gateway API
// routing the front-door is an HTTPRoute attached to the configured Gateway
// (not an Ingress), TLS is delegated to the Gateway (no wildcard secret), and
// the install still completes.
func TestConnectedInstall_Knative_GatewayAPI(t *testing.T) {
	_, _, k8s, evs := runKnativeInstall(t, func(c *installconfig.Config) {
		c.ExternalConnection = &installconfig.ExternalConnection{
			Type:       "gatewayApi",
			GatewayAPI: &installconfig.GatewayAPIConnection{Name: "runai-gw", Namespace: "gw-system"},
		}
		// The Gateway terminates TLS, so no wildcard certificate is supplied.
		c.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
	})

	assert.Nil(t, applyCallByKind(k8s, "Ingress"), "no fronting Ingress under Gateway API")

	route := applyCallByKind(k8s, "HTTPRoute")
	require.NotNil(t, route, "an HTTPRoute front-door is applied under Gateway API")
	assert.Equal(t, "knative-serving", route.Namespace)
	ru := route.Object.(*unstructured.Unstructured)
	parents, found, _ := unstructured.NestedSlice(ru.Object, "spec", "parentRefs")
	require.True(t, found)
	require.Len(t, parents, 1)
	parent := parents[0].(map[string]any)
	assert.Equal(t, "runai-gw", parent["name"], "the route attaches to the configured Gateway")
	assert.Equal(t, "gw-system", parent["namespace"])
	hostnames, _, _ := unstructured.NestedStringSlice(ru.Object, "spec", "hostnames")
	assert.Equal(t, []string{"*.inference.test.runai.local"}, hostnames)
	assert.Equal(t, "kourier", httpRouteBackendService(t, ru), "the route forwards to the Kourier gateway")

	// TLS is the Gateway's responsibility: no wildcard secret, no TLS step.
	assert.False(t, appliedNames(k8s)["runai-cluster-inference-tls"], "no wildcard TLS secret under Gateway API")
	assert.Nil(t, findEvent(evs, "step_succeeded", "install-knative-serving-tls"), "no TLS-secret step under Gateway API")

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-knative-serving"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// httpRouteBackendService returns the first backendRef's service name of the
// first rule.
func httpRouteBackendService(t *testing.T, route *unstructured.Unstructured) string {
	t.Helper()
	rules, found, _ := unstructured.NestedSlice(route.Object, "spec", "rules")
	require.True(t, found)
	require.NotEmpty(t, rules)
	backends, found, _ := unstructured.NestedSlice(rules[0].(map[string]any), "backendRefs")
	require.True(t, found)
	require.NotEmpty(t, backends)
	name, _, _ := unstructured.NestedString(backends[0].(map[string]any), "name")
	return name
}

// runKnativeWithSeed runs a connected Knative install with a caller-supplied
// seed map and per-step timeout, returning the fakes, events, and the run error
// (nil on success) so gating/failure paths can be asserted. The config always
// carries a valid wildcard domain + certificate so input validation passes.
func runKnativeWithSeed(t *testing.T, seed map[client.ObjectKey]*unstructured.Unstructured, stepTimeout time.Duration) (*harness.FakeHelmInstaller, *harness.FakeK8sExecutor, []map[string]any, error) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	k8s := &harness.FakeK8sExecutor{SeededObjects: seed}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	wcCert, wcKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.Inference = &installconfig.Inference{
		WildcardDomain: "*.inference.test.runai.local",
		WildcardTLS:    &installconfig.TLS{CertPath: wcCert, KeyPath: wcKey},
	}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	exec := execOpts()
	exec.StepTimeout = stepTimeout
	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, exec)
	return helm, k8s, readEvents(t, eventsPath), err
}

// TestConnectedInstall_Knative_GatesOnOperatorReadiness verifies two things at
// once: the operator is installed when absent, and the KnativeServing CR apply
// is gated on the operator Deployment becoming available. The CRD is seeded
// established but the operator Deployment is left unseeded (so it never becomes
// available), so the serving step blocks on operator readiness and the apply
// never happens within the step timeout.
func TestConnectedInstall_Knative_GatesOnOperatorReadiness(t *testing.T) {
	seed := map[client.ObjectKey]*unstructured.Unstructured{
		{Name: "knativeservings.operator.knative.dev"}: knativeServingCRD(true),
	}
	helm, k8s, evs, err := runKnativeWithSeed(t, seed, 200*time.Millisecond)

	require.Error(t, err, "the serving step must time out waiting for operator readiness")
	assert.Len(t, helm.CallsByRelease("knative-operator"), 1, "the operator is installed when absent")
	require.NotNil(t, findEvent(evs, "step_failed", "install-knative-serving"))
	assert.Nil(t, applyCallByKind(k8s, "KnativeServing"), "the CR is not applied until the operator is available")
}

// TestConnectedInstall_Knative_GatesOnCRDEstablished verifies the serving step
// blocks on the operator CRD being established before applying the CR. The
// operator Deployment is seeded available but the CRD is seeded NOT established,
// so the first gate never opens and the apply never happens.
func TestConnectedInstall_Knative_GatesOnCRDEstablished(t *testing.T) {
	seed := map[client.ObjectKey]*unstructured.Unstructured{
		{Name: "knativeservings.operator.knative.dev"}:           knativeServingCRD(false),
		{Namespace: "knative-operator", Name: "knative-operator"}: knativeOperatorDeployment(true),
	}
	_, k8s, evs, err := runKnativeWithSeed(t, seed, 200*time.Millisecond)

	require.Error(t, err, "the serving step must time out waiting for the CRD to be established")
	require.NotNil(t, findEvent(evs, "step_failed", "install-knative-serving"))
	assert.Nil(t, applyCallByKind(k8s, "KnativeServing"), "the CR is not applied until the CRD is established")
}

// TestConnectedInstall_Knative_TLSFromClusterCert verifies the wildcard TLS
// secret falls back to the cluster certificate when no dedicated wildcard
// certificate is set (the cluster cert's SAN covers the wildcard domain).
func TestConnectedInstall_Knative_TLSFromClusterCert(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	k8s := &harness.FakeK8sExecutor{SeededObjects: seedKnativeReady()}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	// A cluster cert whose SAN covers both the cluster domain and the wildcard;
	// no dedicated wildcard cert, so the install falls back to it.
	cert, key := harness.GenCert(t, []string{"test.runai.local", "*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.TLS = &installconfig.TLS{CertPath: cert, KeyPath: key}
	cfg.Cluster.Inference = &installconfig.Inference{WildcardDomain: "*.inference.test.runai.local"}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.True(t, appliedNames(k8s)["runai-cluster-inference-tls"], "the wildcard TLS secret is created from the cluster cert fallback")
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-knative-serving-tls"))
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-knative-serving"))
}

// TestConnectedInstall_Knative_AlreadyPresent verifies an existing Knative
// install is detected and skipped: the operator Deployment, the KnativeServing
// CR, and the serving-HPA Deployment are all present, so no step reinstalls.
func TestConnectedInstall_Knative_AlreadyPresent(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	k8s := &harness.FakeK8sExecutor{
		Existing: map[client.ObjectKey]bool{
			{Namespace: "knative-operator", Name: "knative-operator"}: true,
			{Namespace: "knative-serving", Name: "knative-serving"}:   true,
			{Namespace: "knative-serving", Name: "autoscaler-hpa"}:    true,
		},
		SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
			{Namespace: "knative-serving", Name: "knative-serving"}: knativeServingObjectForTest(),
		},
	}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	wcCert, wcKey := harness.GenCert(t, []string{"*.inference.test.runai.local"}, time.Now().Add(24*time.Hour))
	cfg.Cluster.Inference = &installconfig.Inference{
		WildcardDomain: "*.inference.test.runai.local",
		WildcardTLS:    &installconfig.TLS{CertPath: wcCert, KeyPath: wcKey},
	}
	cfg.Features = &installconfig.Features{Knative: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	assert.Empty(t, helm.CallsByRelease("knative-operator"), "present operator is not reinstalled")
	assert.Empty(t, manifests.CallsByOp("fetch"), "present serving-HPA is not reapplied")

	for _, step := range []string{"install-knative-operator", "install-knative-serving", "install-knative-serving-hpa"} {
		skipped := findEvent(evs, "step_skipped", step)
		require.NotNil(t, skipped, "step %s must be skipped", step)
		assert.Equal(t, "already_installed", skipped["skipReason"])
	}
}

// knativeServingObjectForTest builds a minimal KnativeServing CR for seeding the
// fake cluster's serving-step detection.
func knativeServingObjectForTest() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "operator.knative.dev/v1beta1",
		"kind":       "KnativeServing",
		"metadata":   map[string]any{"name": "knative-serving", "namespace": "knative-serving"},
	}}
}
