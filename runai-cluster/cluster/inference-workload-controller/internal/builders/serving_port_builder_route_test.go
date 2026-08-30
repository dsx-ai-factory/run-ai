package builders

import (
	"context"
	"os"
	"testing"

	openshiftv1 "github.com/openshift/api/route/v1"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	dinfv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

func newRouteTestBuilder(t *testing.T) *ServingPortBuilder {
	return NewServingPortBuilder(initClientWithResources(t, namespace), true)
}

func newRouteTestWorkload() *dinfv1.DistributedInferenceWorkload {
	return &dinfv1.DistributedInferenceWorkload{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dinfw-example",
			Namespace: namespace,
			Labels:    map[string]string{},
		},
	}
}

func TestBuildRoute_EdgeTerminationKeepsPathAndRewriteAnnotation(t *testing.T) {
	_ = os.Unsetenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy)

	b := newRouteTestBuilder(t)
	url, err := apis.ParseURL("https://runai-team-a-dinfw-example.cluster.example.com/team-a/dinfw-example")
	assert.NoError(t, err)

	route, err := b.buildRoute(context.Background(), newRouteTestWorkload(), "svc-name", url)
	assert.NoError(t, err)
	assert.NotNil(t, route)

	assert.Equal(t, "/team-a/dinfw-example", route.Spec.Path)
	assert.NotNil(t, route.Spec.TLS)
	assert.Equal(t, openshiftv1.TLSTerminationEdge, route.Spec.TLS.Termination)
	assert.Equal(t, openshiftv1.InsecureEdgeTerminationPolicyRedirect, route.Spec.TLS.InsecureEdgeTerminationPolicy)
	assert.Equal(t, "/\\1", route.Annotations["haproxy.router.openshift.io/rewrite-target"])
}

func TestBuildRoute_PassthroughStripsPathAndRewriteAnnotation(t *testing.T) {
	_ = os.Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, workloadconfig.OpenshiftRouteTerminationPolicyPassthrough)
	defer func() { _ = os.Unsetenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy) }()

	b := newRouteTestBuilder(t)
	url, err := apis.ParseURL("https://dinfw-example.cluster.example.com/team-a/dinfw-example")
	assert.NoError(t, err)

	route, err := b.buildRoute(context.Background(), newRouteTestWorkload(), "svc-name", url)
	assert.NoError(t, err)
	assert.NotNil(t, route)

	assert.Empty(t, route.Spec.Path, "passthrough Routes must not carry spec.path — OpenShift rejects them")
	assert.NotNil(t, route.Spec.TLS)
	assert.Equal(t, openshiftv1.TLSTerminationPassthrough, route.Spec.TLS.Termination)
	assert.Equal(t, openshiftv1.InsecureEdgeTerminationPolicyRedirect, route.Spec.TLS.InsecureEdgeTerminationPolicy)
	_, hasRewrite := route.Annotations["haproxy.router.openshift.io/rewrite-target"]
	assert.False(t, hasRewrite, "haproxy rewrite annotation is meaningless on passthrough Routes and should not be set")
}

func TestBuildRoute_PassthroughWithRootPathStillStripsPath(t *testing.T) {
	_ = os.Setenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy, workloadconfig.OpenshiftRouteTerminationPolicyPassthrough)
	defer func() { _ = os.Unsetenv(workloadconfig.EnvVarOpenshiftRouteTerminationPolicy) }()

	b := newRouteTestBuilder(t)
	url, err := apis.ParseURL("https://dinfw-example.cluster.example.com/")
	assert.NoError(t, err)

	route, err := b.buildRoute(context.Background(), newRouteTestWorkload(), "svc-name", url)
	assert.NoError(t, err)

	assert.Empty(t, route.Spec.Path, "even '/' is rejected on passthrough Routes by some OpenShift validators")
	assert.Equal(t, openshiftv1.TLSTerminationPassthrough, route.Spec.TLS.Termination)
}
