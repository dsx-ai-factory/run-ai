/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package k8sobject

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
)

// Shared test fixture names — kept as consts so goconst doesn't flag repeat
// usage across the test file.
const (
	testWorkloadAlpha = "alpha"
)

// --- in-memory test cacher (implements cache.Cacher) -----------------------

type testCacher struct {
	mu       sync.Mutex
	store    map[string]testCacheEntry
	ttl      time.Duration
	getError error // when non-nil, Get returns this error — used to test fail-closed behavior
}

type testCacheEntry struct {
	value     any
	expiresAt time.Time
}

func newTestCacher(ttl time.Duration) *testCacher {
	return &testCacher{
		store: map[string]testCacheEntry{},
		ttl:   ttl,
	}
}

func (c *testCacher) keyOf(storeKey reflect.Type, key string) string {
	return storeKey.String() + ":" + key
}

func (c *testCacher) Get(_ context.Context, storeKey reflect.Type, key string) (any, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.getError != nil {
		return nil, c.getError
	}
	entry, ok := c.store[c.keyOf(storeKey, key)]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, nil
	}
	return entry.value, nil // stored as *T via reflection in Set
}

func (c *testCacher) Set(_ context.Context, storeKey reflect.Type, key string, value any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// cache.Set[T] passes value as T (not *T), but cache.Get[T] does res.(*T).
	// Mirror the real cacher: wrap a copy of value in a *T pointer.
	v := reflect.ValueOf(value)
	ptr := reflect.New(v.Type())
	ptr.Elem().Set(v)
	c.store[c.keyOf(storeKey, key)] = testCacheEntry{
		value:     ptr.Interface(),
		expiresAt: time.Now().Add(c.ttl),
	}
	return nil
}

func (c *testCacher) Del(_ context.Context, storeKey reflect.Type, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, c.keyOf(storeKey, key))
	return nil
}

func (c *testCacher) RegisterStore(_ reflect.Type, _ string, _ ...options.Option) error {
	return nil
}

func (c *testCacher) RegisterSharedStore(_ reflect.Type, _ string, _ ...options.Option) error {
	return nil
}

func (c *testCacher) ClearStore(_ context.Context, storeKey reflect.Type) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	prefix := storeKey.String() + ":"
	for k := range c.store {
		if strings.HasPrefix(k, prefix) {
			delete(c.store, k)
		}
	}
	return nil
}

func (c *testCacher) ClearCache(_ context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = map[string]testCacheEntry{}
	return nil
}

func (c *testCacher) Close() error { return nil }

// --- fake workloads-getter ---------------------------------------------------

type fakeAPIWorkloadsGetter struct {
	mu        sync.Mutex
	calls     []callRecord
	responses []fakeResponse
}

type callRecord struct {
	workloadID string
	ctxToken   any // value of workloadsapi.ContextAccessToken on the ctx, for forwarding test
}

type fakeResponse struct {
	workload *AuthorizedWorkloadInfo
	resp     *http.Response
	err      error
}

func (g *fakeAPIWorkloadsGetter) GetWorkloadById(ctx context.Context, workloadID string) (*AuthorizedWorkloadInfo, *http.Response, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.calls = append(g.calls, callRecord{
		workloadID: workloadID,
		ctxToken:   ctx.Value(workloadsapi.ContextAccessToken),
	})
	if len(g.responses) == 0 {
		return nil, nil, errors.New("unexpected upstream call: no canned response")
	}
	r := g.responses[0]
	g.responses = g.responses[1:]
	return r.workload, r.resp, r.err
}

func (g *fakeAPIWorkloadsGetter) callCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.calls)
}

// --- fixture helpers --------------------------------------------------------

func makeWorkload(name string) *AuthorizedWorkloadInfo {
	return &AuthorizedWorkloadInfo{
		Namespace: "runai-test",
		Name:      name,
		Kind:      "training",
	}
}

func makeHTTPResp(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
}

// --- specs ------------------------------------------------------------------

var _ = Describe("CachedWorkloadAuthorizer", func() {
	var (
		getter     *fakeAPIWorkloadsGetter
		authorizer *CachedWorkloadAuthorizer
	)

	BeforeEach(func() {
		cache.SetMock(newTestCacher(time.Minute))
		getter = &fakeAPIWorkloadsGetter{}
		var err error
		authorizer, err = NewCachedWorkloadAuthorizer(getter, time.Minute)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("on a cache miss with a 200 OK upstream response", func() {
		var expected *AuthorizedWorkloadInfo

		BeforeEach(func() {
			expected = makeWorkload(testWorkloadAlpha)
			getter.responses = []fakeResponse{
				{workload: expected, resp: makeHTTPResp(http.StatusOK, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("returns the projected workload info", func() {
			res, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(res.HTTPStatus).To(Equal(http.StatusOK))
			Expect(res.Workload).NotTo(BeNil())
			Expect(res.Workload.Name).To(Equal(expected.Name))
			Expect(res.Workload.Namespace).To(Equal(expected.Namespace))
			Expect(res.Workload.Kind).To(Equal(expected.Kind))
			Expect(getter.callCount()).To(Equal(1))
		})
	})

	Context("on a cache hit", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{workload: makeWorkload(testWorkloadAlpha), resp: makeHTTPResp(http.StatusOK, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
				// Second response intentionally absent — any second upstream call
				// would hit the "unexpected upstream call" error.
			}
			_, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
		})

		It("serves from cache without calling upstream again", func() {
			res, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Workload).NotTo(BeNil())
			Expect(res.Workload.Name).To(Equal(testWorkloadAlpha))
			Expect(getter.callCount()).To(Equal(1))
		})
	})

	Context("when upstream returns 401 Unauthorized", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{resp: makeHTTPResp(http.StatusUnauthorized, ""), err: errors.New("401 Unauthorized")}, //nolint:bodyclose // test fixture; status-only, body unused
			}
		})

		It("relays the status and the upstream error message without surfacing as an error", func() {
			res, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(res.HTTPStatus).To(Equal(http.StatusUnauthorized))
			Expect(res.Workload).To(BeNil())
			Expect(res.UpstreamMsg).To(Equal("401 Unauthorized"))
		})

		It("caches the response so a second call doesn't hit upstream", func() {
			_, _ = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			_, _ = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(getter.callCount()).To(Equal(1))
		})
	})

	Context("when upstream returns 403 Forbidden", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{resp: makeHTTPResp(http.StatusForbidden, ""), err: errors.New("403 Forbidden")}, //nolint:bodyclose // test fixture; status-only, body unused
			}
		})

		It("relays the status and the upstream error message without surfacing as an error", func() {
			res, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(res.HTTPStatus).To(Equal(http.StatusForbidden))
			Expect(res.Workload).To(BeNil())
			Expect(res.UpstreamMsg).To(Equal("403 Forbidden"))
		})

		It("caches the response so a second call doesn't hit upstream", func() {
			_, _ = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			_, _ = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(getter.callCount()).To(Equal(1))
		})
	})

	Context("when upstream returns 404 Not Found", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{resp: makeHTTPResp(http.StatusNotFound, "not found"), err: errors.New("404 NotFound")}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("relays the status without surfacing as an error", func() {
			res, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(res.HTTPStatus).To(Equal(http.StatusNotFound))
		})

		It("caches the response so a second call doesn't hit upstream", func() {
			_, _ = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			_, _ = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(getter.callCount()).To(Equal(1))
		})
	})

	Context("when upstream returns 500 Internal Server Error", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{resp: makeHTTPResp(http.StatusInternalServerError, "boom"), err: errors.New("500 InternalError")}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
				{resp: makeHTTPResp(http.StatusInternalServerError, "boom"), err: errors.New("500 InternalError")}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("surfaces the error and does not cache, so retries hit upstream", func() {
			_, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).To(HaveOccurred())

			_, err = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).To(HaveOccurred())

			Expect(getter.callCount()).To(Equal(2))
		})
	})

	Context("when upstream returns a network error", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{err: errors.New("connection refused")},
				{err: errors.New("connection refused")},
			}
		})

		It("surfaces the error and does not cache, so retries hit upstream", func() {
			_, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).To(HaveOccurred())

			_, err = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).To(HaveOccurred())

			Expect(getter.callCount()).To(Equal(2))
		})
	})

	Context("after the cache TTL expires", func() {
		const shortTTL = 50 * time.Millisecond

		BeforeEach(func() {
			cache.SetMock(newTestCacher(shortTTL))
			getter.responses = []fakeResponse{
				{workload: makeWorkload(testWorkloadAlpha), resp: makeHTTPResp(http.StatusOK, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
				{workload: makeWorkload(testWorkloadAlpha), resp: makeHTTPResp(http.StatusOK, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("refetches from upstream", func() {
			_, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())

			time.Sleep(shortTTL + 50*time.Millisecond)

			_, err = authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			Expect(getter.callCount()).To(Equal(2))
		})
	})

	Context("when two different subjects access the same workload", func() {
		BeforeEach(func() {
			getter.responses = []fakeResponse{
				{workload: makeWorkload(testWorkloadAlpha), resp: makeHTTPResp(http.StatusOK, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
				{workload: makeWorkload("beta"), resp: makeHTTPResp(http.StatusOK, ""), err: nil},            //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("isolates cache entries per subject — a 200 cached for one user doesn't leak to another", func() {
			resA, err := authorizer.Authorize(context.Background(), "user-A", "wl-1")
			Expect(err).NotTo(HaveOccurred())
			resB, err := authorizer.Authorize(context.Background(), "user-B", "wl-1")
			Expect(err).NotTo(HaveOccurred())

			Expect(resA.Workload.Name).To(Equal(testWorkloadAlpha))
			Expect(resB.Workload.Name).To(Equal("beta"))
			Expect(getter.callCount()).To(Equal(2))
		})
	})

	Context("when upstream returns a non-200 2xx (e.g. 204)", func() {
		BeforeEach(func() {
			// Simulate upstream returning err=nil but a status the spec doesn't
			// define for this endpoint (e.g., 204 No Content). The defensive
			// guard should reject it rather than silently treating as success.
			getter.responses = []fakeResponse{
				{workload: nil, resp: makeHTTPResp(http.StatusNoContent, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("surfaces an error rather than silently treating the response as 200", func() {
			_, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected upstream success state"))
		})
	})

	Context("when cache Get returns an error (cache outage)", func() {
		BeforeEach(func() {
			// Replace the default cacher with one that errors on every Get.
			brokenCacher := newTestCacher(time.Minute)
			brokenCacher.getError = errors.New("cache broken")
			cache.SetMock(brokenCacher)
		})

		It("fails closed instead of silently dumping all traffic onto upstream", func() {
			_, err := authorizer.Authorize(context.Background(), "user-1", "wl-1")
			Expect(err).To(HaveOccurred())
			// Upstream must NOT be called on a cache outage — fail-closed surfaces
			// the cache problem rather than hammering control-plane.
			Expect(getter.callCount()).To(Equal(0))
		})
	})

	Context("when subjects contain the cache key separator character", func() {
		BeforeEach(func() {
			// Two subjects that would collide under naive "%s|%s" formatting:
			// ("a|b", "c") and ("a", "b|c") both produce "a|b|c" without escaping.
			getter.responses = []fakeResponse{
				{workload: makeWorkload(testWorkloadAlpha), resp: makeHTTPResp(http.StatusOK, ""), err: nil}, //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
				{workload: makeWorkload("beta"), resp: makeHTTPResp(http.StatusOK, ""), err: nil},            //nolint:bodyclose // test fixture; body closed by authorizer defer at runtime
			}
		})

		It("does not collide because both components are url.QueryEscaped", func() {
			resA, err := authorizer.Authorize(context.Background(), "a|b", "c")
			Expect(err).NotTo(HaveOccurred())
			resB, err := authorizer.Authorize(context.Background(), "a", "b|c")
			Expect(err).NotTo(HaveOccurred())

			Expect(resA.Workload.Name).To(Equal(testWorkloadAlpha))
			Expect(resB.Workload.Name).To(Equal("beta"))
			Expect(getter.callCount()).To(Equal(2))
		})
	})
})
