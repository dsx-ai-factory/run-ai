package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/meta"

	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	mocks "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/http"
)

var _ = Describe("Router", func() {
	var (
		container           Container
		recorder            *httptest.ResponseRecorder
		request             *http.Request
		logsHandler         *mocks.MockUrlHandler
		logsDownloadHandler *mocks.MockUrlHandler
		pfHandler           *mocks.MockUrlHandler
	)

	BeforeEach(func() {
		router := chi.NewRouter()
		workloadMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulating middleware behavior
				w.Header().Set("X-Middleware", "true")
				next.ServeHTTP(w, r)
			})
		}
		mapHandler := map[string]UrlHandler{}
		authMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulating middleware behavior
				w.Header().Set("X-AuthMiddleware", "true")
				next.ServeHTTP(w, r)
			})
		}
		authzMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulating middleware behavior
				w.Header().Set("X-AuthzMiddleware", "true")
				next.ServeHTTP(w, r)
			})
		}
		projectMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulating middleware behavior
				w.Header().Set("X-ProjectMiddleware", "true")
				next.ServeHTTP(w, r)
			})
		}
		workloadAttributeMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulating middleware behavior
				w.Header().Set("X-WorkloadAttributeMiddleware", "true")
				next.ServeHTTP(w, r)
			})
		}
		metricsMiddleware := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simulating middleware behavior
				w.Header().Set("X-MetricsMiddleware", "true")
				next.ServeHTTP(w, r)
			})
		}
		logsHandler = &mocks.MockUrlHandler{}
		logsDownloadHandler = &mocks.MockUrlHandler{}
		pfHandler = &mocks.MockUrlHandler{}
		mapHandler[PathLogs] = logsHandler
		mapHandler[PathLogsDownload] = logsDownloadHandler
		mapHandler[PathPortforward] = pfHandler
		container = NewContainer(router, workloadMiddleware, projectMiddleware, workloadAttributeMiddleware, authzMiddleware, authMiddleware, metricsMiddleware, mapHandler)
		recorder = httptest.NewRecorder()
	})

	Describe("CreateRouter", func() {
		It("should create a router with the health endpoint", func() {
			request, _ = http.NewRequest(http.MethodGet, "/health", nil)
			router := CreateRouter(container)
			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusOK))
		})

		It("should create a router with the status endpoint", func() {
			request, _ = http.NewRequest(http.MethodGet, "/cluster-api/status", nil)
			router := CreateRouter(container)
			router.ServeHTTP(recorder, request)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			metaObj := &meta.Metadata{}
			err := json.Unmarshal(recorder.Body.Bytes(), metaObj)
			Expect(err).ToNot(HaveOccurred())
			Expect(metaObj.Version).To(Equal("1.0"))
			Expect(metaObj.Features.MessageEnabled).To(Equal(false))
		})

		Context("with framework", func() {
			const (
				workloadFramework = "test-framework"
				project           = "test-project"
				workloadType      = "test-workload-type"
				workloadName      = "test-workload-name"
			)

			DescribeTableSubtree("created router should include endpoint",
				func(actionPath string) {
					It("GET method", func() {
						handler := &mocks.MockUrlHandler{}
						container.MapHandler[actionPath] = handler
						url := "/cluster-api/api/v1/" + project + "/workloads/" + workloadType + "/" + workloadFramework + "/" + workloadName + actionPath

						request, _ = http.NewRequest(http.MethodGet, url, nil)
						router := CreateRouter(container)
						router.ServeHTTP(recorder, request)

						Expect(recorder.Code).To(Equal(http.StatusOK))
						Expect(recorder.Header().Get("X-Middleware")).To(Equal("true"))
						Expect(handler.GetCalled).To(BeTrue())
					})

					It("POST method", func() {
						handler := &mocks.MockUrlHandler{}
						container.MapHandler[actionPath] = handler

						url := "/cluster-api/api/v1/" + project + "/workloads/" + workloadType + "/" + workloadFramework + "/" + workloadName + actionPath

						request, _ = http.NewRequest(http.MethodPost, url, nil)
						router := CreateRouter(container)
						router.ServeHTTP(recorder, request)

						Expect(recorder.Code).To(Equal(http.StatusOK))
						Expect(recorder.Header().Get("X-Middleware")).To(Equal("true"))
						Expect(handler.PostCalled).To(BeTrue())
					})
				},
				Entry("logs", PathLogs),
				Entry("logsDownload", PathLogsDownload),
				Entry("pf", PathPortforward),
				Entry("exec", PathExec),
				Entry("attach", PathAttach),
			)
		})
	})
})
