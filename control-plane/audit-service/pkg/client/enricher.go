package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/google/uuid"
	auditconfig "github.com/run-ai/runai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
)

type Enricher interface {
	WithRequestId(string) *AuditLog
	WithTenantId(string) *AuditLog
	WithSubjectType(string) *AuditLog
	WithSubject(string) *AuditLog
	WithSourceIp(string) *AuditLog
	WithUrl(string) *AuditLog
	WithHttpMethod(string) *AuditLog
	WithTimestamp(time.Time) *AuditLog
	WithAction(string) *AuditLog
	WithEntityId(string) *AuditLog
	WithEntityType(string) *AuditLog
	WithEntityName(string) *AuditLog
	WithClusterName(string) *AuditLog
	WithClusterId(string) *AuditLog
	WithStatusCode(int) *AuditLog
	WithMetadata(string, string) *AuditLog
	WithRequest(*http.Request) *AuditLog
}

type AuditLog struct {
	Log  openapi.AuditLog
	drop bool // provides the option to instruct the middleware, not to send the Log
}

func NewAuditLog() *AuditLog {
	return &AuditLog{Log: openapi.AuditLog{
		Id:        ptr.String(uuid.New().String()),
		Timestamp: ptr.Time(time.Now()),
	}, drop: false}
}

func (al *AuditLog) WithSubject(v string) *AuditLog {
	al.Log.SetSubject(v)
	return al
}

func (al *AuditLog) WithAction(v string) *AuditLog {
	al.Log.SetAction(v)
	return al
}

func (al *AuditLog) WithEntityType(v string) *AuditLog {
	al.Log.SetEntityType(v)
	return al
}

func (al *AuditLog) WithEntityName(v string) *AuditLog {
	al.Log.SetEntityName(v)
	return al
}

func (al *AuditLog) WithEntityId(v string) *AuditLog {
	al.Log.SetEntityId(v)
	return al
}

func (al *AuditLog) WithClusterName(v string) *AuditLog {
	al.Log.SetClusterName(v)
	return al
}

func (al *AuditLog) WithClusterId(v string) *AuditLog {
	al.Log.SetClusterId(v)
	return al
}

func (al *AuditLog) WithRequestId(s string) *AuditLog {
	al.Log.SetRequestId(s)
	return al
}

func (al *AuditLog) WithTenantId(tenantId string) *AuditLog {
	al.Log.SetTenantId(tenantId)
	return al
}

func (al *AuditLog) WithSubjectType(subjectType string) *AuditLog {
	al.Log.SetSubjectType(subjectType)
	return al
}

func (al *AuditLog) WithSourceIp(ip string) *AuditLog {
	al.Log.SetSourceIp(ip)
	return al
}

func (al *AuditLog) WithUrl(url string) *AuditLog {
	al.Log.SetUrl(url)
	return al
}

func (al *AuditLog) WithHttpMethod(method string) *AuditLog {
	al.Log.SetHttpMethod(method)
	return al
}

func (al *AuditLog) WithTimestamp(t time.Time) *AuditLog {
	al.Log.SetTimestamp(t)
	return al
}

func (al *AuditLog) WithMetadata(key, value string) *AuditLog {
	currentMetadata := al.Log.GetMetadata()
	if currentMetadata == nil {
		currentMetadata = make(map[string]string)
	}

	currentMetadata[key] = value
	al.Log.SetMetadata(currentMetadata)
	return al
}

func (al *AuditLog) WithStatusCode(responseStatusCode int) *AuditLog {
	al.Log.SetHttpStatusCode(int32(responseStatusCode))
	al.Log.SetResult(statusCodeToResult(responseStatusCode))
	return al
}

func (al *AuditLog) WithRequest(r *http.Request) *AuditLog {
	return al.
		WithHttpMethod(r.Method).
		WithAction(concludeAction(r)).
		WithUrl(r.URL.String()).
		WithSourceIp(getClientIP(r))
}

// concludeAction attempts to conclude action types from request.
// Supported actions are:
//   - create
//   - update
//   - delete
//
// TODO: can we conclude here from URL?
//   - sign-in
//   - sign-out
//   - reset-password
func concludeAction(r *http.Request) string {
	switch r.Method {
	case http.MethodGet:
		return "Get"
	case http.MethodPost:
		return "Create"
	case http.MethodPut:
		return "Update"
	case http.MethodPatch:
		return "Update"
	case http.MethodDelete:
		return "Delete"
	default:
		logging.Warnf("Failed mapping method %q for url %q", r.Method, r.URL.String())
	}
	return ""
}

func (al *AuditLog) WithCtx(ctx context.Context) *AuditLog {
	clientId := extractContextField(ctx, runaictx.CfClientId, "")
	if clientId != "" {
		al.WithMetadata(auditconfig.ClientIdMetadataKey, clientId)
	}

	return al.
		WithTenantId(extractContextField(ctx, runaictx.CfTenantId, "")).
		WithSubject(extractContextField(ctx, runaictx.CfSubjectID, "")).
		WithSubjectType(extractContextField(ctx, runaictx.CfSubjectType, "")).
		WithRequestId(extractContextField(ctx, runaictx.CfTraceId, ""))
}

func extractContextField(ctx context.Context, cfId runaictx.ContextField, defaultVal string) string {
	if value, ok := ctx.Value(cfId).(string); ok {
		return value
	}
	return defaultVal
}

func (al *AuditLog) Drop() {
	al.drop = true
}

func (al *AuditLog) Write() error {
	if al.drop {
		return nil
	}

	err := validate(&al.Log)
	if err != nil {
		return err
	}

	if enablePrintAuditLogsToStdout {
		PrintLogToStdout(&al.Log)
	}

	return singleInstance.Write(al.Log)
}

type AuditWrap struct {
	*openapi.AuditLog
	Type string `json:"type"`
}

// This method is important- it's necessary to override openapi.AuditLog's MarshalJSON
func (aw *AuditWrap) MarshalJSON() ([]byte, error) {
	return json.Marshal(aw)
}

func PrintLogToStdout(log *openapi.AuditLog) {
	logToPrint := AuditWrap{
		AuditLog: log,
		Type:     "audit",
	}

	jsonData, err := json.Marshal(logToPrint)
	if err != nil {
		logging.Warnf("Error marshaling audit log: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

func statusCodeToResult(statusCode int) string {
	if 200 <= statusCode && statusCode < 300 {
		return "Succeeded"
	}

	return "Failed"
}

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (may contain multiple IPs)
	success, ip := getIpFromXForwarded(r)
	if success {
		return ip
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" && isValidIP(realIP) {
		return realIP
	}

	// Check Forwarded header (RFC 7239)
	success, s := getIpFromForwarded(r)
	if success {
		return s
	}

	// Fallback to RemoteAddr
	if isValidIP(r.RemoteAddr) {
		return r.RemoteAddr
	}

	return "unknown"
}

func getIpFromForwarded(r *http.Request) (bool, string) {
	forwarded := r.Header.Get("Forwarded")
	re := regexp.MustCompile(`(For=|for=)\"?\[?([^\[\]\",;\n]+)\"?\[?`)
	matches := re.FindAllStringSubmatch(forwarded, -1)
	if matches == nil {
		return false, ""
	}

	allIp := []string{}
	for _, match := range matches {
		ip := match[2]
		if isValidIP(ip) {
			allIp = append(allIp, ip)
		}
	}

	if len(allIp) == 0 {
		return false, ""
	}

	return true, strings.Join(allIp, ", ")
}

func getIpFromXForwarded(r *http.Request) (bool, string) {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor == "" {
		return false, ""
	}

	allIp := []string{}
	ips := strings.Split(forwardedFor, ",")
	for _, ip := range ips {
		trimmed := strings.TrimSpace(ip)
		if isValidIP(trimmed) {
			allIp = append(allIp, trimmed)
		}
	}

	if len(allIp) == 0 {
		return false, ""
	}

	return true, strings.Join(allIp, ", ")
}

// isValidIP ensures the extracted IP is not empty and not a local or private address
func isValidIP(ip string) bool {
	return ip != "" && ip != "127.0.0.1" && ip != "::1"
}
