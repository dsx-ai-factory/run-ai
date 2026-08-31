/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validity_gate_monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Degraded only when depth stays above the threshold for the whole window.
const (
	workQueueDepthThreshold = 25
	workQueueSustainedSecs  = 120
)

// Placeholder from observed data: p99 across all non-longpoll verbs was ~0.6s (LIST) on a
// healthy large run. Retune once more runs land.
const apiserverLatencyP99Threshold = 2.0

// Above this, Prometheus missed enough scrapes that any metric-derived gate reading clean over
// the same window is reporting absence of data rather than absence of a fault.
const prometheusScrapeGapThreshold = 30.0

// Monitor runs periodic heartbeat checks during the test and assembles the
// final structured report when the run is complete.
type Monitor struct {
	cfg        *MonitorConfig
	restConfig *rest.Config

	// knownPods maps pod UID → "namespace/pod" as of the previous heartbeat. A UID that
	// disappears means the pod was replaced, not restarted.
	knownPods map[string]string
	// replacements accumulates what was detected, for the final report.
	replacements []string

	// Completion progress is measured across heartbeats and cannot be recomputed at the end:
	// the heartbeat cadence defines it, and a TTL deletes the evidence before the final
	// queries run. Carried here so it reaches the report, not just the log.
	//
	// Counting starts one workload duration in — before that nothing can have completed yet,
	// so the whole ramp would read as one long stall — and only while workloads are resident.
	startTime          time.Time
	zeroCompletionRun  int    // consecutive heartbeats with nothing completing
	worstCompletionGap int    // longest such run seen; a healthy run reports non-zero here
	progressGateReason string // set when the gate could not be evaluated

	// capturedRestarts maps "ns/pod/container" → the cumulative restart count at the last
	// capture, so the increase([5m]) window returning the same crash for two heartbeats
	// does not log and stream it twice.
	capturedRestarts map[string]float64
}

func NewMonitor(cfg *MonitorConfig, restConfig *rest.Config) *Monitor {
	return &Monitor{cfg: cfg, restConfig: restConfig}
}

// Run blocks until ctx is cancelled (either by SIGTERM or the configured RunDuration),
// firing a heartbeat every HeartbeatInterval. When ctx is done it performs the
// final range queries and returns the complete report.
func (m *Monitor) Run(ctx context.Context, startTime time.Time) (*ValidityGateReport, error) {
	interval, err := time.ParseDuration(m.cfg.HeartbeatInterval)
	if err != nil {
		return nil, fmt.Errorf("invalid heartbeat_interval %q: %w", m.cfg.HeartbeatInterval, err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	m.startTime = startTime
	m.validateProgressParams(interval)

	log.Printf("[validity-gate] monitor started at %s, heartbeat every %s", startTime.Format(time.RFC3339), interval)

	for {
		select {
		case <-ctx.Done():
			elapsed := time.Since(startTime)
			log.Printf("[validity-gate] run complete after %s — collecting final report", elapsed.Round(time.Second))
			return m.collectFinalReport(context.Background(), startTime)

		case t := <-ticker.C:
			log.Printf("[validity-gate] heartbeat at %s", t.Format(time.RFC3339))
			m.runHeartbeat(ctx)
		}
	}
}

// validateProgressParams records, at startup, the ways the completion gate can be silently
// useless — all of them decided by parameters the monitor has no say over. The reason is kept
// so it reaches the report rather than only the log, where a multi-hour run buries it.
//
// The TTL case is the subtle one. Completions are counted by asking, once a heartbeat, which
// pods finished inside the last heartbeat window and are still present to be seen. A TTL
// shorter than two heartbeats can delete a pod after it completed but before the heartbeat
// that would have counted it, so real completions read as zero and a healthy run trips the
// gate. That makes the gate unreliable rather than merely noisy, so it is disabled outright.
func (m *Monitor) validateProgressParams(heartbeat time.Duration) {
	if m.cfg.WorkloadDurationSecs <= 0 || m.cfg.CompletionGapThresholdSecs <= 0 {
		m.progressGateReason = "workload_duration_secs and completion_gap_threshold_secs must both be set"
	} else if ttl := time.Duration(m.cfg.AutoDeletionSecs) * time.Second; m.cfg.AutoDeletionSecs > 0 && ttl < 2*heartbeat {
		m.progressGateReason = fmt.Sprintf(
			"workload TTL %s is below 2x the heartbeat interval %s — completions can be deleted "+
				"before the heartbeat that would count them; raise the TTL or shorten the heartbeat",
			ttl, heartbeat)
	}
	if m.progressGateReason != "" {
		log.Printf("[validity-gate] [WARN] workload completion gate disabled: %s", m.progressGateReason)
	}
}

// runHeartbeat opens a short-lived port-forward, checks all validity metrics
// at their current instant values, and logs a PASS/FAIL line per metric.
func (m *Monitor) runHeartbeat(ctx context.Context) {
	hbCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	err := withPrometheusAPI(hbCtx, m.cfg, m.restConfig, func(api v1.API) error {
		m.checkApiserverLatency(hbCtx, api)
		m.checkEtcdWALFsync(hbCtx, api)
		m.checkEtcdProposalsPending(hbCtx, api)
		m.checkEtcdDBSize(hbCtx, api)
		m.checkNodeNotReady(hbCtx, api)
		m.checkPodRestarts(hbCtx, api)
		m.checkSystemPods(hbCtx, api)
		m.checkPrometheusScrapGaps(hbCtx, api)
		m.checkSchedulerCycleLatency(hbCtx, api)
		m.checkWorkloadSubmissionLatency(hbCtx, api)
		m.checkReconciliationLatency(hbCtx, api)
		m.checkWorkQueueDepth(hbCtx, api)
		m.checkPodReplacements(hbCtx, api)
		m.checkNodeCPUUtilization(hbCtx, api)
		m.checkNodeMemoryUtilization(hbCtx, api)
		m.checkWorkloadListLatency(hbCtx, api)
		m.checkWorkloadProgress(hbCtx, api)
		return nil
	})
	if err != nil {
		log.Printf("[validity-gate] [WARN] heartbeat failed to connect to Prometheus: %v", err)
	}
}

func (m *Monitor) checkApiserverLatency(ctx context.Context, api v1.API) {
	val, available, err := heartbeatApiserverLatencyP50(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] apiserver_latency_p50: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] apiserver_latency_p50: no data")
		return
	}
	status := StatusPass
	if val > 0.5 {
		status = StatusFail
	}
	log.Printf("[validity-gate] [%s] apiserver_request_latency_p50 value=%.3fs threshold=0.5s", status, val)
}

func (m *Monitor) checkEtcdWALFsync(ctx context.Context, api v1.API) {
	val, available, err := heartbeatEtcdWALFsyncP99(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] etcd_wal_fsync_p99: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] etcd_wal_fsync_p99: no data")
		return
	}
	status := StatusPass
	if val > 0.1 {
		status = StatusFail
	}
	log.Printf("[validity-gate] [%s] etcd_wal_fsync_p99 value=%.3fs threshold=0.1s", status, val)
}

func (m *Monitor) checkEtcdProposalsPending(ctx context.Context, api v1.API) {
	val, available, err := heartbeatEtcdProposalsPending(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] etcd_proposals_pending: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] etcd_proposals_pending: no data")
		return
	}
	status := StatusPass
	if val > float64(m.cfg.EtcdProposalDepthThreshold) {
		status = StatusFail
	}
	log.Printf("[validity-gate] [%s] etcd_proposals_pending value=%.0f threshold=%d",
		status, val, m.cfg.EtcdProposalDepthThreshold)
}

func (m *Monitor) checkEtcdDBSize(ctx context.Context, api v1.API) {
	val, available, err := heartbeatEtcdDBSizeRatio(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] etcd_db_size_ratio: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] etcd_db_size_ratio: no data")
		return
	}
	status := StatusPass
	if val > 0.70 {
		status = StatusFail
	}
	log.Printf("[validity-gate] [%s] etcd_db_size_ratio value=%.1f%% threshold=70%%", status, val*100)
}

func (m *Monitor) checkNodeNotReady(ctx context.Context, api v1.API) {
	val, available, err := heartbeatNodeNotReadyCount(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] node_not_ready: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] node_not_ready: no data")
		return
	}
	status := StatusPass
	if val > 0 {
		status = StatusFail
	}
	log.Printf("[validity-gate] [%s] node_not_ready count=%.0f threshold=0", status, val)
}

func (m *Monitor) checkPodRestarts(ctx context.Context, api v1.API) {
	val, available, err := heartbeatPodRestartsDelta(ctx, api, m.cfg.TargetNamespaces)
	if err != nil {
		log.Printf("[validity-gate] [WARN] pod_restarts: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] pod_restarts: no data")
		return
	}
	status := StatusPass
	if val > 0 {
		status = StatusFail
		// Collect crash logs immediately — previous-container logs are only guaranteed
		// to be on the node while the crash is recent. The final report will still list
		// which pods crashed (from Prometheus data) but logs may be gone by then.
		detail, detailErr := heartbeatPodRestartDetails(ctx, api, m.cfg.TargetNamespaces)
		if detailErr != nil {
			log.Printf("[validity-gate] [WARN] pod_restarts: could not get detail: %v", detailErr)
		} else {
			totals, totalsErr := heartbeatPodRestartTotals(ctx, api, m.cfg.TargetNamespaces)
			if totalsErr != nil {
				log.Printf("[validity-gate] [WARN] pod_restarts: could not get restart totals: %v", totalsErr)
			}
			fresh := m.freshCrashes(detail, totals)
			for _, d := range fresh {
				log.Printf("[validity-gate] [FAIL] pod_restarts: %s", d)
			}
			m.collectCrashLogs(ctx, fresh, m.cfg.OutputDir)
		}
	}
	log.Printf("[validity-gate] [%s] pod_restarts_last_5m value=%.0f threshold=0", status, val)
}

func (m *Monitor) checkSystemPods(ctx context.Context, api v1.API) {
	val, available, err := heartbeatSystemPodsUnhealthy(ctx, api, m.cfg.TargetNamespaces)
	if err != nil {
		log.Printf("[validity-gate] [WARN] system_pods_unhealthy: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] system_pods_unhealthy: no data")
		return
	}
	status := StatusPass
	if val > 0 {
		status = StatusFail
	}
	log.Printf("[validity-gate] [%s] system_pods_unhealthy value=%.0f threshold=0", status, val)
}

func (m *Monitor) checkPrometheusScrapGaps(ctx context.Context, api v1.API) {
	down, err := heartbeatMetricSourcesDown(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] metric_source_health: query error: %v", err)
		return
	}
	if len(down) == 0 {
		log.Printf("[validity-gate] [%s] metric_source_health: all sources up", StatusPass)
		return
	}
	for _, src := range down {
		log.Printf("[validity-gate] [%s] metric_source_health: %s is down — validity gate data unreliable", StatusFail, src)
	}
}

func (m *Monitor) checkSchedulerCycleLatency(ctx context.Context, api v1.API) {
	avg, peak, available, err := heartbeatSchedulerCycleLatency(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] scheduler_allocate_cycle_latency_ms: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] scheduler_allocate_cycle_latency_ms: no data")
		return
	}
	status := StatusPass
	if peak > m.cfg.SchedulerCyclePeakThresholdMs {
		status = StatusDegraded
	}
	log.Printf("[validity-gate] [%s] scheduler_allocate_cycle_latency_ms avg=%.3fms peak=%.3fms threshold=%.0fms",
		status, avg, peak, m.cfg.SchedulerCyclePeakThresholdMs)
}

func (m *Monitor) checkWorkloadSubmissionLatency(ctx context.Context, api v1.API) {
	p99, available, err := heartbeatWorkloadSubmissionLatencyP99(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] workload_submission_latency_p99: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [UNAVAILABLE] workload_submission_latency_p99: no data (no workloads submitted yet)")
		return
	}
	status := StatusPass
	if p99 > m.cfg.SubmissionLatencyP99Threshold {
		status = StatusDegraded
	}
	log.Printf("[validity-gate] [%s] workload_submission_latency_p99 value=%.3fs threshold=%.1fs",
		status, p99, m.cfg.SubmissionLatencyP99Threshold)
}

func (m *Monitor) checkReconciliationLatency(ctx context.Context, api v1.API) {
	latencies, err := heartbeatReconciliationLatencyP99(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] reconciliation_latency_p99: query error: %v", err)
		return
	}
	if latencies == nil {
		log.Printf("[validity-gate] [UNAVAILABLE] reconciliation_latency_p99: no data")
		return
	}
	for controller, latency := range latencies {
		if math.IsNaN(latency) {
			continue
		}
		status := StatusPass
		if latency > m.cfg.ReconciliationLatencyP99Threshold {
			status = StatusDegraded
		}
		log.Printf("[validity-gate] [%s] reconciliation_latency_p99{controller=%q} p99=%.3fs threshold=%.1fs",
			status, controller, latency, m.cfg.ReconciliationLatencyP99Threshold)
	}
}

// checkPodReplacements detects pods that vanished and were recreated by their owner.
// Such a pod starts with restartCount 0, so the pod_restarts gate never sees it. Evidence
// must be collected now: once the pod object is gone its logs are unrecoverable, and
// Kubernetes events expire in about an hour.
func (m *Monitor) checkPodReplacements(ctx context.Context, api v1.API) {
	current, err := heartbeatPodUIDs(ctx, api, m.cfg.TargetNamespaces)
	if err != nil {
		log.Printf("[validity-gate] [WARN] pod_replacements: query error: %v", err)
		return
	}
	if current == nil {
		return
	}
	if m.knownPods == nil {
		m.knownPods = current
		return
	}

	var gone []string
	for uid, name := range m.knownPods {
		if _, still := current[uid]; !still {
			gone = append(gone, name)
		}
	}
	m.knownPods = current
	if len(gone) == 0 {
		return
	}
	sort.Strings(gone)
	for _, name := range gone {
		log.Printf("[validity-gate] [FAIL] pod_replacements: %s disappeared (replaced, not restarted)", name)
	}
	m.replacements = append(m.replacements, gone...)
	m.collectReplacementEvidence(ctx, gone, m.cfg.OutputDir)
}

func (m *Monitor) checkWorkQueueDepth(ctx context.Context, api v1.API) {
	depths, err := heartbeatWorkQueueDepth(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] workqueue_depth: query error: %v", err)
		return
	}
	if depths == nil {
		log.Printf("[validity-gate] [INFO] workqueue_depth: all controllers at 0")
		return
	}
	waits := map[float64]map[string]float64{}
	for _, q := range []float64{0.50, 0.95, 0.99} {
		w, waitErr := heartbeatWorkQueueLatency(ctx, api, q)
		if waitErr != nil {
			log.Printf("[validity-gate] [WARN] workqueue wait p%.0f: query error: %v", q*100, waitErr)
		}
		waits[q] = w
	}
	for _, controller := range sortedKeys(depths) {
		depth := depths[controller]
		status := StatusPass
		if depth > workQueueDepthThreshold {
			status = StatusDegraded
		}
		log.Printf("[validity-gate] [%s] workqueue_depth{%s} depth=%.0f "+
			"wait_p50=%.3fs wait_p95=%.3fs wait_p99=%.3fs threshold=%d",
			status, controller, depth,
			waits[0.50][controller], waits[0.95][controller], waits[0.99][controller],
			workQueueDepthThreshold)
	}
}

func (m *Monitor) checkNodeMemoryUtilization(ctx context.Context, api v1.API) {
	const thresholdFraction = 0.85
	nodes, err := heartbeatNodeMemoryUtilization(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] node_memory_utilization: query error: %v", err)
		return
	}
	if nodes == nil {
		log.Printf("[validity-gate] [UNAVAILABLE] node_memory_utilization: no data")
		return
	}
	for _, instance := range sortedKeys(nodes) {
		fraction := nodes[instance]
		status := StatusPass
		if fraction > thresholdFraction {
			status = StatusFail
		}
		log.Printf("[validity-gate] [%s] node_memory_utilization{instance=%q} value=%.1f%% threshold=%.0f%%",
			status, instance, fraction*100, thresholdFraction*100)
	}
}

func (m *Monitor) checkNodeCPUUtilization(ctx context.Context, api v1.API) {
	const thresholdFraction = 0.80
	nodes, err := heartbeatNodeCPUUtilization(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] node_cpu_utilization: query error: %v", err)
		return
	}
	if nodes == nil {
		log.Printf("[validity-gate] [UNAVAILABLE] node_cpu_utilization: no data")
		return
	}
	for instance, fraction := range nodes {
		status := StatusPass
		if fraction > thresholdFraction {
			status = StatusFail
		}
		log.Printf("[validity-gate] [%s] node_cpu_utilization{instance=%q} value=%.1f%% threshold=%.0f%%", status, instance, fraction*100, thresholdFraction*100)
	}
}

func (m *Monitor) checkWorkloadListLatency(ctx context.Context, api v1.API) {
	const thresholdSecs = 2.0
	val, available, err := heartbeatWorkloadListLatencyP95(ctx, api)
	if err != nil {
		log.Printf("[validity-gate] [WARN] workload_list_latency_p95: query error: %v", err)
		return
	}
	if !available {
		log.Printf("[validity-gate] [INFO] workload_list_latency_p95: no data (no workload list traffic yet)")
		return
	}
	if val > thresholdSecs {
		log.Printf("[validity-gate] [WARN] workload_list_latency_p95 value=%.3fs threshold=%.1fs (degraded)", val, thresholdSecs)
		return
	}
	log.Printf("[validity-gate] [INFO] workload_list_latency_p95 value=%.3fs threshold=%.1fs", val, thresholdSecs)
}

// collectFinalReport runs all range queries over the full run duration and
// assembles the structured JSON report.
// checkWorkloadProgress tracks how long the run has gone with nothing completing. It is a
// heartbeat-only measurement: with a TTL in play the completed pods are gone by the time the
// final queries run, so this cannot be recomputed at the end.
//
// Two conditions suspend counting rather than reset it, because in both the absence of
// completions carries no information: the first workload duration of the run, during which
// nothing can have finished yet, and any heartbeat with no workloads resident.
func (m *Monitor) checkWorkloadProgress(ctx context.Context, api v1.API) {
	if m.progressGateReason != "" {
		return
	}
	nsRegex := workloadNamespaceRegex(m.cfg.ProjectPrefix)

	running, err := heartbeatWorkloadPhase(ctx, api, nsRegex, "Running")
	if err != nil {
		log.Printf("[validity-gate] [WARN] workload_progress: running query error: %v", err)
		return
	}
	pending, err := heartbeatWorkloadPhase(ctx, api, nsRegex, "Pending")
	if err != nil {
		log.Printf("[validity-gate] [WARN] workload_progress: pending query error: %v", err)
		return
	}

	heartbeat, _ := time.ParseDuration(m.cfg.HeartbeatInterval)
	completions, err := heartbeatCompletionsInWindow(ctx, api, nsRegex, heartbeat)
	if err != nil {
		log.Printf("[validity-gate] [WARN] workload_progress: completions query error: %v", err)
		return
	}

	log.Printf("[validity-gate] workload_progress: pods_running=%.0f pods_pending=%.0f completed_in_%s=%.0f",
		running, pending, heartbeat, completions)

	warmup := time.Duration(m.cfg.WorkloadDurationSecs) * time.Second
	if time.Since(m.startTime) < warmup {
		return
	}
	if running+pending == 0 {
		return
	}

	if completions > 0 {
		m.zeroCompletionRun = 0
		return
	}
	m.zeroCompletionRun++
	if m.zeroCompletionRun > m.worstCompletionGap {
		m.worstCompletionGap = m.zeroCompletionRun
	}

	gap := time.Duration(m.zeroCompletionRun) * heartbeat
	threshold := time.Duration(m.cfg.CompletionGapThresholdSecs) * time.Second
	if gap >= threshold {
		log.Printf("[validity-gate] [FAIL] workload_completion_gap: %s with nothing completing "+
			"(threshold %s) while %.0f workloads are resident", gap, threshold, running+pending)
	}
}

func (m *Monitor) collectFinalReport(ctx context.Context, startTime time.Time) (*ValidityGateReport, error) {
	// endTime is the moment the run ended (when ctx was cancelled), not time.Now().
	// Using time.Now() would silently shift the query window forward by however long
	// report collection takes, causing events near the start of the run to fall
	// outside the increase([elapsed]) window and be missed.
	endTime := startTime.Add(time.Since(startTime))
	// Clamp to the actual context deadline when RunDuration was set.
	if deadline, ok := ctx.Deadline(); ok && deadline.Before(endTime) {
		endTime = deadline
	}
	elapsed := endTime.Sub(startTime)
	report := &ValidityGateReport{}

	err := withPrometheusAPI(ctx, m.cfg, m.restConfig, func(api v1.API) error {
		report.ValidityGates = m.buildValidityGates(ctx, api, elapsed, endTime)
		report.HardwareSufficiency = m.buildHardwareSufficiency(ctx, api, elapsed, endTime)
		report.DegradedMetrics = m.buildDegradedMetrics(ctx, api, elapsed, endTime)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Prometheus for final report: %w", err)
	}

	return report, nil
}

func (m *Monitor) buildValidityGates(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) []MetricResult {
	results := []MetricResult{}

	// API server latency. Gated on P99 — P50 sits in the first histogram bucket and cannot
	// distinguish a healthy cluster from a struggling one.
	p50, p95, p99, available, err := finalApiserverLatency(ctx, api, elapsed, endTime)
	if err != nil || !available {
		results = append(results, unavailableOrError("apiserver_request_latency", err))
	} else {
		status := StatusPass
		if p99 > apiserverLatencyP99Threshold {
			status = StatusFail
		}
		results = append(results, MetricResult{
			Metric:    "apiserver_request_latency",
			Value:     fmt.Sprintf("p50=%.3fs p95=%.3fs p99=%.3fs", p50, p95, p99),
			Threshold: fmt.Sprintf("p99 %.1fs", apiserverLatencyP99Threshold),
			Headroom:  fmt.Sprintf("%.1f%%", (apiserverLatencyP99Threshold-p99)/apiserverLatencyP99Threshold*100),
			Status:    status,
		})
	}

	// etcd WAL fsync P99
	val, available, err := finalEtcdWALFsyncP99(ctx, api, elapsed, endTime)
	results = append(results, buildLatencyResult("etcd_wal_fsync_p99", val, 0.1, available, err))

	// etcd proposals pending
	peak, sustained, available, err := finalEtcdProposalsPeakAndSustained(
		ctx, api, elapsed, endTime, m.cfg.EtcdProposalSustainedSecs, m.cfg.EtcdProposalDepthThreshold)
	if err != nil || !available {
		results = append(results, unavailableOrError("etcd_proposals_pending", err))
	} else {
		status := StatusPass
		if sustained {
			status = StatusFail
		}
		results = append(results, MetricResult{
			Metric: "etcd_proposals_pending",
			Value: fmt.Sprintf("peak=%.0f sustained_over_%ds=%v",
				peak, m.cfg.EtcdProposalSustainedSecs, sustained),
			Threshold: fmt.Sprintf(">%d sustained for %ds",
				m.cfg.EtcdProposalDepthThreshold, m.cfg.EtcdProposalSustainedSecs),
			Status: status,
		})
	}

	// etcd DB size ratio
	val, available, err = finalEtcdDBSizeRatioPeak(ctx, api, elapsed, endTime)
	if err != nil || !available {
		results = append(results, unavailableOrError("etcd_db_size_ratio", err))
	} else {
		status := StatusPass
		if val > 0.70 {
			status = StatusFail
		}
		results = append(results, MetricResult{
			Metric:    "etcd_db_size_ratio",
			Value:     fmt.Sprintf("%.1f%%", val*100),
			Threshold: "70%",
			Headroom:  fmt.Sprintf("%.1f%%", (0.70-val)*100),
			Status:    status,
		})
	}

	// Pod restarts
	val, available, err = finalPodRestarts(ctx, api, elapsed, endTime, m.cfg.TargetNamespaces)
	if err != nil || !available {
		results = append(results, unavailableOrError("pod_restarts", err))
	} else {
		status := StatusPass
		var detail []string
		if val > 0 {
			status = StatusFail
			var detailErr error
			detail, detailErr = finalPodRestartDetails(ctx, api, elapsed, endTime, m.cfg.TargetNamespaces)
			if detailErr != nil {
				log.Printf("[validity-gate] [WARN] could not fetch pod restart details: %v", detailErr)
			}
		}
		results = append(results, MetricResult{
			Metric:    "pod_restarts",
			Value:     fmt.Sprintf("%.0f", val),
			Threshold: "0",
			Status:    status,
			Detail:    detail,
		})
	}

	// Node NotReady
	val, available, err = finalNodeNotReadyPeak(ctx, api, elapsed, endTime)
	if err != nil || !available {
		results = append(results, unavailableOrError("node_not_ready_count", err))
	} else {
		status := StatusPass
		if val > 0 {
			status = StatusFail
		}
		results = append(results, MetricResult{
			Metric:    "node_not_ready_count",
			Value:     fmt.Sprintf("%.0f", val),
			Threshold: "0",
			Status:    status,
		})
	}

	// System pods health (all pods in runai/runai-backend — workloads run in project namespaces)
	val, available, err = finalSystemPodsUnhealthyPeak(ctx, api, elapsed, endTime, m.cfg.TargetNamespaces)
	if err != nil || !available {
		results = append(results, unavailableOrError("system_pods_unhealthy", err))
	} else {
		status := StatusPass
		if val > 0 {
			status = StatusFail
		}
		results = append(results, MetricResult{
			Metric:    "system_pods_unhealthy",
			Value:     fmt.Sprintf("%.0f", val),
			Threshold: "0",
			Status:    status,
		})
	}

	// Pod replacements seen across heartbeats. Not derivable from a final range query:
	// a pod that came and went between heartbeats leaves no restart counter behind.
	replStatus := StatusPass
	if len(m.replacements) > 0 {
		replStatus = StatusFail
	}
	replResult := MetricResult{
		Metric:    "pod_replacements",
		Value:     fmt.Sprintf("%d", len(m.replacements)),
		Threshold: "0",
		Status:    replStatus,
	}
	if len(m.replacements) > 0 {
		replResult.Detail = m.replacements
	}
	results = append(results, replResult)

	// Prometheus scrape gaps. Read before the availability gates because those are measured
	// from scraped samples: across a gap there are none, so a clean result there would mean
	// "nothing was observed", not "nothing was wrong".
	gapSecs, gapAvailable, gapErr := finalPrometheusScrapGapSecs(ctx, api, elapsed, endTime)
	scrapeUnreliable := gapErr == nil && gapAvailable && gapSecs > prometheusScrapeGapThreshold

	// Essential service availability. Distinct from prometheus_scrape_gaps: that asks
	// whether Prometheus could see a target, this asks whether the service was usable.
	unavail, err := finalServiceUnavailableSecs(ctx, api, elapsed, endTime)
	if err != nil {
		results = append(results, unavailableOrError("service_unavailable_seconds", err))
	} else {
		for _, svc := range EssentialServices {
			metric := fmt.Sprintf("service_unavailable_seconds{service=%q}", svc.Name)
			secs, scraped := unavail[svc.Name]
			if !scraped {
				results = append(results, MetricResult{
					Metric: metric, Value: "not scraped", Status: StatusUnavailable,
				})
				continue
			}
			// A clean reading cannot be trusted when Prometheus stopped scraping for longer
			// than the gap threshold; a fault found anyway is still a fault.
			if secs == 0 && scrapeUnreliable {
				results = append(results, MetricResult{
					Metric: metric,
					Value: fmt.Sprintf("not evaluated: Prometheus scrape gap of %.0fs "+
						"leaves no samples to judge availability from", gapSecs),
					Status: StatusUnavailable,
				})
				continue
			}
			status := StatusPass
			if secs > 0 {
				status = StatusFail
			}
			results = append(results, MetricResult{
				Metric:    metric,
				Value:     fmt.Sprintf("%.0fs", secs),
				Threshold: "0s",
				Status:    status,
			})
		}
	}

	results = append(results, m.completionGapResult(elapsed))

	// Workloads resident. The completion gate suspends itself when nothing is resident, so
	// without this a run where no workload ever started would report a zero gap and pass.
	// Measured over the counting window only: during warmup the count climbs from zero, and a
	// min taken across that is always zero regardless of what the run did afterwards.
	if m.progressGateReason == "" {
		steadyWindow, windowErr := m.steadyWindow(elapsed)
		nsRegex := workloadNamespaceRegex(m.cfg.ProjectPrefix)
		if windowErr != nil {
			results = append(results, MetricResult{
				Metric: "workload_pods_running", Value: windowErr.Error(), Status: StatusUnavailable,
			})
		} else if levels, err := finalWorkloadLevels(ctx, api, steadyWindow, endTime, nsRegex, "Running"); err != nil {
			results = append(results, unavailableOrError("workload_pods_running", err))
		} else {
			status := StatusPass
			if levels.Max == 0 {
				status = StatusFail
			}
			results = append(results, MetricResult{
				Metric: "workload_pods_running",
				Value: fmt.Sprintf("min=%.0f avg=%.0f max=%.0f%s",
					levels.Min, levels.Avg, levels.Max, m.asWorkloads(levels.Max)),
				Threshold: "max >0",
				Status:    status,
			})
		}
	}

	// Prometheus scrape gaps. Queried above, since the availability gates depend on it.
	if gapErr != nil || !gapAvailable {
		results = append(results, unavailableOrError("prometheus_scrape_gaps", gapErr))
	} else {
		status := StatusPass
		if gapSecs > prometheusScrapeGapThreshold {
			status = StatusFail
		}
		results = append(results, MetricResult{
			Metric:    "prometheus_scrape_gaps",
			Value:     fmt.Sprintf("%.0fs", gapSecs),
			Threshold: fmt.Sprintf("%.0fs", prometheusScrapeGapThreshold),
			Status:    status,
		})
	}

	return results
}

// steadyWindow is how much of the run remains once the ramp is excluded. It is negative when
// the run ended before the ramp finished — an aborted run — and a negative range renders as
// "-2100s", which Prometheus rejects. Returning an error keeps that case a readable report
// entry rather than a PromQL parse failure on the very run being diagnosed.
func (m *Monitor) steadyWindow(elapsed time.Duration) (time.Duration, error) {
	w := elapsed - time.Duration(m.cfg.WorkloadDurationSecs)*time.Second
	if w <= 0 {
		return 0, fmt.Errorf("run ended during the ramp: %s elapsed, ramp is %ds",
			elapsed.Round(time.Second), m.cfg.WorkloadDurationSecs)
	}
	return w, nil
}

// asWorkloads renders a pod count as the workload count it implies, for distributed workloads
// where the two differ.
func (m *Monitor) asWorkloads(pods float64) string {
	if m.cfg.PodsPerWorkload <= 1 {
		return ""
	}
	return fmt.Sprintf(" (%.0f workloads at %d pods each)", pods/float64(m.cfg.PodsPerWorkload),
		m.cfg.PodsPerWorkload)
}

// completionGapResult reports the longest observed stretch with nothing completing, against the
// threshold it was judged by. The two belong together: a healthy run reports a non-zero gap, so
// the bare number reads as a problem on its own.
func (m *Monitor) completionGapResult(elapsed time.Duration) MetricResult {
	result := MetricResult{Metric: "workload_completion_gap"}
	if m.progressGateReason != "" {
		result.Value = m.progressGateReason
		result.Status = StatusUnavailable
		return result
	}

	heartbeat, _ := time.ParseDuration(m.cfg.HeartbeatInterval)
	threshold := time.Duration(m.cfg.CompletionGapThresholdSecs) * time.Second
	result.Threshold = threshold.String()

	// Counting only begins one workload duration in, so a short run can end before the gate
	// has had room to reach its threshold. Reporting PASS there would claim an all-clear the
	// run never had the length to establish.
	counting := elapsed - time.Duration(m.cfg.WorkloadDurationSecs)*time.Second
	if counting < threshold {
		result.Value = fmt.Sprintf("run too short to evaluate: %s of counting window after the "+
			"%ds warmup, threshold is %s", counting.Round(time.Second), m.cfg.WorkloadDurationSecs, threshold)
		result.Status = StatusUnavailable
		return result
	}

	worst := time.Duration(m.worstCompletionGap) * heartbeat
	result.Value = fmt.Sprintf("%s (%d heartbeats)", worst, m.worstCompletionGap)
	result.Status = StatusPass
	if worst >= threshold {
		result.Status = StatusFail
	}
	return result
}

func (m *Monitor) buildDegradedMetrics(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) []MetricResult {
	results := []MetricResult{}

	// Workloads waiting to be scheduled, and the rate they actually arrived at. Reported
	// without a threshold: what they should be is a property of the run parameters, which live
	// in the orchestrator, not here. They are what the gates above are read against when one
	// of them fails.
	// Skipped entirely when the progress gate is off: the completion gate already reports why.
	if steadyWindow, windowErr := m.steadyWindow(elapsed); m.progressGateReason == "" && windowErr != nil {
		for _, name := range []string{"workload_pods_pending", "workload_submission_rate"} {
			results = append(results, MetricResult{
				Metric: name, Value: windowErr.Error(), Status: StatusUnavailable,
			})
		}
	} else if m.progressGateReason == "" {
		nsRegex := workloadNamespaceRegex(m.cfg.ProjectPrefix)
		if levels, err := finalWorkloadLevels(ctx, api, steadyWindow, endTime, nsRegex, "Pending"); err != nil {
			results = append(results, unavailableOrError("workload_pods_pending", err))
		} else {
			results = append(results, MetricResult{
				Metric: "workload_pods_pending",
				Value: fmt.Sprintf("min=%.0f avg=%.0f max=%.0f%s",
					levels.Min, levels.Avg, levels.Max, m.asWorkloads(levels.Max)),
				Status: StatusPass,
			})
		}

		rate, available, err := finalSubmissionRatePerMin(ctx, api, steadyWindow, endTime)
		if err == nil && !available {
			// The selector pins a job and two exact paths, so nothing matching may mean the
			// service no longer exposes them rather than that nothing was submitted.
			if present, perr := submissionSeriesPresent(ctx, api, steadyWindow, endTime); perr == nil && !present {
				err = fmt.Errorf("no series matched %s — the job label or paths may have changed",
					submissionSelector)
			}
		}
		if err != nil || !available {
			results = append(results, unavailableOrError("workload_submission_rate", err))
		} else {
			value := fmt.Sprintf("%.1f/min", rate)
			if m.cfg.SubmissionRatePerMin > 0 {
				value += fmt.Sprintf(" (offered %.1f/min)", m.cfg.SubmissionRatePerMin)
			}
			results = append(results, MetricResult{
				Metric: "workload_submission_rate",
				Value:  value,
				Status: StatusPass,
			})
		}
	}

	// Workload submission latency
	p50, p99, available, err := finalWorkloadSubmissionLatency(ctx, api, elapsed, endTime)
	if err != nil || !available {
		results = append(results, unavailableOrError("workload_submission_latency", err))
	} else {
		status := StatusPass
		if p99 > m.cfg.SubmissionLatencyP99Threshold {
			status = StatusDegraded
		}
		results = append(results, MetricResult{
			Metric:    "workload_submission_latency",
			Value:     fmt.Sprintf("p50=%.3fs p99=%.3fs", p50, p99),
			Threshold: fmt.Sprintf("p99 %.1fs", m.cfg.SubmissionLatencyP99Threshold),
			Status:    status,
		})
	}

	// Scheduler allocate cycle latency, per scheduler (one per nodepool)
	cycles, err := finalSchedulerCycleLatencyByScheduler(ctx, api, elapsed, endTime)
	if err != nil || len(cycles) == 0 {
		results = append(results, unavailableOrError("scheduler_allocate_cycle_latency_ms", err))
	} else {
		for _, name := range sortedKeys(cycles) {
			c := cycles[name]
			status := StatusPass
			if c.Peak > m.cfg.SchedulerCyclePeakThresholdMs {
				status = StatusDegraded
			}
			results = append(results, MetricResult{
				Metric:    fmt.Sprintf("scheduler_allocate_cycle_latency_ms{scheduler=%q}", name),
				Value:     fmt.Sprintf("avg=%.1fms peak=%.1fms", c.Avg, c.Peak),
				Threshold: fmt.Sprintf("peak %.0fms", m.cfg.SchedulerCyclePeakThresholdMs),
				Status:    status,
			})
		}
	}

	// Reconciliation latency P99 per controller
	reconcile, err := finalReconciliationLatencyP99ByController(ctx, api, elapsed, endTime)
	if err != nil {
		results = append(results, unavailableOrError("reconciliation_latency_p99", err))
	} else if reconcile == nil {
		results = append(results, unavailableOrError("reconciliation_latency_p99", nil))
	} else {
		for _, controller := range sortedKeys(reconcile) {
			latency := reconcile[controller]
			// Controllers that never reconciled produce NaN from histogram_quantile.
			if math.IsNaN(latency) {
				results = append(results, MetricResult{
					Metric: fmt.Sprintf("reconciliation_latency_p99{controller=%q}", controller),
					Value:  "no reconciles",
					Status: StatusUnavailable,
				})
				continue
			}
			status := StatusPass
			if latency > m.cfg.ReconciliationLatencyP99Threshold {
				status = StatusDegraded
			}
			results = append(results, MetricResult{
				Metric:    fmt.Sprintf("reconciliation_latency_p99{controller=%q}", controller),
				Value:     fmt.Sprintf("%.3fs", latency),
				Threshold: fmt.Sprintf("%.1fs", m.cfg.ReconciliationLatencyP99Threshold),
				Status:    status,
			})
		}
	}

	// Workload list API P95 latency — UI responsiveness gate.
	// Uses path="/api/v1/workloads" which is distinct from submission paths.
	// A breach sets UIDegraded in the report but does not fail the run.
	const workloadListThreshold = 2.0
	listP50, listP95, available, err := finalWorkloadListLatency(ctx, api, elapsed, endTime)
	if err != nil || !available {
		results = append(results, unavailableOrError("workload_list_latency", err))
	} else {
		status := StatusPass
		if listP95 > workloadListThreshold {
			status = StatusDegraded
		}
		headroom := (workloadListThreshold - listP95) / workloadListThreshold * 100
		results = append(results, MetricResult{
			Metric:    "workload_list_latency",
			Value:     fmt.Sprintf("p50=%.3fs p95=%.3fs", listP50, listP95),
			Threshold: fmt.Sprintf("p95 %.1fs", workloadListThreshold),
			Headroom:  fmt.Sprintf("%.1f%%", headroom),
			Status:    status,
		})
	}

	// Work queue depth per controller
	depths, err := finalWorkQueueDepthByController(ctx, api, elapsed, endTime,
		workQueueDepthThreshold, workQueueSustainedSecs)
	if err != nil {
		results = append(results, unavailableOrError("workqueue_depth_peak", err))
	} else if depths == nil {
		results = append(results, unavailableOrError("workqueue_depth_peak", nil))
	} else if len(depths) == 0 {
		results = append(results, MetricResult{
			Metric: "workqueue_depth_peak",
			Value:  "all controllers: 0",
			Status: StatusPass,
		})
	} else {
		// A non-empty queue is normal — that is what a queue is for. Degraded means the
		// backlog persisted, not that it briefly spiked.
		for _, controller := range sortedKeys(depths) {
			stat := depths[controller]
			status := StatusPass
			if stat.Sustained {
				status = StatusDegraded
			}
			results = append(results, MetricResult{
				Metric: fmt.Sprintf("workqueue_depth{%s}", controller),
				Value: fmt.Sprintf("peak=%.0f sustained_over_%ds=%v wait_p50=%.3fs wait_p95=%.3fs wait_p99=%.3fs",
					stat.Peak, workQueueSustainedSecs, stat.Sustained,
					stat.WaitP50, stat.WaitP95, stat.WaitP99),
				Threshold: fmt.Sprintf(">%d for %ds", workQueueDepthThreshold, workQueueSustainedSecs),
				Status:    status,
			})
		}
	}

	return results
}

func (m *Monitor) buildHardwareSufficiency(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) []MetricResult {
	const nodeCPUThreshold = 0.80

	results := []MetricResult{}

	// Per-node peak CPU utilization (host-level, all workloads combined).
	// Saturation on the k8s control-plane node invalidates etcd/apiserver metrics;
	// saturation on worker nodes invalidates run:ai backend and scheduler metrics.
	nodePeaks, err := finalNodeCPUUtilizationPeak(ctx, api, elapsed, endTime)
	if err != nil {
		results = append(results, unavailableOrError("node_cpu_utilization_peak", err))
	} else if nodePeaks == nil {
		results = append(results, unavailableOrError("node_cpu_utilization_peak", nil))
	} else {
		for _, instance := range sortedKeys(nodePeaks) {
			fraction := nodePeaks[instance]
			status := StatusPass
			if fraction > nodeCPUThreshold {
				status = StatusFail
			}
			results = append(results, MetricResult{
				Metric:    fmt.Sprintf("node_cpu_utilization_peak{instance=%q}", instance),
				Value:     fmt.Sprintf("%.1f%%", fraction*100),
				Threshold: fmt.Sprintf("%.0f%%", nodeCPUThreshold*100),
				Headroom:  fmt.Sprintf("%.1f%%", (nodeCPUThreshold-fraction)*100),
				Status:    status,
			})
		}
	}

	// A node under memory pressure evicts pods and stalls controllers, which invalidates
	// the run just as CPU saturation does.
	const nodeMemThreshold = 0.85
	memPeaks, err := finalNodeMemoryUtilizationPeak(ctx, api, elapsed, endTime)
	if err != nil {
		results = append(results, unavailableOrError("node_memory_utilization_peak", err))
	} else if memPeaks == nil {
		results = append(results, unavailableOrError("node_memory_utilization_peak", nil))
	} else {
		for _, instance := range sortedKeys(memPeaks) {
			fraction := memPeaks[instance]
			status := StatusPass
			if fraction > nodeMemThreshold {
				status = StatusFail
			}
			results = append(results, MetricResult{
				Metric:    fmt.Sprintf("node_memory_utilization_peak{instance=%q}", instance),
				Value:     fmt.Sprintf("%.1f%%", fraction*100),
				Threshold: fmt.Sprintf("%.0f%%", nodeMemThreshold*100),
				Headroom:  fmt.Sprintf("%.1f%%", (nodeMemThreshold-fraction)*100),
				Status:    status,
			})
		}
	}

	return results
}

// collectCrashLogs fetches previous-container logs for each entry in detail
// (format: "namespace/pod/container: N restarts") and writes them to
// outputDir/crash-logs/<namespace>-<pod>-<container>.log.
// Errors are logged but never propagate — log collection is best-effort.
// collectReplacementEvidence captures why a pod was replaced. The pod's own logs are
// already unrecoverable at this point, so it gathers what still exists: namespace events
// (which name the actor — Evicted, Preempted, NodeNotReady — and expire in ~1h) and the
// conditions of every node, to catch MemoryPressure or DiskPressure driving the eviction.
func (m *Monitor) collectReplacementEvidence(ctx context.Context, replaced []string, outputDir string) {
	client, err := kubernetes.NewForConfig(m.restConfig)
	if err != nil {
		log.Printf("[validity-gate] [WARN] replacement evidence: could not create kube client: %v", err)
		return
	}
	dir := filepath.Join(outputDir, "replacements")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[validity-gate] [WARN] replacement evidence: could not create %s: %v", dir, err)
		return
	}

	stamp := time.Now().UTC().Format("20060102T150405Z")
	var b strings.Builder
	fmt.Fprintf(&b, "detected at %s\nreplaced pods:\n", stamp)
	for _, p := range replaced {
		fmt.Fprintf(&b, "  %s\n", p)
	}

	seen := map[string]bool{}
	for _, p := range replaced {
		ns, _, ok := strings.Cut(p, "/")
		if !ok || seen[ns] {
			continue
		}
		seen[ns] = true
		events, err := client.CoreV1().Events(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(&b, "\nevents in %s: ERROR %v\n", ns, err)
			continue
		}
		fmt.Fprintf(&b, "\nevents in %s (%d):\n", ns, len(events.Items))
		for _, e := range events.Items {
			fmt.Fprintf(&b, "  %s %s/%s %s: %s\n",
				e.LastTimestamp.UTC().Format(time.RFC3339), e.Type, e.Reason,
				e.InvolvedObject.Name, e.Message)
		}
	}

	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(&b, "\nnode conditions: ERROR %v\n", err)
	} else {
		fmt.Fprintf(&b, "\nnode conditions (non-Ready or pressure only):\n")
		for _, n := range nodes.Items {
			for _, c := range n.Status.Conditions {
				bad := (c.Type == corev1.NodeReady && c.Status != corev1.ConditionTrue) ||
					(c.Type != corev1.NodeReady && c.Status == corev1.ConditionTrue)
				if bad {
					fmt.Fprintf(&b, "  %s %s=%s %s\n", n.Name, c.Type, c.Status, c.Message)
				}
			}
		}
	}

	path := filepath.Join(dir, fmt.Sprintf("replacement-%s.txt", stamp))
	if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
		log.Printf("[validity-gate] [WARN] replacement evidence: could not write %s: %v", path, err)
		return
	}
	log.Printf("[validity-gate] replacement evidence written to %s", path)
}

// freshCrashes returns the detail entries whose restart count has advanced since the last
// capture. A container missing from totals is treated as fresh — capturing twice beats
// dropping a crash.
func (m *Monitor) freshCrashes(detail []string, totals map[string]float64) []string {
	if m.capturedRestarts == nil {
		m.capturedRestarts = map[string]float64{}
	}
	fresh := make([]string, 0, len(detail))
	for _, d := range detail {
		key, _, ok := strings.Cut(d, ": ")
		if !ok {
			continue
		}
		n, known := totals[key]
		if known && n <= m.capturedRestarts[key] {
			continue
		}
		if known {
			m.capturedRestarts[key] = n
		}
		fresh = append(fresh, d)
	}
	return fresh
}

func (m *Monitor) collectCrashLogs(ctx context.Context, detail []string, outputDir string) {
	if len(detail) == 0 {
		return
	}

	client, err := kubernetes.NewForConfig(m.restConfig)
	if err != nil {
		log.Printf("[validity-gate] [WARN] crash logs: could not create kube client: %v", err)
		return
	}

	dir := filepath.Join(outputDir, "crash-logs")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[validity-gate] [WARN] crash logs: could not create output dir %s: %v", dir, err)
		return
	}

	// Stamped, because a container that crashes repeatedly would otherwise overwrite its
	// own earlier logs and only the last crash would survive.
	stamp := time.Now().UTC().Format("20060102T150405Z")

	prev := true
	for _, entry := range detail {
		// entry format: "namespace/pod/container: N restarts"
		parts := strings.SplitN(entry, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		npc := strings.SplitN(parts[0], "/", 3)
		if len(npc) != 3 {
			continue
		}
		ns, pod, container := npc[0], npc[1], npc[2]

		req := client.CoreV1().Pods(ns).GetLogs(pod, &corev1.PodLogOptions{
			Container: container,
			Previous:  prev,
		})
		stream, err := req.Stream(ctx)
		if err != nil {
			log.Printf("[validity-gate] [WARN] crash logs: could not stream logs for %s/%s/%s: %v", ns, pod, container, err)
			continue
		}

		filename := fmt.Sprintf("%s-%s-%s-%s.log", ns, pod, container, stamp)
		path := filepath.Join(dir, filename)
		f, err := os.Create(path)
		if err != nil {
			stream.Close()
			log.Printf("[validity-gate] [WARN] crash logs: could not create file %s: %v", path, err)
			continue
		}

		if _, err := io.Copy(f, stream); err != nil {
			log.Printf("[validity-gate] [WARN] crash logs: error writing %s: %v", path, err)
		}
		f.Close()
		stream.Close()
		log.Printf("[validity-gate] crash log saved: %s", path)
	}
}

// --- helpers ---

func buildLatencyResult(metric string, val, thresholdSecs float64, available bool, err error) MetricResult {
	if err != nil || !available {
		return unavailableOrError(metric, err)
	}
	status := StatusPass
	if val > thresholdSecs {
		status = StatusFail
	}
	headroom := (thresholdSecs - val) / thresholdSecs * 100
	return MetricResult{
		Metric:    metric,
		Value:     fmt.Sprintf("%.3fs", val),
		Threshold: fmt.Sprintf("%.3fs", thresholdSecs),
		Headroom:  fmt.Sprintf("%.1f%%", headroom),
		Status:    status,
	}
}

func unavailableOrError(metric string, err error) MetricResult {
	val := "no data in Prometheus"
	if err != nil {
		val = err.Error()
	}
	return MetricResult{
		Metric: metric,
		Value:  val,
		Status: StatusUnavailable,
	}
}

// sortedKeys gives deterministic report ordering; Go map iteration is random, which makes
// two runs of the same cluster produce non-comparable reports.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
