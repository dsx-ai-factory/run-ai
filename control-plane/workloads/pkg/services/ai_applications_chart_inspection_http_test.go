/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	authorizationclient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	authorizationinternal "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/clientinternal"
	authorizationapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	chartinspection "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/chart_inspection"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	runaiconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/google/uuid"
)

func TestChartInspectionRouteReauthorizesWarmResultsThroughRealClients(t *testing.T) {
	cachetesting.EnableCache()
	t.Cleanup(cachetesting.DisableCache)
	var allowApplication, allowCredential atomic.Bool
	allowApplication.Store(true)
	allowCredential.Store(true)
	var permissionChecks, credentialChecks atomic.Int32
	var expectedSubject atomic.Value
	expectedSubject.Store("alice")
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Runai-Tenant-Id") != "1" || r.Header.Get("X-Runai-Subject-Id") != expectedSubject.Load().(string) {
			t.Error("trusted caller identity was not forwarded to authorization/Assets")
		}
		switch r.URL.Path {
		case "/api/v1/authorization/evaluate":
			permissionChecks.Add(1)
			var evaluation map[string]any
			if err := json.NewDecoder(r.Body).Decode(&evaluation); err != nil {
				t.Error(err)
			}
			if evaluation["resourceType"] != "applications" || evaluation["requestAction"] != "create" {
				t.Errorf("wrong route permission: %v", evaluation)
			}
			if !allowApplication.Load() {
				http.Error(w, "application permission denied", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case "/internal/api/v1/asset/ngc-api-key/11111111-1111-4111-8111-111111111111":
			credentialChecks.Add(1)
			w.Header().Set("Content-Type", "application/json")
			if !allowCredential.Load() {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":403,"message":"credential access denied"}`))
				return
			}
			_, _ = w.Write([]byte(`{"key":"synthetic-test-key"}`))
		default:
			t.Errorf("unexpected upstream request: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(upstream.Close)
	oldServices := runaiconfig.Get[*config.ServicesConfig]()
	servicesConfig := *oldServices
	servicesConfig.AssetsServiceUrl = upstream.URL
	if err := runaiconfig.Set(&servicesConfig); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := runaiconfig.Set(oldServices); err != nil {
			t.Error(err)
		}
	})
	oldAuthorization := authorizationclient.NewAuthorizationClient()
	authConfig := authorizationapi.NewConfiguration()
	authConfig.Servers = authorizationapi.ServerConfigurations{{URL: upstream.URL}}
	authorizationclient.SetMockAuthorizationClient(&authorizationinternal.AuthorizationClient{PermissionsAPI: authorizationapi.NewAPIClient(authConfig).PermissionsAPI})
	t.Cleanup(func() { authorizationclient.SetMockAuthorizationClient(oldAuthorization) })
	inspector := newApplicationChartInspector()
	archive := inspectionTestArchive(t)
	downloads := 0
	inspector.inspector = chartinspection.NewInspector(func(_ context.Context, _ chartinspection.Selection, key string) ([]byte, *api.Error) {
		downloads++
		if key != "synthetic-test-key" {
			t.Error("saved credential did not reach acquisition")
		}
		return archive, nil
	}, 0)
	controller := api.NewAIApplicationsAPIController(&AIApplicationsService{chartInspection: inspector})
	handler := controller.Routes()["InspectApplicationChart"].HandlerFunc
	for step, want := range []int{200, 200, 403, 403} {
		if step == 2 {
			allowCredential.Store(false)
		}
		if step == 3 {
			allowCredential.Store(true)
			allowApplication.Store(false)
		}
		request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", strings.NewReader(`{"chartRef":{"repo":"https://helm.ngc.nvidia.com/nvidia","chartName":"example","chartVersion":"1.0.0"},"auth":{"assetId":"11111111-1111-4111-8111-111111111111","userCredential":false}}`))
		subject := "alice"
		if step > 0 {
			subject = "bob"
		}
		expectedSubject.Store(subject)
		request = request.WithContext(inspectionTestContext("1", subject))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("step %d: status=%d want=%d: %s", step, response.Code, want, response.Body.String())
		}
	}
	if downloads != 1 || permissionChecks.Load() != 4 || credentialChecks.Load() != 3 {
		t.Fatalf("authorization/cache ordering: downloads=%d permission checks=%d credential checks=%d", downloads, permissionChecks.Load(), credentialChecks.Load())
	}
}

// Exercise the generated body decoder directly; middleware authorization is tested independently.
func TestChartInspectionHTTPRejectsUnexpectedFieldsBeforeSourceAccess(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"missing chart reference", `{}`},
		{"unknown top-level field", `{"chartRef":{"chartId":"11111111-1111-4111-8111-111111111111"},"defaultValues":{}}`},
		{"unknown chart field", `{"chartRef":{"chartId":"11111111-1111-4111-8111-111111111111","clusterId":"unexpected"}}`},
		{"inline credential material", `{"chartRef":{"chartURL":"https://example.com/chart.tgz"},"auth":{"key":"<test-only-key>"}}`},
		{"invalid credential flag", `{"chartRef":{"chartURL":"https://example.com/chart.tgz"},"auth":{"userCredential":"true"}}`},
		{"invalid body JSON", `{"chartRef":`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sourceAccesses := 0
			inspector := &applicationChartInspector{
				inspector: chartinspection.NewInspector(nil, time.Minute),
				loadUpload: func(context.Context, uuid.UUID, string) ([]byte, *api.Error) {
					sourceAccesses++
					return inspectionTestArchive(t), nil
				},
				loadKey: func(context.Context, string) (string, *api.Error) {
					sourceAccesses++
					return "", nil
				},
			}
			controller := api.NewAIApplicationsAPIController(&AIApplicationsService{chartInspection: inspector})
			request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", strings.NewReader(tc.body))
			request = request.WithContext(inspectionTestContext("1", "alice"))
			response := httptest.NewRecorder()
			controller.InspectApplicationChart(response, request)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
			}
			if sourceAccesses != 0 {
				t.Fatalf("invalid request accessed source %d times", sourceAccesses)
			}
		})
	}
}

func TestChartInspectionHTTPReturnsMetadataWithoutWholeDefaults(t *testing.T) {
	sourceAccesses := 0
	inspector := &applicationChartInspector{
		inspector: chartinspection.NewInspector(nil, time.Minute),
		loadUpload: func(context.Context, uuid.UUID, string) ([]byte, *api.Error) {
			sourceAccesses++
			return inspectionTestArchive(t), nil
		},
	}
	controller := api.NewAIApplicationsAPIController(&AIApplicationsService{chartInspection: inspector})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", strings.NewReader(`{"chartRef":{"chartId":"11111111-1111-4111-8111-111111111111"}}`))
	request.Header.Set("Accept", "application/json")
	request = request.WithContext(inspectionTestContext("1", "alice"))
	response := httptest.NewRecorder()
	controller.InspectApplicationChart(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected status %d: %s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"chartName", "chartVersion", "archiveSHA256", "hasQuestions", "questionsStatus", "groups", "questions", "warnings"} {
		if _, ok := body[required]; !ok {
			t.Fatalf("missing response field %s", required)
		}
	}
	for _, deferred := range []string{"defaultValues", "values", "valuesSchema", "questionsYaml", "auth"} {
		if _, ok := body[deferred]; ok {
			t.Fatalf("response included deferred/private field %s", deferred)
		}
	}
	if body["questionsStatus"] != "usable" || sourceAccesses != 1 {
		t.Fatalf("unexpected metadata or source access count: %v, %d", body, sourceAccesses)
	}
}

func TestChartInspectionHTTPDistinguishesDeclarationsAndAcquisitionFailures(t *testing.T) {
	usable := "questions:\n- variable: replicas\n  label: Replicas\n  type: int\n"
	unusable := usable + "- variable: replicas\n  label: Duplicate destination\n  type: int\n"
	malformed := "questions: ["
	for _, tc := range []struct {
		name      string
		questions *string
		fetchErr  *api.Error
		corrupt   bool
		code      int
		status    string
	}{
		{name: "absent", code: 200, status: "absent"},
		{name: "usable", questions: &usable, code: 200, status: "usable"},
		{name: "invalid declaration", questions: &unusable, code: 200, status: "unusable"},
		{name: "malformed questions", questions: &malformed, code: 200, status: "unusable"},
		{name: "download failure", fetchErr: api.NewError(http.StatusServiceUnavailable, "chart_download_failed: Chart download failed.", nil), code: http.StatusServiceUnavailable},
		{name: "upstream denial", fetchErr: api.NewError(http.StatusForbidden, "chart_download_denied: Chart access denied.", nil), code: 403},
		{name: "corrupt archive", corrupt: true, code: http.StatusBadRequest},
		{name: "unexpected internal failure", fetchErr: api.NewError(http.StatusInternalServerError, "chart inspection failed unexpectedly", errors.New("private internal diagnostic")), code: http.StatusInternalServerError},
		{name: "expired request", fetchErr: api.NewError(http.StatusRequestTimeout, "chart inspection request cancelled or expired", nil), code: 408},
	} {
		t.Run(tc.name, func(t *testing.T) {
			archive := inspectionTestArchiveWithQuestions(t, tc.questions)
			if tc.corrupt {
				archive = []byte("invalid gzip")
			}
			inspector := &applicationChartInspector{inspector: chartinspection.NewInspector(func(context.Context, chartinspection.Selection, string) ([]byte, *api.Error) {
				return archive, tc.fetchErr
			}, 0)}
			controller := api.NewAIApplicationsAPIController(&AIApplicationsService{chartInspection: inspector})
			request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", strings.NewReader(`{"chartRef":{"chartURL":"https://example.com/chart.tgz"}}`))
			request = request.WithContext(inspectionTestContext("1", "alice"))
			response := httptest.NewRecorder()
			controller.InspectApplicationChart(response, request)
			if response.Code != tc.code {
				t.Fatalf("status=%d, want=%d: %s", response.Code, tc.code, response.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.code != 200 {
				if tc.code == http.StatusInternalServerError && strings.Contains(response.Body.String(), "private internal diagnostic") {
					t.Fatal("unexpected internal failure leaked its details")
				}
				if _, exists := body["questionsStatus"]; exists {
					t.Fatal("inspection failure was returned as questions metadata")
				}
				return
			}
			if body["questionsStatus"] != tc.status || body["hasQuestions"] != (tc.questions != nil) {
				t.Fatalf("wrong declaration state: %v", body)
			}
			questions := body["questions"].([]any)
			warnings := body["warnings"].([]any)
			if tc.status == "unusable" && (len(questions) != 0 || len(warnings) == 0) {
				t.Fatal("unusable declaration did not discard the entire form with warnings")
			}
		})
	}
}
