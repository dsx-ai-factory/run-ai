/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// --- fake WorkloadAuthorizer (records args, returns canned response) --------

type fakeAuthorizer struct {
	calls       []authorizeCall
	authorizeFn func(ctx context.Context, subject, workloadID string) (*WorkloadAuthorizationResponse, error)
}

type authorizeCall struct {
	subject    string
	workloadID string
	// ctxToken captures the value of workloadsapi.ContextAccessToken on the
	// ctx passed to Authorize, so tests can assert JWT forwarding.
	ctxToken any
}

func (a *fakeAuthorizer) Authorize(ctx context.Context, subject, workloadID string) (*WorkloadAuthorizationResponse, error) {
	a.calls = append(a.calls, authorizeCall{
		subject:    subject,
		workloadID: workloadID,
		ctxToken:   ctx.Value(workloadsapi.ContextAccessToken),
	})
	return a.authorizeFn(ctx, subject, workloadID)
}

// --- fixture helpers --------------------------------------------------------

// requestWithChiParam returns an httptest request with the chi URLParam set
// and (optionally) auth info installed via auth.ComposeAuthContext.
func requestWithChiParam(workloadID string, authInfo *auth.Info) *http.Request {
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	if authInfo != nil {
		req = req.WithContext(auth.ComposeAuthContext(req.Context(), authInfo))
	}

	rctx := chi.NewRouteContext()
	if workloadID != "" {
		rctx.URLParams.Add(WorkloadIDParam, workloadID)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	return req
}

// nextHandlerThatCapturesCtx returns an http.Handler that records the request
// context it received (so tests can assert WithWorkloadInfo populated it) and
// writes a 200 OK.
func nextHandlerThatCapturesCtx(captured *context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = r.Context()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("inner handler ran"))
	})
}

// authInfoFor returns an *auth.Info populated with the given subject + raw token.
func authInfoFor(subject, rawToken string) *auth.Info {
	return &auth.Info{Subject: subject, RawToken: rawToken}
}

// --- specs ------------------------------------------------------------------

var _ = Describe("WorkloadAuthorizerMiddleware", func() {
	var (
		authorizer  *fakeAuthorizer
		rw          *httptest.ResponseRecorder
		capturedCtx context.Context
	)

	BeforeEach(func() {
		authorizer = &fakeAuthorizer{}
		capturedCtx = nil
		rw = httptest.NewRecorder()
	})

	Context("on the happy path (well-formed request, 200 from authorizer)", func() {
		var workloadInfo *AuthorizedWorkloadInfo

		BeforeEach(func() {
			workloadInfo = &AuthorizedWorkloadInfo{Namespace: "runai-test", Name: testWorkloadAlpha, Kind: "training"}
			authorizer.authorizeFn = func(ctx context.Context, subject, workloadID string) (*WorkloadAuthorizationResponse, error) {
				return &WorkloadAuthorizationResponse{Workload: workloadInfo, HTTPStatus: http.StatusOK}, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", authInfoFor("user-1", "jwt-abcdef"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 200", func() {
			Expect(rw.Code).To(Equal(http.StatusOK))
		})

		It("calls the next handler", func() {
			Expect(capturedCtx).NotTo(BeNil())
		})

		It("places the AuthorizedWorkloadInfo on the downstream context", func() {
			got, ok := WorkloadInfoFromContext(capturedCtx)
			Expect(ok).To(BeTrue())
			Expect(got).To(Equal(workloadInfo))
		})

		It("propagates the JWT to the downstream context for workloads-client calls", func() {
			Expect(capturedCtx.Value(workloadsapi.ContextAccessToken)).To(Equal("jwt-abcdef"))
		})

		It("forwards the JWT to the authorizer call", func() {
			Expect(authorizer.calls).To(HaveLen(1))
			Expect(authorizer.calls[0].ctxToken).To(Equal("jwt-abcdef"))
		})

		It("calls the authorizer with the correct subject and workloadID", func() {
			Expect(authorizer.calls).To(HaveLen(1))
			Expect(authorizer.calls[0].subject).To(Equal("user-1"))
			Expect(authorizer.calls[0].workloadID).To(Equal("11111111-1111-1111-1111-111111111111"))
		})
	})

	Context("when {workloadId} is missing from the URL", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				Fail("authorizer should not be called")
				return nil, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("", authInfoFor("user-1", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 400", func() {
			Expect(rw.Code).To(Equal(http.StatusBadRequest))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when {workloadId} is not a valid UUID", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				Fail("authorizer should not be called for a malformed workloadId")
				return nil, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("not-a-uuid", authInfoFor("user-1", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 400", func() {
			Expect(rw.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns a body that names the offending field", func() {
			Expect(rw.Body.String()).To(ContainSubstring("workloadId"))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when the JWT is missing in auth info", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				Fail("authorizer should not be called without JWT")
				return nil, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", authInfoFor("user-1", ""))
			handler.ServeHTTP(rw, req)
		})

		It("returns 401", func() {
			Expect(rw.Code).To(Equal(http.StatusUnauthorized))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when there is no auth context at all", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				Fail("authorizer should not be called without auth context")
				return nil, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", nil)
			handler.ServeHTTP(rw, req)
		})

		It("returns 401", func() {
			Expect(rw.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("when the subject is empty (token present, subject missing)", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				Fail("authorizer should not be called without subject")
				return nil, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", authInfoFor("", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 401", func() {
			Expect(rw.Code).To(Equal(http.StatusUnauthorized))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when the authorizer returns 403 with a body", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				return &WorkloadAuthorizationResponse{
					HTTPStatus:  http.StatusForbidden,
					UpstreamMsg: "user has no read access",
				}, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", authInfoFor("user-1", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 403", func() {
			Expect(rw.Code).To(Equal(http.StatusForbidden))
		})

		It("relays the upstream body", func() {
			Expect(rw.Body.String()).To(ContainSubstring("user has no read access"))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when the authorizer returns 404", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				return &WorkloadAuthorizationResponse{HTTPStatus: http.StatusNotFound, UpstreamMsg: "workload not found"}, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("22222222-2222-2222-2222-222222222222", authInfoFor("user-1", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 404", func() {
			Expect(rw.Code).To(Equal(http.StatusNotFound))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when the authorizer returns an internal error", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				return nil, errors.New("upstream connection refused")
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(nextHandlerThatCapturesCtx(&capturedCtx))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", authInfoFor("user-1", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns 500", func() {
			Expect(rw.Code).To(Equal(http.StatusInternalServerError))
		})

		It("does not leak the upstream error detail to the client", func() {
			Expect(rw.Body.String()).NotTo(ContainSubstring("connection refused"))
		})

		It("writes a generic error message", func() {
			Expect(rw.Body.String()).To(ContainSubstring("internal server error"))
		})

		It("does not call the next handler", func() {
			Expect(capturedCtx).To(BeNil())
		})
	})

	Context("when the authorizer returns non-OK with an empty UpstreamMsg", func() {
		BeforeEach(func() {
			authorizer.authorizeFn = func(context.Context, string, string) (*WorkloadAuthorizationResponse, error) {
				return &WorkloadAuthorizationResponse{HTTPStatus: http.StatusForbidden, UpstreamMsg: ""}, nil
			}
			handler := WorkloadAuthorizerMiddleware(authorizer)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
			req := requestWithChiParam("11111111-1111-1111-1111-111111111111", authInfoFor("user-1", "jwt-abc"))
			handler.ServeHTTP(rw, req)
		})

		It("returns the relayed status", func() {
			Expect(rw.Code).To(Equal(http.StatusForbidden))
		})

		It("wraps the empty upstream body with the operational context prefix", func() {
			Expect(rw.Body.String()).To(ContainSubstring("resolving workload: ''"))
		})
	})
})
