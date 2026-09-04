/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Middleware Suite")
}

var _ = Describe("RequestContextMiddleware", func() {
	var (
		router          chi.Router
		testHandler     http.HandlerFunc
		capturedContext context.Context
	)

	BeforeEach(func() {
		// Create a test handler that captures the context
		testHandler = func(w http.ResponseWriter, r *http.Request) {
			capturedContext = r.Context()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}

		// Create router with middleware
		router = chi.NewRouter()
		router.Use(RequestContextMiddleware)
		router.Post("/test-path", testHandler)
		router.Get("/test-path/{id}", testHandler)
		router.Put("/api/v2/workloads/nim-services", testHandler)
	})

	Context("request path capture", func() {
		It("should capture simple request path", func() {
			req := httptest.NewRequest("POST", "/test-path", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(GetRequestPath(capturedContext)).To(Equal("/test-path"))
		})

		It("should capture request path with parameters", func() {
			req := httptest.NewRequest("GET", "/test-path/123", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(GetRequestPath(capturedContext)).To(Equal("/test-path/123"))
		})

		It("should capture nim-services API path", func() {
			req := httptest.NewRequest("PUT", "/api/v2/workloads/nim-services", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(GetRequestPath(capturedContext)).To(Equal("/api/v2/workloads/nim-services"))
		})

		It("should handle empty path", func() {
			req := httptest.NewRequest("POST", "/", nil)
			w := httptest.NewRecorder()

			// Add a handler for root path
			router.Post("/", testHandler)
			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(GetRequestPath(capturedContext)).To(Equal("/"))
		})
	})

	Context("request body capture", func() {
		It("should capture JSON request body", func() {
			jsonBody := `{"metadata":{"name":"test-workload","projectId":"project-123"},"spec":{"image":"nginx:latest"}}`
			req := httptest.NewRequest("POST", "/test-path", strings.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			capturedBody := GetRequestBody(capturedContext)
			Expect(capturedBody).NotTo(BeNil())
			Expect(string(capturedBody)).To(Equal(jsonBody))
		})

		It("should capture large request body", func() {
			largeBody := strings.Repeat("a", 10000) // 10KB body
			req := httptest.NewRequest("POST", "/test-path", strings.NewReader(largeBody))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			capturedBody := GetRequestBody(capturedContext)
			Expect(capturedBody).NotTo(BeNil())
			Expect(string(capturedBody)).To(Equal(largeBody))
		})

		It("should handle empty request body", func() {
			req := httptest.NewRequest("POST", "/test-path", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			capturedBody := GetRequestBody(capturedContext)
			Expect(capturedBody).To(BeNil())
		})

		It("should handle request with empty body reader", func() {
			req := httptest.NewRequest("POST", "/test-path", strings.NewReader(""))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			capturedBody := GetRequestBody(capturedContext)
			Expect(capturedBody).To(BeNil()) // Empty body should result in nil
		})

		It("should preserve request body for downstream handlers", func() {
			originalBody := `{"test":"data"}`
			var downstreamBody []byte

			// Create a handler that reads the body
			bodyReadingHandler := func(w http.ResponseWriter, r *http.Request) {
				capturedContext = r.Context()
				if r.Body != nil {
					body, err := io.ReadAll(r.Body)
					Expect(err).To(BeNil())
					downstreamBody = body
				}
				w.WriteHeader(http.StatusOK)
			}

			router.Post("/body-test", bodyReadingHandler)

			req := httptest.NewRequest("POST", "/body-test", strings.NewReader(originalBody))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))

			// Both middleware and downstream handler should have access to the body
			capturedBody := GetRequestBody(capturedContext)
			Expect(string(capturedBody)).To(Equal(originalBody))
			Expect(string(downstreamBody)).To(Equal(originalBody))
		})
	})

	Context("context helper functions", func() {
		It("should return empty string for missing request path", func() {
			ctx := context.Background()
			path := GetRequestPath(ctx)
			Expect(path).To(Equal(""))
		})

		It("should return nil for missing request body", func() {
			ctx := context.Background()
			body := GetRequestBody(ctx)
			Expect(body).To(BeNil())
		})

		It("should handle wrong type in context for path", func() {
			ctx := context.WithValue(context.Background(), RequestPathKey, 123) // Wrong type
			path := GetRequestPath(ctx)
			Expect(path).To(Equal(""))
		})

		It("should handle wrong type in context for body", func() {
			ctx := context.WithValue(context.Background(), RequestBodyKey, "not-bytes") // Wrong type
			body := GetRequestBody(ctx)
			Expect(body).To(BeNil())
		})
	})

	Context("middleware integration", func() {
		It("should work with multiple middlewares", func() {
			// Add another middleware before our middleware
			authMiddleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Add some auth context
					ctx := context.WithValue(r.Context(), "user", "test-user")
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			}

			testRouter := chi.NewRouter()
			testRouter.Use(authMiddleware)
			testRouter.Use(RequestContextMiddleware)
			testRouter.Post("/multi-middleware", testHandler)

			jsonBody := `{"test":"multi-middleware"}`
			req := httptest.NewRequest("POST", "/multi-middleware", strings.NewReader(jsonBody))
			w := httptest.NewRecorder()

			testRouter.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))

			// Both middlewares should work
			Expect(capturedContext.Value("user")).To(Equal("test-user"))
			Expect(GetRequestPath(capturedContext)).To(Equal("/multi-middleware"))
			Expect(string(GetRequestBody(capturedContext))).To(Equal(jsonBody))
		})

		It("should handle middleware chain with error", func() {
			errorMiddleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Simulate an error condition but continue
					if r.Header.Get("X-Trigger-Error") == "true" {
						w.Header().Set("X-Error", "triggered")
					}
					next.ServeHTTP(w, r)
				})
			}

			testRouter := chi.NewRouter()
			testRouter.Use(RequestContextMiddleware)
			testRouter.Use(errorMiddleware)
			testRouter.Post("/error-test", testHandler)

			req := httptest.NewRequest("POST", "/error-test", strings.NewReader(`{"test":"error"}`))
			req.Header.Set("X-Trigger-Error", "true")
			w := httptest.NewRecorder()

			testRouter.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(w.Header().Get("X-Error")).To(Equal("triggered"))

			// Our middleware should still work
			Expect(GetRequestPath(capturedContext)).To(Equal("/error-test"))
			Expect(string(GetRequestBody(capturedContext))).To(Equal(`{"test":"error"}`))
		})
	})

	Context("edge cases", func() {
		It("should handle request with nil body", func() {
			req := &http.Request{
				Method: "POST",
				URL:    &url.URL{Path: "/test-path"},
				Body:   nil,
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(GetRequestPath(capturedContext)).To(Equal("/test-path"))
			Expect(GetRequestBody(capturedContext)).To(BeNil())
		})

		It("should handle body read error gracefully", func() {
			// Create a reader that will error
			errorReader := &errorReader{}
			req := httptest.NewRequest("POST", "/test-path", errorReader)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			Expect(w.Code).To(Equal(http.StatusOK))
			Expect(GetRequestPath(capturedContext)).To(Equal("/test-path"))
			// Should not panic and should not store body on error
			Expect(GetRequestBody(capturedContext)).To(BeNil())
		})

		It("should handle concurrent requests", func() {
			done := make(chan bool, 2)
			var contexts []context.Context

			handler := func(w http.ResponseWriter, r *http.Request) {
				contexts = append(contexts, r.Context())
				w.WriteHeader(http.StatusOK)
				done <- true
			}

			testRouter := chi.NewRouter()
			testRouter.Use(RequestContextMiddleware)
			testRouter.Post("/concurrent/{id}", handler)

			// Start two concurrent requests
			go func() {
				req := httptest.NewRequest("POST", "/concurrent/1", strings.NewReader(`{"id":1}`))
				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)
			}()

			go func() {
				req := httptest.NewRequest("POST", "/concurrent/2", strings.NewReader(`{"id":2}`))
				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)
			}()

			// Wait for both to complete
			<-done
			<-done

			Expect(len(contexts)).To(Equal(2))

			// Each context should have its own data
			paths := []string{
				GetRequestPath(contexts[0]),
				GetRequestPath(contexts[1]),
			}

			Expect(paths).To(ContainElement("/concurrent/1"))
			Expect(paths).To(ContainElement("/concurrent/2"))
		})
	})
})

// errorReader is a helper type that always returns an error when read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}
