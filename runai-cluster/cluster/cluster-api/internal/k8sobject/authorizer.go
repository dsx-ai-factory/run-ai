package k8sobject

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	workloadsapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

// AuthorizedWorkloadInfo is the minimal projection of WorkloadDetailed that downstream
// k8s-object handlers need to resolve the underlying Kubernetes object. Keep this
// lean — anything extra inflates the cache footprint per (subject, workload) pair.
type AuthorizedWorkloadInfo struct {
	Namespace string
	Name      string
	Kind      string
	Group     string
	Version   string

	SourceApi            string
	DistributedFramework string
}

// WorkloadAuthorizationResponse holds the outcome of one upstream GetWorkload call.
// On 200: Workload is set, HTTPStatus=200.
// On 4xx: Workload is nil; HTTPStatus + UpstreamMsg carry the relayed response.
type WorkloadAuthorizationResponse struct {
	Workload    *AuthorizedWorkloadInfo
	HTTPStatus  int
	UpstreamMsg string
}

// APIWorkloadsGetter fetches a workload by ID and returns the minimal projection
// the authorizer needs, plus the raw HTTP response so callers can inspect status
// and body on non-2xx.
type APIWorkloadsGetter interface {
	GetWorkloadById(ctx context.Context, workloadID string) (*AuthorizedWorkloadInfo, *http.Response, error)
}

type apiWorkloadsGetter struct {
	api workloadsapi.WorkloadsAPI
}

// NewAPIWorkloadsGetter returns the canonical APIWorkloadsGetter that fetches
// via the generated workloads-client.
func NewAPIWorkloadsGetter(api workloadsapi.WorkloadsAPI) APIWorkloadsGetter {
	return &apiWorkloadsGetter{api: api}
}

// GetWorkloadById requires ctx to carry workloadsapi.ContextAccessToken — the
// generated client reads it to populate Authorization: Bearer. The upstream
// WorkloadDetailed is projected to AuthorizedWorkloadInfo here so callers
// never see the generated type.
func (g *apiWorkloadsGetter) GetWorkloadById(ctx context.Context, workloadID string) (*AuthorizedWorkloadInfo, *http.Response, error) {
	wl, httpResp, err := g.api.GetWorkload(ctx, workloadID).Execute()
	if err != nil || wl == nil {
		return nil, httpResp, err
	}
	// kubernetesRef is populated only for External workloads.
	// Group/Version stay empty for the five first-class native types whose
	// GVK is resolved by a static map in the handler.
	info := &AuthorizedWorkloadInfo{
		Namespace:            wl.Namespace,
		Name:                 wl.Name,
		Kind:                 wl.Type,
		SourceApi:            string(wl.GetSourceApi()), // "" when unset (Nullable)
		DistributedFramework: wl.GetDistributedFramework(),
	}
	if ref, ok := wl.GetKubernetesRefOk(); ok && ref != nil {
		info.Group = ref.GetGroup()
		info.Version = ref.GetVersion()
	}
	return info, httpResp, nil
}

type WorkloadAuthorizer interface {
	Authorize(ctx context.Context, subject, workloadID string) (*WorkloadAuthorizationResponse, error)
}

// CachedWorkloadAuthorizer caches:
//   - 200 OK (stable: workload exists and caller can read it)
//   - 4xx (stable answers: bad token, denied, gone, etc.)
//
// and does NOT cache 5xx or network errors (transient — should retry next call).
// Note: 408 Request Timeout and 429 Too Many Requests are technically transient
// but are still cached here for simplicity; workloads-service is internal and
// rarely returns those, so a brief 60s cache window is acceptable.
type CachedWorkloadAuthorizer struct {
	getter APIWorkloadsGetter
}

// NewCachedWorkloadAuthorizer registers the package's cache store with the given
// TTL and returns a ready authorizer. Production callers pass config.WorkloadAuthCacheTTL;
// tests may pass shorter values to exercise expiration without sleeping.
func NewCachedWorkloadAuthorizer(getter APIWorkloadsGetter, ttl time.Duration) (*CachedWorkloadAuthorizer, error) {
	if err := cache.RegisterStore[WorkloadAuthorizationResponse](options.Ttl(ttl)); err != nil {
		return nil, err
	}
	return &CachedWorkloadAuthorizer{getter: getter}, nil
}

func (a *CachedWorkloadAuthorizer) Authorize(ctx context.Context, subject, workloadID string) (*WorkloadAuthorizationResponse, error) {
	key := cacheKey(subject, workloadID)

	// Fail closed on cache errors: silent fall-through would let a broken cache
	// dump 100% of requests onto control-plane without any visible signal. Logs
	// carry only the workloadID (UUID, non-PII) — the JWT sub claim is omitted
	// because IdPs can map it to email or other PII.
	res, err := cache.Get[WorkloadAuthorizationResponse](ctx, key)
	if err != nil {
		logging.Errorf("k8sobject: cache get failed for workload %s: %v", workloadID, err)
		return nil, fmt.Errorf("k8sobject: cache get: %w", err)
	}
	if res != nil {
		return res, nil
	}

	workloadInfo, httpResp, err := a.getter.GetWorkloadById(ctx, workloadID)
	if httpResp != nil && httpResp.Body != nil {
		defer func() { _ = httpResp.Body.Close() }()
	}

	// The generated client returns err != nil for any status >= 300 (3xx/4xx/5xx)
	// or network failure. Only 4xx is a stable, per-subject denial we cache;
	// everything else is transient and surfaces as an error.
	not4xx := httpResp == nil || httpResp.StatusCode < 400 || httpResp.StatusCode >= 500
	if err != nil && not4xx {
		return nil, fmt.Errorf("k8sobject: GetWorkloadById: %w", err)
	}

	// Guard against unexpected 2xx other than 200 (e.g., a future 204): the
	// generated client returns nil error for any 2xx with a decoded body, but
	// the workloads-service spec only defines 200 for GET /workloads/{id}.
	if err == nil && (httpResp == nil || httpResp.StatusCode != http.StatusOK || workloadInfo == nil) {
		status := 0
		if httpResp != nil {
			status = httpResp.StatusCode
		}
		return nil, fmt.Errorf("k8sobject: unexpected upstream success state (status=%d, info-nil=%v)", status, workloadInfo == nil)
	}

	authorizationResponse := WorkloadAuthorizationResponse{
		Workload:   workloadInfo,
		HTTPStatus: httpResp.StatusCode,
	}
	if err != nil { // 4xx — cache the denial with the generated client's formatted message.
		authorizationResponse.UpstreamMsg = err.Error()
	}

	if setErr := cache.Set(ctx, key, authorizationResponse); setErr != nil {
		logging.Errorf("k8sobject: cache set failed for workload %s: %v", workloadID, setErr)
		return nil, fmt.Errorf("k8sobject: cache set: %w", setErr)
	}
	return &authorizationResponse, nil
}

// cacheKey scopes per-caller so a cached 200 for one user doesn't leak to another
// after that other user has had their permission revoked. Both components are
// url.QueryEscaped so the "|" separator is collision-safe even when the JWT sub
// claim contains "|" (IdPs can be configured to put arbitrary text there, e.g.
// sub = ${email}|${department}).
func cacheKey(subject, workloadID string) string {
	return fmt.Sprintf("%s|%s", url.QueryEscape(subject), url.QueryEscape(workloadID))
}
