package builders

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
)

func TestHAProxyClassVariantUsesOfficialPathRewrite(t *testing.T) {
	const (
		originalPath = "/team-a/test-workload"
		ingressClass = "HAProxy-public"
	)

	path, pathType := AddPathRewriteIngressRegex(originalPath, true, ingressClass)
	if path != originalPath {
		t.Fatalf("expected HAProxy prefix path %q, got %q", originalPath, path)
	}
	if pathType != networkingv1.PathTypePrefix {
		t.Fatalf("expected HAProxy path type %q, got %q", networkingv1.PathTypePrefix, pathType)
	}

	annotations := AddPathRewriteIngressAnnotations(nil, PathRewriteConfig{
		OriginalPath: originalPath,
		IngressClass: ingressClass,
	})
	const expectedRewrite = "/team-a/test-workload(/|$)(.*) /\\2"
	if annotations["haproxy.org/path-rewrite"] != expectedRewrite {
		t.Fatalf("expected official HAProxy rewrite %q, got %q", expectedRewrite, annotations["haproxy.org/path-rewrite"])
	}
}

func TestUnknownIngressClassDoesNotApplyControllerSpecificRewrite(t *testing.T) {
	const originalPath = "/team-a/test-workload"

	path, pathType := AddPathRewriteIngressRegex(originalPath, true, "traefik")
	if path != originalPath {
		t.Fatalf("expected unknown controller prefix path %q, got %q", originalPath, path)
	}
	if pathType != networkingv1.PathTypePrefix {
		t.Fatalf("expected unknown controller path type %q, got %q", networkingv1.PathTypePrefix, pathType)
	}

	annotations := AddPathRewriteIngressAnnotations(nil, PathRewriteConfig{
		OriginalPath: originalPath,
		IngressClass: "traefik",
	})
	if len(annotations) != 0 {
		t.Fatalf("expected no controller-specific annotations, got %v", annotations)
	}
}
