package aggregator

import (
	"testing"

	openshiftV1 "github.com/openshift/api/route/v1"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	network "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func newTestCalculator() *StatusCalculator {
	sc := NewStatusCalculator().(*StatusCalculator)
	sc.workload = &job.JobWorkload{}
	return sc
}

func ptrTo[T any](v T) *T {
	return &v
}

func makeIngress(name, host, path string) *network.Ingress {
	return &network.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: network.IngressSpec{
			Rules: []network.IngressRule{
				{
					Host: host,
					IngressRuleValue: network.IngressRuleValue{
						HTTP: &network.HTTPIngressRuleValue{
							Paths: []network.HTTPIngressPath{{Path: path}},
						},
					},
				},
			},
		},
	}
}

func makeHTTPRoute(name string, hostname gatewayv1.Hostname, path string) *gatewayv1.HTTPRoute {
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{hostname},
			Rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
								Value: ptrTo(path),
							},
						},
					},
				},
			},
		},
	}
}

func makeRoute(name, host string) *openshiftV1.Route {
	return &openshiftV1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       openshiftV1.RouteSpec{Host: host},
	}
}

func TestHandleJobUrls_HTTPRouteOnly(t *testing.T) {
	sc := newTestCalculator()
	sc.httpRoutes["route-a"] = makeHTTPRoute("route-a", "gw.example.com", "/predict")

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	require.Len(t, urls, 1)
	assert.Equal(t, "https://gw.example.com/predict", urls[0])
}

func TestHandleJobUrls_MultipleHTTPRoutes(t *testing.T) {
	sc := newTestCalculator()
	sc.httpRoutes["route-b"] = makeHTTPRoute("route-b", "b.example.com", "/b")
	sc.httpRoutes["route-a"] = makeHTTPRoute("route-a", "a.example.com", "/a")

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	require.Len(t, urls, 2)
	assert.Equal(t, "https://a.example.com/a", urls[0], "URLs should be sorted by resource name")
	assert.Equal(t, "https://b.example.com/b", urls[1])
}

func TestHandleJobUrls_DualModeDeduplicate(t *testing.T) {
	sc := newTestCalculator()
	sc.ingresses["ing"] = makeIngress("ing", "shared.example.com", "/api")
	sc.httpRoutes["hr"] = makeHTTPRoute("hr", "shared.example.com", "/api")

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	require.Len(t, urls, 1, "duplicate URLs should be deduplicated")
	assert.Equal(t, "https://shared.example.com/api", urls[0])
}

func TestHandleJobUrls_DualModeDistinctURLs(t *testing.T) {
	sc := newTestCalculator()
	sc.ingresses["ing"] = makeIngress("ing", "ingress.example.com", "/v1")
	sc.httpRoutes["hr"] = makeHTTPRoute("hr", "gateway.example.com", "/v2")

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	require.Len(t, urls, 2, "distinct URLs from different backends should both appear")
	assert.Contains(t, urls, "https://ingress.example.com/v1")
	assert.Contains(t, urls, "https://gateway.example.com/v2")
}

func TestHandleJobUrls_EmptyHTTPRoutes(t *testing.T) {
	sc := newTestCalculator()

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	assert.Empty(t, urls)
}

func TestHandleJobUrls_HTTPRouteNoHostnames(t *testing.T) {
	sc := newTestCalculator()
	sc.httpRoutes["bad"] = &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: "bad"},
		Spec:       gatewayv1.HTTPRouteSpec{},
	}

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	require.Len(t, urls, 1, "gracefully falls back to error URL")
}

func TestWithResources_DispatchHTTPRoute(t *testing.T) {
	sc := newTestCalculator()

	hr := makeHTTPRoute("test-hr", "gw.example.com", "/serve")
	ing := makeIngress("test-ing", "ingress.example.com", "/v1")

	sc.WithResources([]client.Object{hr, ing})

	assert.Len(t, sc.httpRoutes, 1)
	assert.Contains(t, sc.httpRoutes, "test-hr")
	assert.Len(t, sc.ingresses, 1)
	assert.Contains(t, sc.ingresses, "test-ing")
}

func TestWithResources_UnknownTypeIgnored(t *testing.T) {
	sc := newTestCalculator()

	sc.WithResources([]client.Object{
		makeHTTPRoute("hr", "gw.example.com", "/"),
	})

	assert.Len(t, sc.httpRoutes, 1)
	assert.Empty(t, sc.ingresses)
	assert.Empty(t, sc.routes)
}

func TestClear_ResetsHTTPRoutes(t *testing.T) {
	sc := newTestCalculator()
	sc.httpRoutes["hr"] = makeHTTPRoute("hr", "gw.example.com", "/")

	sc.Clear()

	assert.Empty(t, sc.httpRoutes)
}

func TestHandleJobUrls_AllThreeSourcesDeduplicate(t *testing.T) {
	sc := newTestCalculator()
	sc.ingresses["ing"] = makeIngress("ing", "shared.example.com", "/api")
	sc.routes["rt"] = makeRoute("rt", "shared.example.com")
	sc.httpRoutes["hr"] = makeHTTPRoute("hr", "shared.example.com", "/api")

	sc.handleJobUrls()

	urls := sc.workload.Status.InferenceStatus.ExternalUrls
	require.Len(t, urls, 2, "route produces https://shared.example.com/ which differs from /api path")

	urlSet := make(map[string]bool)
	for _, u := range urls {
		urlSet[u] = true
	}
	assert.True(t, urlSet["https://shared.example.com/api"])
}
