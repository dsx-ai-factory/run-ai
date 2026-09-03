/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package download_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"

	mocks "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/logs"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/logs/download"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/golang/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// MockReadCloser is a simple mock for testing
type MockReadCloser struct {
	readData  []byte
	readErr   error
	closeErr  error
	readCount int
}

func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	m.readCount++
	if m.readErr != nil {
		return 0, m.readErr
	}

	if len(m.readData) == 0 {
		return 0, io.EOF
	}

	n = copy(p, m.readData)
	m.readData = m.readData[n:]

	return n, nil
}

func (m *MockReadCloser) Close() error {
	return m.closeErr
}

var _ = Describe("Download Handler", func() {
	var (
		ctrl         *gomock.Controller
		handler      *download.Handler
		req          *http.Request
		recorder     *httptest.ResponseRecorder
		testWorkload *workload.Workload
		mockService  *mocks.MockPodLogsService
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockService = mocks.NewMockPodLogsService(ctrl)
		handler = download.NewHandler(mockService).(*download.Handler)
		recorder = httptest.NewRecorder()

		testWorkload = &workload.Workload{
			Name:      "test-workload",
			Namespace: "test-namespace",
			Project:   "test-project",
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Handle", func() {
		Context("when the request is valid and successful", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/download?pod=0-0&container=cont", nil)

				// Set up context with workload
				ctx := workload.SetIntoContext(req.Context(), testWorkload)
				req = req.WithContext(ctx)

				// Set up mock service to return successful reader
				mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(&MockReadCloser{
					readData: []byte("test log data"),
				}, nil)
			})

			It("should successfully download logs", func() {
				handler.Handle(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Header().Get("Content-Type")).To(Equal("application/octet-stream"))
				Expect(recorder.Header().Get("Content-Disposition")).To(ContainSubstring("attachment; filename=\"test-workload_test-project_0-0_cont_"))
				Expect(recorder.Header().Get("Content-Disposition")).To(HaveSuffix(".log\""))
				Expect(recorder.Header().Get("Transfer-Encoding")).To(Equal("chunked"))
				Expect(recorder.Header().Get("Cache-Control")).To(Equal("no-cache"))
				Expect(recorder.Header().Get("Connection")).To(Equal("keep-alive"))
			})

			It("should create correct filename in content disposition", func() {
				handler.Handle(recorder, req)

				contentDisposition := recorder.Header().Get("Content-Disposition")
				Expect(contentDisposition).To(ContainSubstring("test-workload"))
				Expect(contentDisposition).To(ContainSubstring("test-project"))
				Expect(contentDisposition).To(ContainSubstring(".log"))
			})

			It("should stream log data correctly", func() {
				handler.Handle(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal("test log data"))
			})
		})

		Context("when getting workload and options fails", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/download", nil)
				// Don't set context values to simulate failure
			})

			It("should return internal server error", func() {
				handler.Handle(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when LogsReader returns an error", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/download", nil)
				ctx := workload.SetIntoContext(req.Context(), testWorkload)
				req = req.WithContext(ctx)

				mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(nil, errors.New("logs reader error"))
			})

			It("should return service unavailable error", func() {
				handler.Handle(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
			})
		})

		Context("when streaming fails", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/download", nil)
				ctx := workload.SetIntoContext(req.Context(), testWorkload)
				req = req.WithContext(ctx)
			})

			It("should return service unavailable error when read fails", func() {
				mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(nil, errors.New("read error"))

				handler.Handle(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusServiceUnavailable))
			})

			It("should return service unavailable error when write fails", func() {
				// Create a response recorder that fails on write
				failingRecorder := &failingResponseRecorder{
					ResponseRecorder: *httptest.NewRecorder(),
					shouldFail:       true,
				}

				mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(&MockReadCloser{
					readData: []byte("test data"),
				}, nil)

				handler.Handle(failingRecorder, req)

				Expect(failingRecorder.Code).To(Equal(http.StatusServiceUnavailable))
			})
		})

		Context("when response writer doesn't support flushing", func() {
			BeforeEach(func() {
				req = httptest.NewRequest(http.MethodGet, "/download", nil)
				ctx := workload.SetIntoContext(req.Context(), testWorkload)
				req = req.WithContext(ctx)

				mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(&MockReadCloser{
					readData: []byte("test data"),
				}, nil)
			})

			It("should return service unavailable error", func() {
				// Create a response recorder that doesn't support flushing
				nonFlushingRecorder := &nonFlushingResponseRecorder{}

				handler.Handle(nonFlushingRecorder, req)

				Expect(nonFlushingRecorder.Code()).To(Equal(http.StatusServiceUnavailable))
			})
		})

		Context("when follow option is true", func() {
			BeforeEach(func() {
				// Create request with follow=true query parameter
				req = httptest.NewRequest(http.MethodGet, "/download?follow=true", nil)
				ctx := workload.SetIntoContext(req.Context(), testWorkload)
				req = req.WithContext(ctx)
			})

			It("should return internal server error because logs download does not support following", func() {
				handler.Handle(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("Logs download does not support option attribute"))
			})
		})
	})

	Describe("Get", func() {
		It("should call Handle method", func() {
			req = httptest.NewRequest(http.MethodGet, "/download", nil)
			ctx := workload.SetIntoContext(req.Context(), testWorkload)
			req = req.WithContext(ctx)

			mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(&MockReadCloser{
				readData: []byte("test data"),
			}, nil)

			handler.Get(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("Post", func() {
		It("should call Handle method", func() {
			req = httptest.NewRequest(http.MethodPost, "/download", nil)
			ctx := workload.SetIntoContext(req.Context(), testWorkload)
			req = req.WithContext(ctx)

			mockService.EXPECT().LogsReader(gomock.Any(), testWorkload, gomock.Any()).Return(&MockReadCloser{
				readData: []byte("test data"),
			}, nil)

			handler.Post(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("NewHandler", func() {
		It("should create a new handler with the provided service", func() {
			newHandler := download.NewHandler(mockService)

			Expect(newHandler).NotTo(BeNil())
			Expect(newHandler).To(BeAssignableToTypeOf(&download.Handler{}))
		})
	})
})

// Helper struct for testing write failures
type failingResponseRecorder struct {
	httptest.ResponseRecorder
	shouldFail bool
}

func (f *failingResponseRecorder) Write(data []byte) (int, error) {
	if f.shouldFail {
		return 0, errors.New("write error")
	}

	return f.ResponseRecorder.Write(data)
}

func (f *failingResponseRecorder) Flush() {
	// httptest.ResponseRecorder doesn't implement http.Flusher, so we don't need to do anything
}

// Helper struct for testing non-flushing response writers
type nonFlushingResponseRecorder struct {
	headers http.Header
	body    []byte
	code    int
}

func (n *nonFlushingResponseRecorder) Header() http.Header {
	if n.headers == nil {
		n.headers = make(http.Header)
	}

	return n.headers
}

func (n *nonFlushingResponseRecorder) Write(data []byte) (int, error) {
	n.body = append(n.body, data...)
	return len(data), nil
}

func (n *nonFlushingResponseRecorder) WriteHeader(code int) {
	n.code = code
}

func (n *nonFlushingResponseRecorder) Code() int {
	return n.code
}

// Note: This struct intentionally does NOT implement http.Flusher
