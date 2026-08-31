/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package readiness

import (
	"context"
	"fmt"
	"sort"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// checkMetricsServer verifies metrics-server is feeding the VPA recommender.
//
// metrics.k8s.io PodMetrics are VPA's entire input in the default (non-Prometheus) mode:
// once per --recommender-interval it LISTs them and folds each container's usage into that
// container's histograms. So this reads VPA's actual input, not a proxy.
//
// Coverage and freshness are checked because both fail silently. Partial coverage leaves
// uncovered containers pinned at the resource floor, looking like real measurements in the
// lookup table; a stale metrics-server keeps VPA recommending from an old histogram that
// never saw the run's load. Inferring health from VPA output instead ("is some
// recommendation above the floor?") is an existence check that one healthy container
// satisfies, so it catches neither. Coverage cannot be enforced on the VPA side either —
// a genuinely idle sidecar legitimately sits at the floor forever.

// podMetricsMaxAge is how stale the newest sample may be before metrics-server counts as
// not serving. Resolution is 15s and VPA samples once per minute.
const podMetricsMaxAge = 2 * time.Minute

// podMetricsGracePeriod is how long a pod may be Ready before its metrics are required.
// metrics-server needs two scrapes to compute a CPU rate, so a freshly-Ready pod has none.
const podMetricsGracePeriod = 2 * time.Minute

func (c *Checker) checkMetricsServer(ctx context.Context) CheckResult {
	res := CheckResult{Check: "metrics_server", Status: StatusPass}

	gvr := schema.GroupVersionResource{Group: "metrics.k8s.io", Version: "v1beta1", Resource: "pods"}

	covered, expected := 0, 0
	var newest time.Time

	for _, ns := range c.namespaces {
		pods, err := c.listPods(ctx, ns)
		if err != nil {
			return failf(res, "listing pods in %s: %v", ns, err)
		}
		if len(pods) == 0 {
			continue
		}

		// Only pods Ready for longer than the grace period are expected to have metrics.
		want := map[string]bool{}
		for i := range pods {
			if since, ok := podReadySince(&pods[i]); ok && time.Since(since) > podMetricsGracePeriod {
				want[pods[i].Name] = true
			}
		}
		expected += len(want)

		list, err := c.clients.DynamicClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			// The metrics.k8s.io APIService is absent or unavailable — the most likely
			// reason an entire run yields floor-value recommendations, so fail loudly.
			return failf(res, "listing metrics.k8s.io pod metrics in %s: %v", ns, err)
		}

		for _, item := range list.Items {
			if want[item.GetName()] {
				covered++
				delete(want, item.GetName())
			}
			if ts, ok := unstructuredTime(item.Object, "timestamp"); ok && ts.After(newest) {
				newest = ts
			}
		}

		for name := range want {
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf("%s/%s: Ready but no PodMetrics", ns, name))
		}
	}

	if expected == 0 {
		return failf(res, "no pods have been Ready longer than %v; cannot assess metrics-server", podMetricsGracePeriod)
	}
	if newest.IsZero() {
		res.Status = StatusFail
		res.Detail = append(res.Detail, "no PodMetrics carried a timestamp; cannot establish freshness")
	} else if age := time.Since(newest); age > podMetricsMaxAge {
		res.Status = StatusFail
		res.Detail = append(res.Detail, fmt.Sprintf(
			"newest PodMetrics sample is %v old (max %v) — metrics-server has stopped refreshing",
			age.Round(time.Second), podMetricsMaxAge))
	}

	sort.Strings(res.Detail)
	age := "n/a"
	if !newest.IsZero() {
		age = time.Since(newest).Round(time.Second).String()
	}
	res.Summary = fmt.Sprintf("metrics for %d/%d pods Ready over %v, newest sample %s old",
		covered, expected, podMetricsGracePeriod, age)
	return res
}

// unstructuredTime reads an RFC3339 timestamp field from an unstructured object.
func unstructuredTime(obj map[string]any, field string) (time.Time, bool) {
	raw, ok := obj[field].(string)
	if !ok || raw == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
