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

	validity_gate_monitor "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/validity-gate-monitor"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
)

// checkMetricSources returns two results: a gate over the validity-gate monitor's
// CriticalMetricSources allowlist, and an informational sweep of every other scrape
// target. The sweep never blocks an environment — some targets are legitimately absent,
// e.g. inference-workload-controller when its kubeflow/knative dependency is not
// installed — but a surprise there is worth seeing before a multi-hour run.
func (c *Checker) checkMetricSources(ctx context.Context) (gate CheckResult, sweep CheckResult) {
	gate = CheckResult{Check: "critical_metric_sources", Status: StatusPass}
	sweep = CheckResult{Check: "other_scrape_targets", Status: StatusPass}

	up, err := c.queryUpByJob(ctx)
	if err != nil {
		gate = failf(gate, "querying Prometheus: %v", err)
		sweep.Status = StatusWarn
		sweep.Summary = "skipped: Prometheus query failed"
		return gate, sweep
	}

	critical := map[string]bool{}
	for _, job := range validity_gate_monitor.CriticalMetricSources {
		critical[job] = true
		healthy, present := up[job]
		switch {
		case !present:
			gate.Status = StatusFail
			gate.Detail = append(gate.Detail, fmt.Sprintf("%s: no scrape target", job))
		case !healthy:
			gate.Status = StatusFail
			gate.Detail = append(gate.Detail, fmt.Sprintf("%s: up==0", job))
		}
	}
	sort.Strings(gate.Detail)
	gate.Summary = fmt.Sprintf("%d required sources, %d unhealthy", len(critical), len(gate.Detail))

	for job, healthy := range up {
		if critical[job] || healthy {
			continue
		}
		sweep.Status = StatusWarn
		sweep.Detail = append(sweep.Detail, fmt.Sprintf("%s: up==0", job))
	}
	sort.Strings(sweep.Detail)
	sweep.Summary = fmt.Sprintf("%d non-critical targets down (informational, does not block)", len(sweep.Detail))
	return gate, sweep
}

// checkKubeletRealNodes gates kubelet scraping on real nodes only. The job as a whole
// cannot be gated: KWOK nodes have no kubelet, so their targets are permanently down. But
// real-node kubelets feed the cAdvisor container metrics behind the hardware gate, so a
// failure there must not hide in the WARN sweep.
func (c *Checker) checkKubeletRealNodes(ctx context.Context) CheckResult {
	res := CheckResult{Check: "kubelet_real_nodes", Status: StatusPass}

	nodes := &corev1.NodeList{}
	if err := c.clients.KubeClient.List(ctx, nodes); err != nil {
		return failf(res, "listing nodes: %v", err)
	}
	byNode, err := c.queryUpByNode(ctx, "kubelet")
	if err != nil {
		return failf(res, "querying kubelet targets: %v", err)
	}

	total := 0
	for i := range nodes.Items {
		n := &nodes.Items[i]
		if n.Labels["type"] == "kwok" {
			continue
		}
		total++
		healthy, found := byNode[n.Name]
		switch {
		case !found:
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf("%s: no kubelet scrape target", n.Name))
		case !healthy:
			res.Status = StatusFail
			res.Detail = append(res.Detail, fmt.Sprintf("%s: kubelet down", n.Name))
		}
	}
	if total == 0 {
		return failf(res, "no real (non-kwok) nodes found")
	}

	sort.Strings(res.Detail)
	res.Summary = fmt.Sprintf("%d/%d real-node kubelets healthy", total-len(res.Detail), total)
	return res
}

// queryUpByNode returns node name -> healthy for one job, where healthy means every target
// on that node is up.
func (c *Checker) queryUpByNode(ctx context.Context, job string) (map[string]bool, error) {
	out := map[string]bool{}

	err := validity_gate_monitor.WithPrometheusAPI(
		ctx,
		c.cfg.PrometheusNamespace,
		c.cfg.PrometheusPodName,
		c.cfg.PrometheusPort,
		c.restConfig,
		func(api v1.API) error {
			q := fmt.Sprintf("min by (node) (up{job=%q})", job)
			result, _, _, err := api.Query(ctx, q, time.Now())
			if err != nil {
				return fmt.Errorf("query %q failed: %w", q, err)
			}
			vec, ok := result.(model.Vector)
			if !ok {
				return fmt.Errorf("unexpected result type %T for %q", result, q)
			}
			for _, s := range vec {
				out[string(s.Metric["node"])] = s.Value == 1
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// queryUpByJob returns job name -> healthy. A job is healthy only if every one of its
// instances is up, hence min rather than max.
func (c *Checker) queryUpByJob(ctx context.Context) (map[string]bool, error) {
	out := map[string]bool{}

	err := validity_gate_monitor.WithPrometheusAPI(
		ctx,
		c.cfg.PrometheusNamespace,
		c.cfg.PrometheusPodName,
		c.cfg.PrometheusPort,
		c.restConfig,
		func(api v1.API) error {
			result, _, _, err := api.Query(ctx, "min by (job) (up)", time.Now())
			if err != nil {
				return fmt.Errorf("up query failed: %w", err)
			}
			vec, ok := result.(model.Vector)
			if !ok {
				return fmt.Errorf("unexpected result type %T for up query", result)
			}
			for _, s := range vec {
				out[string(s.Metric["job"])] = s.Value == 1
			}
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return out, nil
}
