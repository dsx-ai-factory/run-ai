/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type inspectionCountingBody struct {
	reader    io.Reader
	readBytes int
}

func (r *inspectionCountingBody) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.readBytes += n
	return n, err
}
func (r *inspectionCountingBody) Close() error { return nil }

func TestChartInspectionBodyLimitPreservesSmallBody(t *testing.T) {
	const body = `{"chartRef":{"chartURL":"https://example.com/chart.tgz"}}`
	called := false
	handler := chartInspectionBodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		got, err := io.ReadAll(r.Body)
		if err != nil || string(got) != body {
			t.Fatalf("body changed: %q, %v", got, err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", strings.NewReader(body)))
	if !called || response.Code != http.StatusNoContent {
		t.Fatalf("small request blocked: %d", response.Code)
	}
}

func TestChartInspectionBodyLimitRejectsDeclaredOversizeWithoutReading(t *testing.T) {
	body := &inspectionCountingBody{reader: strings.NewReader(strings.Repeat(" ", (64<<10)+1))}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", nil)
	request.ContentLength = (64 << 10) + 1
	request.Body = body
	response := httptest.NewRecorder()
	called := false
	chartInspectionBodyLimit(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge || called || body.readBytes != 0 {
		t.Fatalf("declared oversize reached handler/read body: status=%d called=%v bytes=%d", response.Code, called, body.readBytes)
	}
}

func TestChartInspectionBodyLimitBoundsChunkedReads(t *testing.T) {
	body := &inspectionCountingBody{reader: strings.NewReader(strings.Repeat(" ", 1<<20))}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", nil)
	request.ContentLength = -1
	request.TransferEncoding = []string{"chunked"}
	request.Body = body
	response := httptest.NewRecorder()
	called := false
	chartInspectionBodyLimit(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge || called || body.readBytes > (64<<10)+1 {
		t.Fatalf("chunked oversize was not bounded: status=%d called=%v bytes=%d", response.Code, called, body.readBytes)
	}
}

func TestChartInspectionBodyLimitLeavesOtherRoutesAlone(t *testing.T) {
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/ai-applications/charts"},
		{http.MethodPost, "/api/v1/ai-applications"},
		{http.MethodGet, "/api/v1/ai-applications/charts/inspect"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			body := &inspectionCountingBody{reader: strings.NewReader(strings.Repeat(" ", 1<<20))}
			request := httptest.NewRequest(tc.method, tc.path, nil)
			request.ContentLength = 1 << 20
			request.Body = body
			response := httptest.NewRecorder()
			called := false
			chartInspectionBodyLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Body != body {
					t.Fatal("non-inspection body replaced")
				}
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(response, request)
			if !called || response.Code != http.StatusNoContent || body.readBytes != 0 {
				t.Fatalf("non-inspection request affected: status=%d called=%v bytes=%d", response.Code, called, body.readBytes)
			}
		})
	}
}

func TestChartInspectionBodyLimitRejectsConcatenatedJSONWithoutDispatch(t *testing.T) {
	body := `{"chartRef":{"chartURL":"https://example.com/chart.tgz"}}{"unexpected":true}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/ai-applications/charts/inspect", strings.NewReader(body))
	response := httptest.NewRecorder()
	called := false
	chartInspectionBodyLimit(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })).ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || called {
		t.Fatalf("concatenated documents reached handler: status=%d called=%v", response.Code, called)
	}
}
