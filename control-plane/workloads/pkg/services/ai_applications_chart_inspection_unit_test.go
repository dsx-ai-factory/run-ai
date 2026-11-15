/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	chartinspection "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/chart_inspection"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

func TestChartInspectionRejectsAmbiguousSources(t *testing.T) {
	for _, ref := range []api.AiApplicationSpecChartRef{
		{},
		{Repo: "https://helm.ngc.nvidia.com/nvidia", ChartName: "example"},
		{ChartURL: "https://example.com/chart.tgz", ChartId: "11111111-1111-4111-8111-111111111111"},
		{ChartURL: "https://example.com/chart.tgz", Repo: "https://example.com", ChartName: "example", ChartVersion: "1.0.0"},
		{ChartId: "not-a-uuid"},
		{ChartId: "00000000-0000-0000-0000-000000000000"},
	} {
		if _, e := chartInspectionSelection(ref); e == nil {
			t.Fatalf("accepted ambiguous/incomplete source: %#v", ref)
		}
	}
}

func TestChartInspectionReauthorizesCredentialsOnCacheHits(t *testing.T) {
	cachetesting.EnableCache()
	t.Cleanup(cachetesting.DisableCache)
	allowed := true
	downloads := 0
	c := &applicationChartInspector{
		inspector: chartinspection.NewInspector(func(context.Context, chartinspection.Selection, string) ([]byte, *api.Error) {
			downloads++
			return inspectionTestArchive(t), nil
		}, time.Minute),
		loadKey: func(context.Context, string) (string, *api.Error) {
			if !allowed {
				return "", api.NewError(http.StatusForbidden, "credential access denied", nil)
			}
			return "synthetic-test-key", nil
		},
	}
	req := api.AiApplicationChartInspectionRequest{
		ChartRef: api.AiApplicationSpecChartRef{Repo: "https://helm.ngc.nvidia.com/nvidia", ChartName: "example", ChartVersion: "1.0.0"},
		Auth:     api.AiApplicationAuth{AssetId: "11111111-1111-4111-8111-111111111111"},
	}
	for i := 0; i < 2; i++ {
		out, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(inspectionTestContext("1", "alice"), req)
		if err != nil || out.Code != 200 {
			t.Fatalf("inspection failed: %#v %v", out, err)
		}
	}
	if downloads != 1 {
		t.Fatalf("warm selection downloaded %d times", downloads)
	}
	allowed = false
	out, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(inspectionTestContext("1", "alice"), req)
	if err != nil || out.Code != 403 {
		t.Fatalf("cached metadata bypassed revoked access: %#v %v", out, err)
	}
}

func TestChartInspectionUploadsAreTenantScopedAndRechecked(t *testing.T) {
	present := true
	chartID := "11111111-1111-4111-8111-111111111111"
	c := &applicationChartInspector{
		inspector: chartinspection.NewInspector(nil, time.Minute),
		loadUpload: func(_ context.Context, id uuid.UUID, tenant string) ([]byte, *api.Error) {
			if !present || tenant != "1" || id.String() != chartID {
				return nil, api.NewError(http.StatusNotFound, "chart not found", nil)
			}
			return inspectionTestArchive(t), nil
		},
	}
	req := api.AiApplicationChartInspectionRequest{ChartRef: api.AiApplicationSpecChartRef{ChartId: chartID}}
	out, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(inspectionTestContext("1", "alice"), req)
	if err != nil || out.Code != 200 {
		t.Fatalf("owner inspection failed: %#v %v", out, err)
	}
	out, err = (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(inspectionTestContext("2", "alice"), req)
	if err != nil || out.Code != 404 {
		t.Fatalf("cross-tenant upload visible: %#v %v", out, err)
	}
	present = false
	out, err = (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(inspectionTestContext("1", "alice"), req)
	if err != nil || out.Code != 404 {
		t.Fatalf("deleted upload cached: %#v %v", out, err)
	}
}

func TestChartInspectionBoundsConcurrentUploadLookups(t *testing.T) {
	for _, requests := range []int{8, 32} {
		t.Run(fmt.Sprintf("%d requests", requests), func(t *testing.T) {
			archive := inspectionTestArchive(t)
			chartID := "11111111-1111-4111-8111-111111111111"
			ctx, cancel := context.WithCancel(inspectionTestContext("1", "alice"))
			defer cancel()
			release := make(chan struct{})
			var releaseOnce sync.Once
			unblock := func() { releaseOnce.Do(func() { close(release) }) }
			defer unblock()
			type lookupEvent struct {
				entered  bool
				response api.ImplResponse
				err      error
			}
			events := make(chan lookupEvent, 2*requests)
			var active, peak atomic.Int32
			c := &applicationChartInspector{
				inspector: chartinspection.NewInspector(nil, time.Minute),
				loadUpload: func(ctx context.Context, id uuid.UUID, tenant string) ([]byte, *api.Error) {
					if id.String() != chartID || tenant != "1" {
						return nil, api.NewError(http.StatusNotFound, "chart not found", nil)
					}
					current := active.Add(1)
					defer active.Add(-1)
					for {
						previous := peak.Load()
						if current <= previous || peak.CompareAndSwap(previous, current) {
							break
						}
					}
					events <- lookupEvent{entered: true}
					select {
					case <-release:
						return archive, nil
					case <-ctx.Done():
						return nil, api.NewError(http.StatusRequestTimeout, "upload lookup cancelled", nil)
					}
				},
			}
			request := api.AiApplicationChartInspectionRequest{ChartRef: api.AiApplicationSpecChartRef{ChartId: chartID}}
			start := make(chan struct{})
			for i := 0; i < requests; i++ {
				go func() {
					<-start
					response, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(ctx, request)
					events <- lookupEvent{response: response, err: err}
				}()
			}
			close(start)
			entered, rejected := 0, 0
			deadline := time.NewTimer(3 * time.Second)
			defer deadline.Stop()
			for i := 0; i < requests; i++ {
				select {
				case event := <-events:
					if event.entered {
						entered++
					} else {
						rejected++
						if event.err != nil || event.response.Code != http.StatusServiceUnavailable {
							t.Errorf("capacity response = %#v, error = %v", event.response, event.err)
						}
					}
				case <-deadline.C:
					t.Fatal("requests neither entered lookup nor received a capacity response")
				}
			}
			if current := peak.Load(); current > 4 || entered != 4 {
				t.Errorf("lookup concurrency exceeded its permit: peak=%d entered=%d, want 4", current, entered)
			}
			if rejected != requests-4 {
				t.Errorf("only %d/%d excess requests rejected before upload lookup completed, want %d", rejected, requests, requests-4)
			}
			// Finish the admitted lookups after every request has received its capacity decision.
			unblock()
			for completed := rejected; completed < requests; completed++ {
				select {
				case event := <-events:
					if event.entered {
						t.Fatal("late upload lookup was admitted after the initial capacity decision")
					}
					if event.err != nil || event.response.Code != http.StatusOK {
						t.Errorf("admitted inspection = %#v, error = %v", event.response, event.err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("admitted upload inspection did not finish after lookup release")
				}
			}
		})
	}
}

func TestChartInspectionUploadLookupPreservesAPIError(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusForbidden, http.StatusInternalServerError, http.StatusBadGateway} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			expected := api.NewError(status, "upload lookup failed", nil)
			expected.Details = "lookup details"
			c := &applicationChartInspector{
				inspector: chartinspection.NewInspector(nil, time.Minute),
				loadUpload: func(context.Context, uuid.UUID, string) ([]byte, *api.Error) {
					return nil, expected
				},
			}
			response, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(inspectionTestContext("1", "alice"), api.AiApplicationChartInspectionRequest{
				ChartRef: api.AiApplicationSpecChartRef{ChartId: "11111111-1111-4111-8111-111111111111"},
			})
			body, ok := response.Body.(api.Error)
			if err != nil || response.Code != status || !ok || int(body.Code) != status || body.Message != "upload lookup failed" || body.Details != "" {
				t.Fatalf("lookup error was remapped: %#v, error=%v", response, err)
			}
		})
	}
}

func TestChartInspectionCancelsUploadLookupWithCaller(t *testing.T) {
	ctx, cancel := context.WithCancel(inspectionTestContext("1", "alice"))
	defer cancel()
	started := make(chan context.Context, 1)
	stopped := make(chan error, 1)
	response := make(chan api.ImplResponse, 1)
	c := &applicationChartInspector{
		inspector: chartinspection.NewInspector(nil, time.Minute),
		loadUpload: func(ctx context.Context, _ uuid.UUID, _ string) ([]byte, *api.Error) {
			started <- ctx
			<-ctx.Done()
			stopped <- ctx.Err()
			return nil, api.NewError(http.StatusRequestTimeout, "upload lookup cancelled", nil)
		},
	}
	go func() {
		out, _ := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(ctx, api.AiApplicationChartInspectionRequest{
			ChartRef: api.AiApplicationSpecChartRef{ChartId: "11111111-1111-4111-8111-111111111111"},
		})
		response <- out
	}()
	select {
	case lookupContext := <-started:
		if tenant := runaictx.GetString(lookupContext, runaictx.CfTenantId); tenant == nil || *tenant != "1" {
			t.Fatal("upload lookup lost authenticated context")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("upload lookup did not start")
	}
	cancel()
	select {
	case err := <-stopped:
		if err != context.Canceled {
			t.Fatalf("lookup cancellation = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("caller cancellation did not reach upload lookup")
	}
	select {
	case out := <-response:
		if out.Code != http.StatusRequestTimeout {
			t.Fatalf("cancelled inspection response = %#v", out)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled upload inspection did not return")
	}
}

func TestChartInspectionRequiresTrustedCallerIdentity(t *testing.T) {
	c := &applicationChartInspector{inspector: chartinspection.NewInspector(nil, time.Minute)}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	out, err := (&AIApplicationsService{chartInspection: c}).InspectApplicationChart(ctx, api.AiApplicationChartInspectionRequest{ChartRef: api.AiApplicationSpecChartRef{ChartURL: "https://example.com/chart.tgz"}})
	if err != nil || out.Code != 401 {
		t.Fatalf("missing caller identity accepted: %#v %v", out, err)
	}
}

func inspectionTestContext(tenant, subject string) context.Context {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenant)
	return context.WithValue(ctx, runaictx.CfSubjectID, subject)
}

type inspectionTestReporter interface {
	Helper()
	Fatal(...any)
}

func inspectionTestArchive(t inspectionTestReporter) []byte {
	questions := "questions:\n- variable: replicas\n  label: Replicas\n  type: int\n"
	return inspectionTestArchiveWithQuestions(t, &questions)
}

func inspectionTestArchiveWithQuestions(t inspectionTestReporter, questions *string) []byte {
	t.Helper()
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	tw := tar.NewWriter(gz)
	files := []struct{ name, data string }{
		{"example/Chart.yaml", "apiVersion: v2\nname: example\nversion: 1.0.0\n"},
		{"example/values.yaml", "replicas: 1\n"},
	}
	if questions != nil {
		files = append(files, struct{ name, data string }{"example/questions.yaml", *questions})
	}
	for _, f := range files {
		if err := tw.WriteHeader(&tar.Header{Name: f.name, Mode: 0600, Size: int64(len(f.data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(f.data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
