/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"

	audit "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client"
	auditconfig "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
)

const (
	url = "http://test.com"
)

var (
	defaultValues = map[runaictx.ContextField]string{
		runaictx.CfTraceId:            "123",
		runaictx.CfIsManagementTenant: "false",
		runaictx.CfTenantId:           "12333",
		runaictx.CfClientId:           "798789",
		runaictx.CfClientType:         "clientType",
		runaictx.CfSubjectID:          "subjectId",
		runaictx.CfSubjectType:        "user",
		runaictx.CfGroups:             "groups",
		runaictx.CfSupplementalGroups: "sup",
		runaictx.CfRequestOrigin:      "orig",
	}
	mockAuditLog = openapi.AuditLog{
		Id:         stringPtr("550e8400-e29b-41d4-a716-446655440000"),
		Subject:    "1",
		TenantId:   "2",
		HttpMethod: stringPtr("GET"),
		SourceIp:   stringPtr("10.1.20.1"),
		Metadata:   &map[string]string{"key": "val"},
	}
)

func stringPtr(s string) *string {
	return &s
}

func TestAuditBuilder_Success(t *testing.T) {
	// In middleware code
	ctx := buildRunaiContext(defaultValues)
	req := createHttpRequest()
	al := audit.NewAuditLog().WithRequest(req).WithCtx(ctx)
	ctx = audit.CtxWithAudit(ctx, al)

	// Following lines demonstrate how to capture the http status code
	// rw should be the real writer which is provided to the middleware
	var rw http.ResponseWriter
	capturer := audit.NewResponseWriterCapture(rw)

	// here is the API code
	cb := audit.AuditFromCtx(ctx)
	cb.WithEntityId("EntityId").
		WithEntityName("EntityName").
		WithEntityType("EntityType").
		WithClusterName("ClusterName").
		WithClusterId("ClusterId").
		WithAction("Action").WithStatusCode(capturer.StatusCode).
		WithMetadata("key", "value")

	// here the middleware part
	log := al.Log
	require.NotEmpty(t, log.Id)
	require.NotEmpty(t, log.Timestamp)
	require.Equal(t, "12333", log.TenantId)
	require.Equal(t, "subjectId", log.Subject)
	require.Equal(t, "user", log.SubjectType)
	require.Equal(t, "1.1.1.1", *log.SourceIp)
	require.Equal(t, "Action", log.Action)

	require.Equal(t, "POST", *log.HttpMethod)
	require.Equal(t, url, *log.Url)
	require.Equal(t, "EntityType", *log.EntityType)
	require.Equal(t, "EntityName", *log.EntityName)
	require.Equal(t, "EntityId", *log.EntityId)
	require.Equal(t, "Succeeded", log.Result)
	require.Equal(t, int32(200), *log.HttpStatusCode)
	require.Equal(t, "ClusterName", *log.ClusterName)
	require.Equal(t, "ClusterId", *log.ClusterId)
	require.Equal(t, "123", *log.RequestId)
	require.Equal(t, map[string]string{"key": "value", auditconfig.ClientIdMetadataKey: "798789"}, *log.Metadata)

	require.Equal(t, 18, reflect.ValueOf(openapi.AuditLog{}).NumField(),
		"Number of AuditLog fields has changed. Dont forget to test the new ones!")
}

func buildRunaiContext(values map[runaictx.ContextField]string) context.Context {
	ctx := context.Background()
	for key, val := range values {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

func createHttpRequest() *http.Request {
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.1.1.1"
	return req
}

func TestPrintLogToStdout(t *testing.T) {
	originalStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		os.Stdout = originalStdout
	}()

	audit.PrintLogToStdout(&mockAuditLog)

	w.Close()
	out, _ := io.ReadAll(r)

	capturedOutput := string(out)

	var output map[string]interface{}
	err := json.Unmarshal([]byte(capturedOutput), &output)

	require.NoError(t, err)
	require.Equal(t, output["type"], "audit")
	require.Equal(t, output["id"], *mockAuditLog.Id)
	require.Equal(t, output["subject"], mockAuditLog.Subject)
	require.Equal(t, output["tenant_id"], mockAuditLog.TenantId)
	require.Equal(t, output["http_method"], *mockAuditLog.HttpMethod)
	require.Equal(t, output["source_ip"], *mockAuditLog.SourceIp)
}

func TestGetClientIp_WithRemoteAddr(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.RemoteAddr = "1.1.1.1"

	al := audit.NewAuditLog().WithRequest(req)

	log := al.Log
	require.Equal(t, "1.1.1.1", *log.SourceIp)
}

func TestGetClientIp_WithXForwardedOverrides(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	req.RemoteAddr = "2.2.2.2"

	al := audit.NewAuditLog().WithRequest(req)

	log := al.Log
	require.Equal(t, "1.1.1.1", *log.SourceIp)
}

func TestGetClientIp_WithXRealIpOverrides(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("X-Real-IP", "1.1.1.1")
	req.RemoteAddr = "2.2.2.2"

	al := audit.NewAuditLog().WithRequest(req)

	log := al.Log
	require.Equal(t, "1.1.1.1", *log.SourceIp)
}

func TestGetClientIp_WithForwarded(t *testing.T) {
	testCases := []struct {
		headerValue string
		expected    string
	}{
		{"for=192.0.2.43, for=198.51.100.17", "192.0.2.43, 198.51.100.17"},
		{"for=192.0.2.60;proto=http;by=203.0.113.43", "192.0.2.60"},
		{"For=\"[2001:db8:cafe::17]:4711\"", "2001:db8:cafe::17"},
		{"by=<identifier>;for=<identifier>;host=<host>;proto=<http|https>", "<identifier>"},
		{"for=\"_obfuscated_id\"", "_obfuscated_id"},
		{"For=\"[2001:db8:cafe::17]:4711\"", "2001:db8:cafe::17"},
		{"for=192.0.2.43,for=198.51.100.17;by=203.0.113.60;proto=http;host=example.com", "192.0.2.43, 198.51.100.17"},
		{"for=192.0.2.43,for=\"[2001:db8:cafe::17]\",for=unknown;by=203.0.113.60;proto=http;host=example.com", "192.0.2.43, 2001:db8:cafe::17, unknown"},
		{"", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.headerValue, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
			req.Header.Set("Forwarded", tc.headerValue)

			al := audit.NewAuditLog().WithRequest(req)
			log := al.Log

			require.Equal(t, tc.expected, *log.SourceIp)
		})
	}
}

func TestGetClientIp_WithXForwardedFor(t *testing.T) {
	testCases := []struct {
		headerValue string
		expected    string
	}{
		{"2001:db8:85a3:8d3:1319:8a2e:370:7348", "2001:db8:85a3:8d3:1319:8a2e:370:7348"},
		{"203.0.113.195", "203.0.113.195"},
		{"203.0.113.195, 2001:db8:85a3:8d3:1319:8a2e:370:7348", "203.0.113.195, 2001:db8:85a3:8d3:1319:8a2e:370:7348"},
		{"<client>, <proxy>", "<client>, <proxy>"},
		{"127.0.0.1", "unknown"},
		{"", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.headerValue, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, "http://example.com", nil)
			req.Header.Set("X-Forwarded-For", tc.headerValue)

			al := audit.NewAuditLog().WithRequest(req)
			log := al.Log

			require.Equal(t, tc.expected, *log.SourceIp)
		})
	}
}
