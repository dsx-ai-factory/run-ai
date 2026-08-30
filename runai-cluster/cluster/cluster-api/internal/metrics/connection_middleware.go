package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/router"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/run-ai/runai/runai-common-packages/metrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/rs/zerolog/log"
)

const (
	FollowQuery      = "follow"
	TtyQuery         = "tty"
	InteractiveQuery = "interactive"
)

var BaseMap = map[string]func(r *http.Request) Interactivity{
	router.PathStatus:       NoneInteractiveFunc,
	router.PathPortforward:  PortForwardInteractivityFunc,
	router.PathLogs:         LogsInteractivityFunc,
	router.PathLogsDownload: NoneInteractiveFunc,
	router.PathAttach:       ExecInteractivityFunc,
	router.PathExec:         ExecInteractivityFunc,
}

var once sync.Once

// ConnectionMiddleware measures and logs the duration of HTTP transactions
type ConnectionMiddleware struct {
	enabled bool
}

// NewConnectionMiddleware creates a new instance of ConnectionMiddleware
func NewConnectionMiddleware(enabled bool) *ConnectionMiddleware {
	register()
	return &ConnectionMiddleware{enabled: enabled}
}

// Middleware returns the HTTP middleware function that measures transaction duration
func (m *ConnectionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response writer hijacker wrapper to allow connection upgrade
		ww := NewResponseHijackerWrapper(w)

		if !m.enabled {
			// disabled, write and finish
			next.ServeHTTP(ww, r)
			return
		}

		start := time.Now()
		ctx := r.Context()

		workloadFromCtx := workload.FromContext(ctx)
		if workloadFromCtx == nil {
			log.Warn().Msgf("no context found, URI: %s", r.RequestURI)
			// no context, we bypass the metrics writing
			next.ServeHTTP(ww, r)

			return
		}

		pathStr := getBasePath(r)
		interactivity := getInteractivity(pathStr, r)

		meta := &EndpointMeta{
			Path:              pathStr,
			Interactivity:     interactivity,
			WorkloadType:      workloadFromCtx.Type,
			WorkloadFramework: workloadFromCtx.JobType,
			TenantId:          getTenantIdFromRequest(r),
		}
		ctx = SetIntoContext(ctx, meta)
		updated := r.WithContext(ctx)
		active := NewActive(meta)

		active.Inc()
		defer active.Dec()
		// Call the next handler
		next.ServeHTTP(ww, updated)

		// Calculate duration
		duration := time.Since(start)

		life := NewLifecycleDuration(meta)
		life.Report(strconv.Itoa(ww.StatusCode), GetReasonFromHeader(w), duration.Seconds())

		// Also log with zerolog for structured logging
		log.Info().
			Str("method", updated.Method).
			Str("url_path", updated.URL.Path).
			Str("base_path", meta.Path).
			Str("remote_addr", updated.RemoteAddr).
			Int("status_code", ww.StatusCode).
			Dur("duration", duration).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("websocket connection duration")
	})
}

func register() {
	once.Do(func() {
		active := NewActive(&EndpointMeta{})
		life := NewLifecycleDuration(&EndpointMeta{})

		active.Register()
		life.Register()
		ConnectionMessageRegister()
	})
}

func Init(enabled bool, port int) error {
	metricsConfig := metrics.Config{Enabled: enabled, ExposePort: port}
	_, err := metrics.Init(&metricsConfig)

	return err
}

func getBasePath(r *http.Request) string {
	endpoint := api.GetEndpointPath(r)

	endpoint = fmt.Sprintf("/%s", endpoint)
	if _, ok := BaseMap[endpoint]; ok {
		return endpoint
	}

	return ""
}

func getInteractivity(path string, r *http.Request) Interactivity {
	if f, ok := BaseMap[path]; ok {
		return f(r)
	}

	return NoneInteractive
}

func getTenantIdFromRequest(r *http.Request) string {
	tenantId, _ := runaictx.ParseTenantId(r.Context())
	return tenantId
}

func PortForwardInteractivityFunc(_ *http.Request) Interactivity {
	return FullInteractive
}

func LogsInteractivityFunc(r *http.Request) Interactivity {
	follow, err := getBoolFromQueryParam(r, FollowQuery)
	if err != nil {
		return NoneInteractive
	}

	if follow {
		return HalfInteractive
	}

	return NoneInteractive
}

func NoneInteractiveFunc(_ *http.Request) Interactivity {
	return NoneInteractive
}

func ExecInteractivityFunc(r *http.Request) Interactivity {
	tty, _ := getBoolFromQueryParam(r, TtyQuery)
	interactive, _ := getBoolFromQueryParam(r, InteractiveQuery)

	if tty || interactive {
		return FullInteractive
	}

	return HalfInteractive
}

func getBoolFromQueryParam(r *http.Request, value string) (bool, error) {
	query := r.URL.Query().Get(value)

	parsed, err := strconv.ParseBool(query)
	if err != nil {
		return false, err
	}

	return parsed, nil
}
