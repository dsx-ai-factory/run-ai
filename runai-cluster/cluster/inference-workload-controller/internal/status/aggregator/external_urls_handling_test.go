package aggregator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBuildURLFromHTTPRoute(t *testing.T) {
	tests := []struct {
		name        string
		httpRoute   *gatewayv1.HTTPRoute
		expectedURL string
		expectErr   bool
	}{
		{
			name:      "nil HTTPRoute",
			httpRoute: nil,
			expectErr: true,
		},
		{
			name: "no hostnames",
			httpRoute: &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{},
			},
			expectErr: true,
		},
		{
			name: "hostname with path prefix match",
			httpRoute: &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Hostnames: []gatewayv1.Hostname{"app.example.com"},
					Rules: []gatewayv1.HTTPRouteRule{
						{
							Matches: []gatewayv1.HTTPRouteMatch{
								{
									Path: &gatewayv1.HTTPPathMatch{
										Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
										Value: ptr.To("/my-workspace"),
									},
								},
							},
						},
					},
				},
			},
			expectedURL: "https://app.example.com/my-workspace",
		},
		{
			name: "hostname with root path",
			httpRoute: &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Hostnames: []gatewayv1.Hostname{"app.example.com"},
					Rules: []gatewayv1.HTTPRouteRule{
						{
							Matches: []gatewayv1.HTTPRouteMatch{
								{
									Path: &gatewayv1.HTTPPathMatch{
										Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
										Value: ptr.To("/"),
									},
								},
							},
						},
					},
				},
			},
			expectedURL: "https://app.example.com/",
		},
		{
			name: "hostname with no rules defaults to root path",
			httpRoute: &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Hostnames: []gatewayv1.Hostname{"app.example.com"},
				},
			},
			expectedURL: "https://app.example.com/",
		},
		{
			name: "hostname with rule but no matches defaults to root path",
			httpRoute: &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Hostnames: []gatewayv1.Hostname{"app.example.com"},
					Rules: []gatewayv1.HTTPRouteRule{
						{},
					},
				},
			},
			expectedURL: "https://app.example.com/",
		},
		{
			name: "uses first hostname when multiple exist",
			httpRoute: &gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Hostnames: []gatewayv1.Hostname{"first.example.com", "second.example.com"},
					Rules: []gatewayv1.HTTPRouteRule{
						{
							Matches: []gatewayv1.HTTPRouteMatch{
								{
									Path: &gatewayv1.HTTPPathMatch{
										Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
										Value: ptr.To("/api"),
									},
								},
							},
						},
					},
				},
			},
			expectedURL: "https://first.example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := BuildURLFromHTTPRoute(tt.httpRoute)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedURL, result.String())
		})
	}
}

func TestSortHTTPRoutesByName(t *testing.T) {
	routes := map[string]*gatewayv1.HTTPRoute{
		"charlie": {ObjectMeta: metav1.ObjectMeta{Name: "charlie"}},
		"alpha":   {ObjectMeta: metav1.ObjectMeta{Name: "alpha"}},
		"bravo":   {ObjectMeta: metav1.ObjectMeta{Name: "bravo"}},
	}

	sorted := sortHTTPRoutesByName(routes)

	require.Len(t, sorted, 3)
	assert.Equal(t, "alpha", sorted[0].Name)
	assert.Equal(t, "bravo", sorted[1].Name)
	assert.Equal(t, "charlie", sorted[2].Name)
}
