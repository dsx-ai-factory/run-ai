package builders

import (
	"fmt"
	"strings"

	networkingv1 "k8s.io/api/networking/v1"
)

const (
	IngressClassNginx   = "nginx"
	IngressClassHAProxy = "haproxy"
)

// PathRewriteConfig contains the configuration for path-based routing with rewriting
type PathRewriteConfig struct {
	OriginalPath string
	IngressClass string
}

func resolvePathRewriteIngressController(ingressClass string) string {
	ingressClass = strings.ToLower(ingressClass)
	switch {
	case ingressClass == "", strings.Contains(ingressClass, IngressClassNginx):
		return IngressClassNginx
	case strings.Contains(ingressClass, IngressClassHAProxy):
		return IngressClassHAProxy
	default:
		return ""
	}
}

// AddPathRewriteIngressRegex adds the regex pattern to the path for Nginx-style path rewriting
// when using path-based routing. Returns the modified path and the appropriate path type.
func AddPathRewriteIngressRegex(originalPath string, enablePathRewrite bool, ingressClass string) (string, networkingv1.PathType) {
	if originalPath == "" || originalPath == "/" {
		return "/", networkingv1.PathTypePrefix
	}

	pathPrefix := strings.TrimSuffix(originalPath, "/")
	if !enablePathRewrite || resolvePathRewriteIngressController(ingressClass) != IngressClassNginx {
		return pathPrefix, networkingv1.PathTypePrefix
	}

	// Add regex pattern for path rewriting: /path(/|$)(.*)
	// This allows Nginx to strip the path prefix before forwarding to the backend
	ingressPath := pathPrefix + "(/|$)(.*)"
	return ingressPath, networkingv1.PathTypeImplementationSpecific
}

// AddPathRewriteIngressAnnotations adds the appropriate path rewrite annotations based on the ingress controller type
func AddPathRewriteIngressAnnotations(annotations map[string]string, config PathRewriteConfig) map[string]string {
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Only add path rewrite annotations when using path-based routing and path rewrite is enabled
	if config.OriginalPath == "" || config.OriginalPath == "/" {
		return annotations
	}

	// Determine which annotations to add based on ingress class
	switch resolvePathRewriteIngressController(config.IngressClass) {
	case IngressClassNginx:
		annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$2"
	case IngressClassHAProxy:
		pathPrefix := strings.TrimSuffix(config.OriginalPath, "/")
		annotations["haproxy.org/path-rewrite"] = fmt.Sprintf("%s(/|$)(.*) /\\2", pathPrefix)
	}

	return annotations
}

// AddPathRewriteRouteAnnotations adds path rewrite annotations for OpenShift routes
func AddPathRewriteRouteAnnotations(annotations map[string]string, path string, enablePathRewrite bool) map[string]string {
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Only add path rewrite annotation when using path-based routing and path rewrite is enabled
	if path != "" && path != "/" && enablePathRewrite {
		// OpenShift HAProxy router - rewrite to root path
		annotations["haproxy.router.openshift.io/rewrite-target"] = "/\\1"
	}

	return annotations
}

// StripPathRewritePattern removes the Nginx path rewrite regex pattern from a path
// This is used when displaying URLs to users
func StripPathRewritePattern(path string) string {
	if strings.Contains(path, "(/|$)(.*)") {
		i := strings.Index(path, "(/|$)(.*)")
		return path[:i]
	}

	return path
}
