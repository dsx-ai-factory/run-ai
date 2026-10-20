/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package chart_inspection

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	cachetesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

func cacheChart() []byte {
	return cacheChartMetadata(fixtureChartName, "1")
}

func mockInspectionCache(registerErr error) *cachetesting.MockCacher {
	GinkgoHelper()
	cachetesting.DisableCache()
	shared := cachetesting.SetMock(GinkgoT())
	shared.On("Close").Return(nil)
	shared.On("RegisterStore", cachetesting.KeyFromType[chartInspectionMetadata](), mock.Anything,
		mock.Anything, mock.Anything, mock.Anything).Return(registerErr)
	return shared
}

func ngcCacheSelection() Selection {
	return Selection{Repo: "https://helm.ngc.nvidia.com/nvidia", ChartName: fixtureChartName, ChartVersion: "1"}
}

func cacheChartMetadata(name, version string) []byte {
	GinkgoHelper()

	var output bytes.Buffer
	gz := gzip.NewWriter(&output)
	tw := tar.NewWriter(gz)
	for _, file := range []struct{ name, body string }{
		{"chart/Chart.yaml", "apiVersion: v2\nname: " + name + "\nversion: '" + version + "'\n"},
		{"chart/questions.yaml", "questions:\n  - variable: replicas\n    label: Replicas\n    type: int\n    default: 1\n"},
	} {
		{
			err := tw.WriteHeader(&tar.Header{Name: file.name, Mode: 0600, Size: int64(len(file.body))})
			Expect(err).NotTo(HaveOccurred())
		}
		{
			_, err := tw.Write([]byte(file.body))
			Expect(err).NotTo(HaveOccurred())
		}
	}
	{
		err := tw.Close()
		Expect(err).NotTo(HaveOccurred())
	}
	{
		err := gz.Close()
		Expect(err).NotTo(HaveOccurred())
	}

	return output.Bytes()
}

var _ = Describe("Inspection cache and admission", func() {
	BeforeEach(func() {
		cachetesting.EnableCache()
	})
	AfterEach(func() {
		cachetesting.DisableCache()
	})

	It("honors the CP cache disable setting without failing inspection", func() {
		cachetesting.DisableCache()
		archive := cacheChart()
		calls := 0
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			calls++
			return archive, nil
		}, 0)
		for n := 0; n < 2; n++ {
			result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChartName).To(Equal(fixtureChartName))
		}
		Expect(calls).To(Equal(2))
	})

	It("continues inspection when cache registration, reads, writes, or decoding fail", func() {
		archive := cacheChart()
		unavailable := errors.New("cache unavailable")
		for _, failure := range []string{"registration", "read", "write", "decode"} {
			By(failure)
			var registerErr error
			if failure == "registration" {
				registerErr = unavailable
			}
			shared := mockInspectionCache(registerErr)
			if registerErr == nil {
				var entry any
				var readErr error
				if failure == "read" {
					readErr = unavailable
				}
				if failure == "decode" {
					entry = &chartInspectionMetadata{Data: json.RawMessage(`"invalid metadata"`)}
				}
				shared.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(entry, readErr)
				shared.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(unavailable)
			}
			calls := 0
			inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
				calls++
				return archive, nil
			}, 0)
			for n := 0; n < 2; n++ {
				result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ChartName).To(Equal(fixtureChartName))
			}
			Expect(calls).To(Equal(2))
		}
	})

	It("rechecks a miss before downloading without refreshing a cached entry's TTL", func() {
		parsed, err := ParseArchive(cacheChart())
		Expect(err).NotTo(HaveOccurred())
		encoded, marshalErr := json.Marshal(parsed)
		Expect(marshalErr).NotTo(HaveOccurred())
		shared := mockInspectionCache(nil)
		shared.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()
		shared.On("Get", mock.Anything, mock.Anything, mock.Anything).
			Return(&chartInspectionMetadata{Data: encoded}, nil).Once()
		called := false
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			called = true
			return nil, inspectionError(errors.New("unexpected download"))
		}, 0)
		result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(result.ChartName).To(Equal(fixtureChartName))
		Expect(called).To(BeFalse())
		shared.AssertNotCalled(GinkgoT(), "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	It("keeps unrelated inspection and cancellation responsive during a blocked cache write", func() {
		archive := cacheChart()
		shared := mockInspectionCache(nil)
		shared.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
		started := make(chan struct{})
		shared.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				close(started)
				<-args.Get(0).(context.Context).Done()
			}).Return(context.Canceled).Once()
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			return archive, nil
		}, 0)
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		done := make(chan *api.Error, 1)
		go func() {
			_, err := inspector.Inspect(ctx, ngcCacheSelection(), "tenant", "")
			done <- err
		}()
		Eventually(started, time.Second).Should(BeClosed())
		other := make(chan *api.Error, 1)
		go func() {
			_, err := inspector.Inspect(context.Background(), Selection{ChartID: fixtureUploadID}, "tenant", "")
			other <- err
		}()
		Eventually(other, 500*time.Millisecond).Should(Receive(BeNil()))
		cancel()
		var cancelled *api.Error
		Eventually(done, 500*time.Millisecond).Should(Receive(&cancelled))
		Expect(cancelled.Code).To(BeNumerically("==", http.StatusRequestTimeout))
		Eventually(func() int {
			inspector.mu.Lock()
			defer inspector.mu.Unlock()
			return inspector.active
		}, time.Second).Should(BeZero())
	})

	It("returns API errors while preserving status and hiding private callback diagnostics", func() {
		public := api.NewError(http.StatusForbidden, "chart access denied", nil)
		for _, tc := range []struct {
			name    string
			err     error
			code    int
			message string
		}{
			{"wrapped API error", fmt.Errorf("private callback: %w", public), http.StatusForbidden, public.Message},
			{"cancelled fetch", context.Canceled, http.StatusRequestTimeout, "chart inspection request cancelled or expired"},
			{"expired fetch", context.DeadlineExceeded, http.StatusRequestTimeout, "chart inspection request cancelled or expired"},
			{"unexpected failure", errors.New("private callback diagnostic"), http.StatusInternalServerError, "chart inspection failed unexpectedly"},
		} {
			By(tc.name)
			inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
				return nil, inspectionError(tc.err)
			}, 0)
			result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(result).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(BeNumerically("==", tc.code))
			Expect(err.Message).To(Equal(tc.message))
			Expect(err.Details).To(BeEmpty())
		}
	})

	It("returns a timeout API error before fetching an already cancelled request", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		called := false
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			called = true
			return nil, nil
		}, 0)
		result, err := inspector.Inspect(ctx, ngcCacheSelection(), "tenant", "")
		Expect(result).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(err.Code).To(BeNumerically("==", http.StatusRequestTimeout))
		Expect(called).To(BeFalse())
	})

	It("Inspector Only Reuses NGC Repository Results", func() {
		first := cacheChart()
		updated := chartArchive(
			archiveEntry{name: "app/Chart.yaml", body: "name: app\nversion: '1'\n"},
			archiveEntry{name: "app/questions.yaml", body: "questions:\n- variable: replicas\n  type: int\n  label: Updated\n"},
		)
		for _, tc := range []struct {
			name      string
			selection Selection
			cached    bool
		}{
			{"NGC repository", ngcCacheSelection(), true},
			{"custom URL", Selection{ChartURL: "https://example.com/app.tgz"}, false},
			{"custom NGC URL", Selection{ChartURL: "https://helm.ngc.nvidia.com/app.tgz"}, false},
			{"upload", Selection{ChartID: fixtureUploadID}, false},
			{"other repository", Selection{Repo: "https://example.com/repo", ChartName: fixtureChartName, ChartVersion: "1"}, false},
			{"NGC lookalike", Selection{Repo: "https://helm.ngc.nvidia.com.example.org/repo", ChartName: fixtureChartName, ChartVersion: "1"}, false},
			{"HTTP NGC repository", Selection{Repo: "http://helm.ngc.nvidia.com/nvidia", ChartName: fixtureChartName, ChartVersion: "1"}, false},
		} {
			{
				By(tc.name)
				calls := 0
				inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
					calls++
					if calls == 1 {
						return first, nil
					}
					return updated, nil
				}, 0)
				before, err := inspector.Inspect(context.Background(), tc.selection, "tenant:caller", "")
				Expect(err).NotTo(HaveOccurred())

				after, err := inspector.Inspect(context.Background(), tc.selection, "tenant:caller", "")
				Expect(err).NotTo(HaveOccurred())

				if tc.cached {
					Expect(calls != 1 || after.ArchiveSHA256 != before.ArchiveSHA256).To(BeFalse(),
						"NGC repository was not cached",
					)
				} else {
					Expect(calls != 2 || after.ArchiveSHA256 == before.ArchiveSHA256).To(BeFalse(),
						"fresh source returned cached metadata",
					)
					entry, cacheErr := cache.Get[chartInspectionMetadata](context.Background(), inspector.key(tc.selection, "tenant:caller", ""))
					Expect(cacheErr).NotTo(HaveOccurred())
					Expect(entry).To(BeNil(), "fresh source populated cache")
				}
			}
		}
	})
	It("registers a ten-minute metadata store with CP cache size limits", func() {
		cachetesting.DisableCache()
		shared := cachetesting.SetMock(GinkgoT())
		shared.On("Close").Return(nil)
		shared.On("RegisterStore", cachetesting.KeyFromType[chartInspectionMetadata](), mock.Anything,
			mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			opts := options.NewStoreOptions(args.Get(2).(options.Option), args.Get(3).(options.Option), args.Get(4).(options.Option))
			Expect(opts.Ttl).To(Equal(10 * time.Minute))
			Expect(opts.MaxStoreSizeMB).To(Equal(20))
			Expect(opts.MaxRecordSizeKB).To(Equal(4096))
		}).Return(nil).Once()
		NewInspector(nil, 0)
	})
	It("expires metadata through the shared CP cache", func() {
		archive := cacheChart()
		calls := 0
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			calls++
			return archive, nil
		}, 100*time.Millisecond)
		for n := 0; n < 2; n++ {
			_, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls).To(Equal(1))
		Eventually(func() int {
			_, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())
			return calls
		}, time.Second, 20*time.Millisecond).Should(Equal(2))
	})
	It("Inspector Fresh Requests Do Not Share An Acquisition", func() {
		archive := cacheChart()
		started := make(chan struct{}, 2)
		release := make(chan struct{})
		var releaseOnce sync.Once
		unblock := func() { releaseOnce.Do(func() { close(release) }) }
		DeferCleanup(unblock)
		done := make(chan *api.Error, 2)
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			started <- struct{}{}
			<-release
			return archive, nil
		}, 0)
		for n := 0; n < 2; n++ {
			go func() {
				_, err := inspector.Inspect(context.Background(), Selection{ChartURL: "https://example.com/app.tgz"}, "tenant", "")
				done <- err
			}()
		}
		for n := 0; n < 2; n++ {
			select {
			case <-started:
			case <-time.After(time.Second):
				Fail("fresh requests shared one acquisition")
			}
		}
		unblock()
		for n := 0; n < 2; n++ {
			{
				err := <-done
				Expect(err).NotTo(HaveOccurred())
			}
		}
	})
	It("Inspector Cache Isolation Mutation And Invalidation", func() {
		archive := cacheChart()
		var calls atomic.Int32
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) { calls.Add(1); return archive, nil }, time.Minute)
		selection := ngcCacheSelection()
		first, err := inspector.Inspect(context.Background(), selection, "tenant1:caller1", "key1")
		Expect(err).NotTo(HaveOccurred())
		Expect(len(first.Questions)).To(Equal(1), func() string {
			return fmt.Sprintf("fixture declaration was not parsed: %+v", first)
		})

		first.ChartName = "mutated"
		first.Questions[0]["label"] = "Changed"
		second, err := inspector.Inspect(context.Background(), selection, "tenant1:caller1", "key1")
		Expect(err).NotTo(HaveOccurred())
		Expect(second.ChartName != fixtureChartName || second.Questions[0]["label"] != "Replicas" || calls.Load() != 1).To(BeFalse(), func() string {
			return fmt.Sprintf("cache mutated or missed: %+v calls=%d", second, calls.Load())
		})

		for _, scope := range []string{"tenant2:caller1", "tenant1:caller2"} {
			{
				_, err := inspector.Inspect(context.Background(), selection, scope, "key1")
				Expect(err).NotTo(HaveOccurred())
			}
		}
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant1:caller1", "rotated-key")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls.Load()).To(BeNumerically("==", 4), func() string {
			return fmt.Sprintf("identity/rotation isolation failed: %d", calls.Load())
		})

		Expect(cache.ClearStore[chartInspectionMetadata]()).To(Succeed())
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant1:caller1", "key1")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls.Load()).To(BeNumerically("==", 5), func() string {
			return fmt.Sprintf("invalidated hit: %d", calls.Load())
		})

		key := inspector.key(selection, "tenant1:caller1", "key1")
		Expect(key).NotTo(ContainSubstring("key1"))
		entry, cacheErr := cache.Get[chartInspectionMetadata](context.Background(), key)
		Expect(cacheErr).NotTo(HaveOccurred())
		Expect(entry).NotTo(BeNil())
		Expect(string(entry.Data)).NotTo(ContainSubstring("key1"))
	})
	It("Inspector Preserves Integer Defaults Across Cache Hits", func() {
		archive := chartArchive(
			archiveEntry{name: "app/Chart.yaml", body: "name: app\nversion: '1'\n"},
			archiveEntry{name: "app/questions.yaml", body: "questions:\n  - variable: counter\n    type: int\n    label: Counter\n    default: 9007199254740993\n"},
		)
		var calls int
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			calls++
			return archive, nil
		}, time.Minute)
		for n := 0; n < 2; n++ {
			result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())

			encoded, marshalErr := json.Marshal(result.Questions[0]["default"])
			Expect(marshalErr).NotTo(HaveOccurred())
			Expect(string(encoded)).To(Equal("9007199254740993"), func() string {
				return fmt.Sprintf("default changed during inspection/cache: %s", encoded)
			})
		}
		Expect(calls).To(Equal(1), func() string {
			return fmt.Sprintf("expected the second response to use cached metadata, downloads=%d", calls)
		})
	})
	It("Inspector Retries Failures", func() {
		archive := cacheChart()
		var calls int
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			calls++
			if calls == 1 {
				return nil, inspectionError(errors.New("temporary failure"))
			}
			if calls == 2 {
				return []byte("invalid archive"), nil
			}
			return archive, nil
		}, time.Minute)
		selection := ngcCacheSelection()
		for n := 0; n < 2; n++ {
			{
				_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
				Expect(err).NotTo(BeNil(), "failure accepted")
			}
		}
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
			Expect(err).NotTo(HaveOccurred())
		}
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
			Expect(err != nil || calls != 3).To(BeFalse(), func() string {
				return fmt.Sprintf("successful retry not cached: %d %v", calls, err)
			})
		}
	})
	It("Inspector Shares Misses And Keeps Waiter Alive", func() {
		archive := cacheChart()
		started := make(chan struct{})
		release := make(chan struct{})
		var calls atomic.Int32
		inspector := NewInspector(func(ctx context.Context, _ Selection, _ string) ([]byte, *api.Error) {
			if calls.Add(1) == 1 {
				close(started)
			}
			select {
			case <-release:
				return archive, nil
			case <-ctx.Done():
				return nil, inspectionError(ctx.Err())
			}
		}, time.Minute)
		selection := ngcCacheSelection()
		ctx, cancel := context.WithCancel(context.Background())
		firstDone := make(chan *api.Error, 1)
		go func() { _, err := inspector.Inspect(ctx, selection, "tenant", "key"); firstDone <- err }()
		<-started
		const others = 7
		var wg sync.WaitGroup
		errs := make(chan *api.Error, others)
		for n := 0; n < others; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r, err := inspector.Inspect(context.Background(), selection, "tenant", "key")
				if err == nil && r.ChartName != fixtureChartName {
					err = api.NewError(http.StatusInternalServerError, "wrong result", nil)
				}
				errs <- err
			}()
		}
		deadline := time.Now().Add(time.Second)
		for {
			inspector.mu.Lock()
			waiters := 0
			for _, f := range inspector.flights {
				waiters = f.waiters
			}
			inspector.mu.Unlock()
			if waiters == 8 {
				break
			}
			Expect(time.Now().After(deadline)).To(BeFalse(),
				"waiters did not attach",
			)

			time.Sleep(time.Millisecond)
		}
		cancel()
		{
			err := <-firstDone
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(BeNumerically("==", http.StatusRequestTimeout))
		}

		close(release)
		wg.Wait()
		close(errs)
		for err := range errs {
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls.Load()).To(BeNumerically("==", 1), func() string {
			return fmt.Sprintf("shared %d downloads", calls.Load())
		})
	})
	It("Inspector Cancels Abandoned Flight And Retries", func() {
		archive := cacheChart()
		started := make(chan struct{})
		abandoned := make(chan struct{})
		var calls atomic.Int32
		inspector := NewInspector(func(ctx context.Context, _ Selection, _ string) ([]byte, *api.Error) {
			if calls.Add(1) == 1 {
				close(started)
				<-ctx.Done()
				close(abandoned)
				return nil, inspectionError(ctx.Err())
			}
			return archive, nil
		}, time.Minute)
		selection := ngcCacheSelection()
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan *api.Error, 1)
		go func() { _, err := inspector.Inspect(ctx, selection, "tenant", ""); done <- err }()
		<-started
		cancel()
		{
			err := <-done
			Expect(err).To(HaveOccurred())
			Expect(err.Code).To(BeNumerically("==", http.StatusRequestTimeout))
		}

		select {
		case <-abandoned:
		case <-time.After(time.Second):
			Fail("abandoned download survived")
		}
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls.Load()).To(BeNumerically("==", 2), func() string {
			return fmt.Sprintf("cancelled miss reused %d", calls.Load())
		})
	})
	It("Inspector Does Not Share Different Selections", func() {
		archive := cacheChart()
		var calls atomic.Int32
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) { calls.Add(1); return archive, nil }, time.Minute)
		for _, sel := range []Selection{{ChartID: "one"}, {ChartID: "two"}} {
			{
				_, err := inspector.Inspect(context.Background(), sel, "tenant", "")
				Expect(err).NotTo(HaveOccurred())
			}
		}
		Expect(calls.Load()).To(BeNumerically("==", 2), "distinct selections shared")
		{
			_, err := inspector.Inspect(context.Background(), Selection{Repo: fixtureRepositoryURL, ChartName: "wrong", ChartVersion: "1"}, "tenant", "")
			Expect(err).NotTo(BeNil(), "wrong package identity accepted")
		}
	})
	It("Inspector Bounds Active Work", func() {
		archive := cacheChart()
		release := make(chan struct{})
		started := make(chan struct{}, 4)
		inspector := NewInspector(func(ctx context.Context, _ Selection, _ string) ([]byte, *api.Error) {
			started <- struct{}{}
			select {
			case <-release:
				return archive, nil
			case <-ctx.Done():
				return nil, inspectionError(ctx.Err())
			}
		}, time.Minute)
		var wg sync.WaitGroup
		for n := 0; n < 4; n++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				_, _ = inspector.Inspect(context.Background(), Selection{ChartID: string(rune('a' + n))}, "tenant", "")
			}(n)
		}
		for n := 0; n < 4; n++ {
			<-started
		}
		{
			_, err := inspector.Inspect(context.Background(), Selection{ChartID: "fifth"}, "tenant", "")
			Expect(err).NotTo(BeNil(), "unbounded active work")
		}

		close(release)
		wg.Wait()
	})
	It("Inspector Rejects Identity Mismatch Without Caching", func() {
		wrong, correct := cacheChart(), cacheChartMetadata(fixtureChartName, "2")
		var calls int
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			calls++
			if calls < 3 {
				return wrong, nil
			}
			return correct, nil
		}, time.Minute)
		selection := Selection{Repo: "https://helm.ngc.nvidia.com/nvidia", ChartName: fixtureChartName, ChartVersion: "2"}
		for n := 0; n < 2; n++ {
			_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
			var typed *api.Error
			Expect(!errors.As(err, &typed) || typed.Code != http.StatusBadRequest || !strings.HasPrefix(typed.Message, "chart_identity_mismatch:")).To(BeFalse(), func() string {
				return fmt.Sprintf("wrong mismatch error: %v", err)
			})
		}
		result, err := inspector.Inspect(context.Background(), selection, "tenant", "")
		Expect(err != nil || result.ChartVersion != "2" || calls != 3).To(BeFalse(), func() string {
			return fmt.Sprintf("failed identity was cached: %+v %v calls=%d", result, err, calls)
		})
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
			Expect(err != nil || calls != 3).To(BeFalse(), func() string {
				return fmt.Sprintf("correct identity not cached: %v calls=%d", err, calls)
			})
		}
	})
	It("Inspector Serves Uploaded Callback Fresh", func() {

		archive := cacheChart()
		var calls int
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			defer GinkgoRecover()
			Fail("remote fetch used for authorized upload")
			return nil, inspectionError(errors.New("wrong source"))
		}, time.Minute)
		uploaded := func(context.Context, Selection, string) ([]byte, *api.Error) { calls++; return archive, nil }
		for n := 0; n < 2; n++ {
			result, err := inspector.InspectWithFetch(context.Background(), Selection{ChartID: fixtureUploadID}, "tenant", "", uploaded)
			Expect(err != nil || result.ChartName != fixtureChartName).To(BeFalse(), func() string {
				return fmt.Sprintf("uncacheable result lost: %+v %v", result, err)
			})
		}
		Expect(calls).To(Equal(2), func() string {
			return fmt.Sprintf("upload was cached: %d", calls)
		})
		entry, cacheErr := cache.Get[chartInspectionMetadata](context.Background(), inspector.key(Selection{ChartID: fixtureUploadID}, "tenant", ""))
		Expect(cacheErr).NotTo(HaveOccurred())
		Expect(entry).To(BeNil(), "upload populated cache")
	})
	It("serves metadata rejected by the shared cache record limit", func() {
		archive := chartArchive(
			archiveEntry{name: "app/Chart.yaml", body: "name: app\nversion: '1'\n"},
			archiveEntry{name: "app/questions.yaml", body: "questions:\n- variable: description\n  label: Description\n  type: string\n  default: " + strings.Repeat("x", 2048) + "\n"},
		)
		calls := 0
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, *api.Error) {
			calls++
			return archive, nil
		}, 0)
		Expect(cache.RegisterStore[chartInspectionMetadata](options.MaxRecordSizeKB(1))).To(Succeed())
		for n := 0; n < 2; n++ {
			result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.ChartName).To(Equal(fixtureChartName))
		}
		Expect(calls).To(Equal(2))
		entry, err := cache.Get[chartInspectionMetadata](context.Background(), inspector.key(ngcCacheSelection(), "tenant", ""))
		Expect(err).NotTo(HaveOccurred())
		Expect(entry).To(BeNil())
	})
})
