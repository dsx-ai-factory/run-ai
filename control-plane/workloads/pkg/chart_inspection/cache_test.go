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
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func cacheChart() []byte {
	return cacheChartMetadata(fixtureChartName, "1")
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
				inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
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
					Expect(len(inspector.entries)).To(Equal(0), "fresh source populated cache")
				}
			}
		}
	})
	It("Inspector Default NGC Cache Expires After Ten Minutes", func() {
		archive := cacheChart()
		calls := 0
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) { calls++; return archive, nil }, 0)
		now := time.Unix(100, 0)
		inspector.now = func() time.Time { return now }
		for _, second := range []int{0, 60, 599, 600} {
			now = time.Unix(100+int64(second), 0)
			{
				_, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant:caller", "")
				Expect(err).NotTo(HaveOccurred())
			}

			want := 1
			if second == 600 {
				want = 2
			}
			Expect(calls).To(Equal(want), func() string {
				return fmt.Sprintf("at %ds: acquisitions=%d, want=%d", second, calls, want)
			})
		}
	})
	It("Inspector Fresh Requests Do Not Share An Acquisition", func() {
		archive := cacheChart()
		started := make(chan struct{}, 2)
		release := make(chan struct{})
		var releaseOnce sync.Once
		unblock := func() { releaseOnce.Do(func() { close(release) }) }
		DeferCleanup(unblock)
		done := make(chan error, 2)
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
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
	It("Inspector Cache Expiry Isolation And Mutation", func() {
		archive := cacheChart()
		var calls atomic.Int32
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) { calls.Add(1); return archive, nil }, time.Minute)
		now := time.Unix(100, 0)
		inspector.now = func() time.Time { return now }
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

		now = now.Add(time.Minute)
		{
			_, err := inspector.Inspect(context.Background(), selection, "tenant1:caller1", "key1")
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(calls.Load()).To(BeNumerically("==", 5), func() string {
			return fmt.Sprintf("expired hit: %d", calls.Load())
		})

		inspector.mu.Lock()
		defer inspector.mu.Unlock()
		for key := range inspector.entries {
			Expect(strings.Contains(key, "key1") || strings.Contains(key, "rotated-key")).To(BeFalse(),
				"plaintext key cached",
			)
		}
	})
	It("Inspector Preserves Integer Defaults Across Cache Hits", func() {
		archive := chartArchive(
			archiveEntry{name: "app/Chart.yaml", body: "name: app\nversion: '1'\n"},
			archiveEntry{name: "app/questions.yaml", body: "questions:\n  - variable: counter\n    type: int\n    label: Counter\n    default: 9007199254740993\n"},
		)
		var calls int
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
			calls++
			return archive, nil
		}, time.Minute)
		for n := 0; n < 2; n++ {
			result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())

			encoded, err := json.Marshal(result.Questions[0]["default"])
			Expect(err).NotTo(HaveOccurred())
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
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
			calls++
			if calls == 1 {
				return nil, errors.New("temporary failure")
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
		inspector := NewInspector(func(ctx context.Context, _ Selection, _ string) ([]byte, error) {
			if calls.Add(1) == 1 {
				close(started)
			}
			select {
			case <-release:
				return archive, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}, time.Minute)
		selection := ngcCacheSelection()
		ctx, cancel := context.WithCancel(context.Background())
		firstDone := make(chan error, 1)
		go func() { _, err := inspector.Inspect(ctx, selection, "tenant", "key"); firstDone <- err }()
		<-started
		const others = 7
		var wg sync.WaitGroup
		errs := make(chan error, others)
		for n := 0; n < others; n++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				r, err := inspector.Inspect(context.Background(), selection, "tenant", "key")
				if err == nil && r.ChartName != fixtureChartName {
					err = errors.New("wrong result")
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
			Expect(errors.Is(err, context.Canceled)).To(BeTrue(), func() string {
				return fmt.Sprintf("leader cancellation lost %v", err)
			})
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
		inspector := NewInspector(func(ctx context.Context, _ Selection, _ string) ([]byte, error) {
			if calls.Add(1) == 1 {
				close(started)
				<-ctx.Done()
				close(abandoned)
				return nil, ctx.Err()
			}
			return archive, nil
		}, time.Minute)
		selection := ngcCacheSelection()
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan error, 1)
		go func() { _, err := inspector.Inspect(ctx, selection, "tenant", ""); done <- err }()
		<-started
		cancel()
		{
			err := <-done
			Expect(errors.Is(err, context.Canceled)).To(BeTrue(), func() string {
				return fmt.Sprint(err)
			})
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
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) { calls.Add(1); return archive, nil }, time.Minute)
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
	It("Inspector Bounds Active And Cached Work", func() {
		archive := cacheChart()
		release := make(chan struct{})
		started := make(chan struct{}, 4)
		inspector := NewInspector(func(ctx context.Context, _ Selection, _ string) ([]byte, error) {
			started <- struct{}{}
			select {
			case <-release:
				return archive, nil
			case <-ctx.Done():
				return nil, ctx.Err()
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
		inspector.maxEntries = 2
		inspector.fetch = func(context.Context, Selection, string) ([]byte, error) { return archive, nil }
		for _, id := range []string{"x", "y", "z"} {
			selection := ngcCacheSelection()
			selection.Repo += "/" + id
			{
				_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
				Expect(err).NotTo(HaveOccurred())
			}
		}
		inspector.mu.Lock()
		count := len(inspector.entries)
		inspector.mu.Unlock()
		Expect(count).To(Equal(2), func() string {
			return fmt.Sprintf("unbounded entries %d", count)
		})

		inspector.maxRecordBytes = 1
		{
			_, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
			Expect(err).NotTo(HaveOccurred())
		}

		inspector.mu.Lock()
		defer inspector.mu.Unlock()
		Expect(inspector.cachedBytes > inspector.maxCacheBytes).To(BeFalse(),
			"unbounded cached bytes",
		)
	})
	It("Inspector Rejects Identity Mismatch Without Caching", func() {
		wrong, correct := cacheChart(), cacheChartMetadata(fixtureChartName, "2")
		var calls int
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
			calls++
			if calls < 3 {
				return wrong, nil
			}
			return correct, nil
		}, time.Minute)
		selection := Selection{Repo: "https://helm.ngc.nvidia.com/nvidia", ChartName: fixtureChartName, ChartVersion: "2"}
		for n := 0; n < 2; n++ {
			_, err := inspector.Inspect(context.Background(), selection, "tenant", "")
			var typed *Error
			Expect(!errors.As(err, &typed) || typed.StatusCode != 422 || typed.Code != "chart_identity_mismatch").To(BeFalse(), func() string {
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
		inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
			defer GinkgoRecover()
			Fail("remote fetch used for authorized upload")
			return nil, errors.New("wrong source")
		}, time.Minute)
		uploaded := func(context.Context, Selection, string) ([]byte, error) { calls++; return archive, nil }
		for n := 0; n < 2; n++ {
			result, err := inspector.InspectWithFetch(context.Background(), Selection{ChartID: fixtureUploadID}, "tenant", "", uploaded)
			Expect(err != nil || result.ChartName != fixtureChartName).To(BeFalse(), func() string {
				return fmt.Sprintf("uncacheable result lost: %+v %v", result, err)
			})
		}
		Expect(calls).To(Equal(2), func() string {
			return fmt.Sprintf("upload was cached: %d", calls)
		})
		Expect(len(inspector.entries)).To(Equal(0), "upload populated cache")
	})
	It("Inspector Serves NGC Result Exceeding Cache Budgets", func() {
		archive := cacheChart()
		for _, budget := range []string{"record", "total"} {
			{
				By(budget)
				calls := 0
				inspector := NewInspector(func(context.Context, Selection, string) ([]byte, error) {
					calls++
					return archive, nil
				}, 0)
				if budget == "record" {
					inspector.maxRecordBytes = 1
				} else {
					inspector.maxCacheBytes = 1
				}
				for n := 0; n < 2; n++ {
					result, err := inspector.Inspect(context.Background(), ngcCacheSelection(), "tenant", "")
					Expect(err != nil || result.ChartName != fixtureChartName).To(BeFalse(), func() string {
						return fmt.Sprintf("uncacheable result lost: %+v %v", result, err)
					})
				}
				Expect(calls != 2 || len(inspector.entries) != 0).To(BeFalse(), func() string {
					return fmt.Sprintf("cache %s budget ignored: acquisitions=%d entries=%d", budget, calls, len(inspector.entries))
				})
			}
		}
	})
})
