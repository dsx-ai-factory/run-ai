/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package readiness gates whether a provisioned environment has actually settled and is
// fit to start a measurement run against.
//
// It exists because the validity-gate monitor cannot serve this purpose: its pod-restart
// gate has a threshold of zero, and remove-limits plus the Prometheus roll both restart
// pods by design. Run it too early and it fails a perfectly good environment; run it late
// enough to avoid that and it will happily pass one that is still stabilising.
//
// Checks are grouped by subject across checks_*.go files.
package readiness

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/resource-rec"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Checker runs every readiness check and assembles a Report.
type Checker struct {
	cfg        *Config
	cluster    *scale.ClusterConfig
	clients    *sampler_actions.Clients
	restConfig *rest.Config
	namespaces []string
}

func NewChecker(cfg *Config, cluster *scale.ClusterConfig, clients *sampler_actions.Clients, restConfig *rest.Config) *Checker {
	ns := cfg.Namespaces
	if len(ns) == 0 {
		ns = resource_rec.DefaultNamespaces
	}
	return &Checker{cfg: cfg, cluster: cluster, clients: clients, restConfig: restConfig, namespaces: ns}
}

// Run executes all checks and returns the assembled report. Checks are independent: a
// failure in one does not skip the rest, so a single run reports everything that is wrong
// rather than only the first problem found.
func (c *Checker) Run(ctx context.Context) *Report {
	report := &Report{}

	// The restart-stability check needs one sample before the quiet period and one after.
	// Take the first up front so the wait overlaps the other checks rather than being
	// tacked onto the end.
	firstSample, sampleErr := c.sampleRestartCounts(ctx)
	sampleStart := time.Now()

	report.Add(c.checkWorkloadsReady(ctx))
	report.Add(c.checkPodsReady(ctx))
	report.Add(c.checkNodepools(ctx))
	report.Add(c.checkProjects(ctx))
	report.Add(c.checkClusterConnected(ctx))
	report.Add(c.checkMetricsServer(ctx))

	gate, sweep := c.checkMetricSources(ctx)
	report.Add(gate)
	report.Add(c.checkKubeletRealNodes(ctx))
	report.Add(sweep)

	report.Add(c.checkRestartsStable(ctx, firstSample, sampleErr, sampleStart))

	report.Finalize()
	return report
}

// --- shared helpers ---

func (c *Checker) namespaceExists(ctx context.Context, ns string) (bool, error) {
	obj := &corev1.Namespace{}
	err := c.clients.KubeClient.Get(ctx, client.ObjectKey{Name: ns}, obj)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// listPods returns the pods in ns, or nil if the namespace does not exist in this cluster.
func (c *Checker) listPods(ctx context.Context, ns string) ([]corev1.Pod, error) {
	exists, err := c.namespaceExists(ctx, ns)
	if err != nil || !exists {
		return nil, err
	}
	pods := &corev1.PodList{}
	if err := c.clients.KubeClient.List(ctx, pods, client.InNamespace(ns)); err != nil {
		return nil, err
	}
	return pods.Items, nil
}

func podReady(p *corev1.Pod) bool {
	for _, cond := range p.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// podReadySince returns when the pod's Ready condition last transitioned to true.
func podReadySince(p *corev1.Pod) (time.Time, bool) {
	for _, cond := range p.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
			return cond.LastTransitionTime.Time, true
		}
	}
	return time.Time{}, false
}

func failf(res CheckResult, format string, args ...any) CheckResult {
	res.Status = StatusFail
	res.Summary = fmt.Sprintf(format, args...)
	return res
}
