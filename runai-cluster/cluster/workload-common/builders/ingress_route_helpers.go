package builders

import (
	"strings"

	openshiftv1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	networkingv1 "k8s.io/api/networking/v1"
)

// EnrichIngressWithDefaults applies cluster-wide defaults to an ingress while preserving user-specified values
func EnrichIngressWithDefaults(ingress *networkingv1.Ingress) *networkingv1.Ingress {
	if ingress == nil {
		return nil
	}

	if ingress.Spec.IngressClassName == nil || *ingress.Spec.IngressClassName == "" {
		ingressClass := workloadconfig.IngressClass()
		if ingressClass != "" {
			ingress.Spec.IngressClassName = &ingressClass
		}
	}

	if ingress.Annotations == nil {
		ingress.Annotations = make(map[string]string)
	}

	defaultAnnotations := workloadconfig.IngressAnnotations()
	for key, value := range defaultAnnotations {
		if _, exists := ingress.Annotations[key]; !exists {
			ingress.Annotations[key] = value
		}
	}

	// Set body size to 8GB (default nginx limit is 1MB)
	if _, exists := ingress.Annotations[constants.IngressBodySizeAnnotationKey]; !exists {
		ingress.Annotations[constants.IngressBodySizeAnnotationKey] = constants.IngressBodySizeAnnotationDefaultValue
	}

	return ingress
}

// EnrichRouteWithDefaults applies cluster-wide defaults to an OpenShift route while preserving user-specified values
func EnrichRouteWithDefaults(route *openshiftv1.Route) *openshiftv1.Route {
	if route == nil {
		return nil
	}

	if workloadconfig.SubdomainSupportEnabled() && route.Spec.Host != "" {
		route.Spec.Host = strings.Replace(route.Spec.Host, workloadconfig.RunaiTrimStr, "", 1)
	}

	if route.Spec.TLS == nil {
		route.Spec.TLS = &openshiftv1.TLSConfig{
			InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyRedirect,
			Termination:                   DefaultRouteTermination(),
		}
	}

	if route.Annotations == nil {
		route.Annotations = make(map[string]string)
	}

	defaultAnnotations := workloadconfig.IngressAnnotations()
	for key, value := range defaultAnnotations {
		if _, exists := route.Annotations[key]; !exists {
			route.Annotations[key] = value
		}
	}

	return route
}

// DefaultRouteTermination resolves the TLS termination type to apply to Run:ai-managed
// OpenShift Routes from the OPENSHIFT_ROUTE_TERMINATION_POLICY env var. Returns
// passthrough when explicitly configured, otherwise edge.
func DefaultRouteTermination() openshiftv1.TLSTerminationType {
	if workloadconfig.OpenshiftRouteTerminationPolicy() == workloadconfig.OpenshiftRouteTerminationPolicyPassthrough {
		return openshiftv1.TLSTerminationPassthrough
	}
	return openshiftv1.TLSTerminationEdge
}
