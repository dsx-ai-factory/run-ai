package redirect

import (
	_ "embed"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/run-ai/runai/runai-common-packages/logging"
)

//go:embed assets/error_page.html
var errorPageHTML string

const clientIDPattern = `^runai-internal\.([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})\.workloads_auth$`

var clientIDRegex = regexp.MustCompile(clientIDPattern)

type Handler struct {
	clusterClient ClusterURLProvider
}

func NewHandler(clusterServiceURL string) *Handler {
	return &Handler{clusterClient: newClusterServiceClient(clusterServiceURL)}
}

// ValidateRedirect is the forwardAuth endpoint for the workload-redirect-validation Traefik middleware.
// Traefik forwardAuth passes non-2xx responses (body included) directly to the client, so returning
// 400 + HTML here reaches the browser without a separate error-page middleware.
func (h *Handler) ValidateRedirect(w http.ResponseWriter, r *http.Request) {
	redirectURI, clusterUUID, err := extractParams(r)
	if err != nil {
		logging.Errorf("workload-redirect-validation: invalid request: %v", err)
		serveErrorPage(w, err.Error())
		return
	}

	clusterURL, err := h.clusterClient.GetClusterURL(r.Context(), clusterUUID)
	if err != nil {
		logging.Errorf("workload-redirect-validation: %v", err)
		serveErrorPage(w, "Failed to get cluster URL")
		return
	}

	if !isValidRedirect(strings.ToLower(redirectURI), strings.ToLower(clusterURL)) {
		logging.Errorf("workload-redirect-validation: redirect_uri %q rejected for cluster %q", redirectURI, clusterURL)
		serveErrorPage(w, "Invalid parameter: redirect_uri")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func serveErrorPage(w http.ResponseWriter, message string) {
	html := strings.ReplaceAll(errorPageHTML, "{{error_message}}", message)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadRequest)
	_, _ = w.Write([]byte(html))
}

func extractParams(r *http.Request) (redirectURI, clusterUUID string, err error) {
	q := queryParams(r)

	redirectURI = q.Get("redirect_uri")
	if redirectURI == "" {
		return "", "", errMissing("redirect_uri")
	}

	clientID := q.Get("client_id")
	if clientID == "" {
		return "", "", errMissing("client_id")
	}

	matches := clientIDRegex.FindStringSubmatch(clientID)
	if len(matches) < 2 {
		return "", "", errInvalid("client_id")
	}
	return redirectURI, matches[1], nil
}

// queryParams returns the query parameters from the original request.
// Traefik forwardAuth strips query params from the forwarded request and puts
// the original URL in X-Forwarded-Uri, so we prefer that header when present.
func queryParams(r *http.Request) url.Values {
	if fwd := r.Header.Get("X-Forwarded-Uri"); fwd != "" {
		if u, err := url.ParseRequestURI(fwd); err == nil {
			return u.Query()
		}
	}
	return r.URL.Query()
}

func errMissing(param string) error {
	return &paramError{"Missing", param}
}

func errInvalid(param string) error {
	return &paramError{"Invalid", param}
}

type paramError struct {
	kind  string
	param string
}

func (e *paramError) Error() string {
	return e.kind + " parameter: " + e.param
}

func isValidRedirect(redirectURI, clusterURL string) bool {
	clusterParsed, err := url.Parse(clusterURL)
	if err != nil {
		logging.Errorf("workload-redirect-validation: failed to parse cluster URL %q: %v", clusterURL, err)
		return false
	}

	redirectParsed, err := url.Parse(redirectURI)
	if err != nil {
		logging.Errorf("workload-redirect-validation: failed to parse redirect_uri %q: %v", redirectURI, err)
		return false
	}

	clusterHost, clusterPort := splitHostPort(clusterParsed)
	redirectHost, redirectPort := splitHostPort(redirectParsed)

	return isSchemeMatching(clusterParsed, redirectParsed) &&
		isHostValid(clusterHost, redirectHost) &&
		isPortMatching(clusterPort, redirectPort) &&
		!hasPathTraversal(redirectParsed.Path)
}

func isSchemeMatching(cluster, redirect *url.URL) bool {
	return cluster.Scheme == redirect.Scheme
}

func isHostValid(clusterHost, redirectHost string) bool {
	if redirectHost == clusterHost || strings.HasSuffix(redirectHost, "."+clusterHost) {
		return true
	}
	parts := strings.Split(clusterHost, ".")
	if len(parts) >= 3 {
		parent := strings.Join(parts[1:], ".")
		if strings.HasSuffix(redirectHost, "."+parent) {
			return true
		}
	}
	return false
}

func isPortMatching(clusterPort, redirectPort string) bool {
	return clusterPort == redirectPort
}

func hasPathTraversal(path string) bool {
	return strings.Contains(path, "..")
}

func splitHostPort(u *url.URL) (host, port string) {
	h, p, _ := net.SplitHostPort(u.Host)
	if p == "" {
		p = defaultPort(u.Scheme)
	}
	if h == "" {
		h = u.Host
	}
	return h, p
}

func defaultPort(scheme string) string {
	switch scheme {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return ""
	}
}
