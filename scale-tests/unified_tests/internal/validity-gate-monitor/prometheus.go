/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validity_gate_monitor

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/sampler_actions/utils/kubernetes"
	"github.com/phayes/freeport"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"
)

// CriticalMetricSources lists the Prometheus job names whose data feeds validity gate panels.
// Used in both the preflight health check and the final scrape-gap calculation so both
// operate on exactly the same set of sources.
//
//   - apiserver          → panel 1  (apiserver request latency)
//   - etcd-metrics-proxy → panels 2-4 (etcd WAL fsync / proposals / DB size)
//   - kube-state-metrics → panels 5-7 (pod restarts / node not-ready / system pods unhealthy)
//   - runai-scheduler-default → panel 9 (scheduler allocate cycle latency)
//   - runai-node-exporter → panel 12a (per-node CPU utilization)
//
// Exported so the env-readiness checker gates on exactly this set rather than defining a
// second, drifting list. It is deliberately an allowlist and not "every scrape target":
// some targets (e.g. inference-workload-controller without kubeflow/knative installed) are
// legitimately absent and must not block a run.
var CriticalMetricSources = []string{
	"apiserver",
	"etcd-metrics-proxy",
	"kube-state-metrics",
	"runai-scheduler-default",
	"runai-node-exporter",
}

// withPrometheusAPI establishes a short-lived port-forward to the configured Prometheus pod,
// calls fn with the ready API client, then tears the connection down.
// The port-forward is held open only for the duration of fn — suitable for
// heartbeat checks that run every few minutes.
func withPrometheusAPI(ctx context.Context, cfg *MonitorConfig, restConfig *rest.Config, fn func(v1.API) error) error {
	return WithPrometheusAPI(ctx, cfg.PrometheusNamespace, cfg.PrometheusPodName, cfg.PrometheusPort, restConfig, fn)
}

// WithPrometheusAPI is the port-forwarding helper behind withPrometheusAPI, with the
// Prometheus location passed explicitly. Exported so callers outside the monitor — notably
// the env-readiness checker — can query Prometheus without depending on MonitorConfig.
func WithPrometheusAPI(
	ctx context.Context,
	namespace, podName string,
	remotePort int,
	restConfig *rest.Config,
	fn func(v1.API) error,
) error {
	port, err := freeport.GetFreePort()
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}

	pfCtx, pfCancel := context.WithCancel(ctx)
	defer pfCancel()

	readyCh := make(chan struct{})
	errCh := make(chan error, 1)

	pf := kubernetes.NewPortForwarder(restConfig)
	go func() {
		err := pf.PortForward(
			namespace,
			podName,
			port,
			remotePort,
			pfCtx.Done(),
			readyCh,
			genericclioptions.IOStreams{ErrOut: os.Stderr},
		)
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case <-readyCh:
	case err := <-errCh:
		return fmt.Errorf("port-forward to %s/%s failed: %w", namespace, podName, err)
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("timed out waiting for port-forward to %s/%s", namespace, podName)
	}

	client, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://localhost:%d", port),
	})
	if err != nil {
		return fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	return fn(v1.NewAPI(client))
}

// --- heartbeat (instant) queries ---
// Each function returns (value, available, error).
// "available" is false when the metric has no data in Prometheus (e.g. etcd not scraped).

func queryInstantScalar(ctx context.Context, api v1.API, query string) (float64, bool, error) {
	return queryInstantScalarAt(ctx, api, query, time.Now())
}

// queryInstantScalarAt evaluates a scalar instant query at a specific time.
// Final-report queries must use the run's actual end time, not time.Now() —
// otherwise the query window shifts forward and events near the run start fall out.
func queryInstantScalarAt(ctx context.Context, api v1.API, query string, t time.Time) (float64, bool, error) {
	result, _, _, err := api.Query(ctx, query, t)
	if err != nil {
		return 0, false, fmt.Errorf("query %q failed: %w", query, err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return 0, false, nil
	}
	return float64(vec[0].Value), true, nil
}

func queryInstantMax(ctx context.Context, api v1.API, query string) (float64, bool, error) {
	result, _, _, err := api.Query(ctx, fmt.Sprintf("max(%s)", query), time.Now())
	if err != nil {
		return 0, false, fmt.Errorf("query %q failed: %w", query, err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return 0, false, nil
	}
	return float64(vec[0].Value), true, nil
}

// heartbeatApiserverLatencyP50 returns current P50 LIST/GET latency (seconds) over the last 5m.
// P50 rises only under genuine API server saturation; P99 is skewed by inherently large LIST
// operations the test itself generates.
func heartbeatApiserverLatencyP50(ctx context.Context, api v1.API) (float64, bool, error) {
	q := `histogram_quantile(0.50, sum(rate(apiserver_request_duration_seconds_bucket{verb=~"LIST|GET"}[5m])) by (le))`
	return queryInstantScalar(ctx, api, q)
}

// heartbeatEtcdWALFsyncP99 returns current P99 etcd WAL fsync latency (seconds) over the last 5m.
func heartbeatEtcdWALFsyncP99(ctx context.Context, api v1.API) (float64, bool, error) {
	q := `histogram_quantile(0.99, sum(rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m])) by (le))`
	return queryInstantScalar(ctx, api, q)
}

// heartbeatEtcdProposalsPending returns the current max etcd proposals pending value.
func heartbeatEtcdProposalsPending(ctx context.Context, api v1.API) (float64, bool, error) {
	return queryInstantMax(ctx, api, "etcd_server_proposals_pending")
}

// heartbeatEtcdDBSizeRatio returns current etcd DB usage as a fraction of quota (0–1).
func heartbeatEtcdDBSizeRatio(ctx context.Context, api v1.API) (float64, bool, error) {
	q := `max(etcd_mvcc_db_total_size_in_bytes / etcd_server_quota_backend_bytes)`
	return queryInstantScalar(ctx, api, q)
}

// heartbeatNodeNotReadyCount returns the number of nodes currently not Ready.
func heartbeatNodeNotReadyCount(ctx context.Context, api v1.API) (float64, bool, error) {
	q := `count(kube_node_status_condition{condition="Ready",status="false"} == 1) or vector(0)`
	return queryInstantScalar(ctx, api, q)
}

// heartbeatPodRestartsDelta returns the total restart count increase in target namespaces
// over the last heartbeat window (5m).
func heartbeatPodRestartsDelta(ctx context.Context, api v1.API, namespaces []string) (float64, bool, error) {
	nsRegex := namespacesRegex(namespaces)
	q := fmt.Sprintf(`sum(increase(kube_pod_container_status_restarts_total{namespace=~"%s"}[5m]))`, nsRegex)
	return queryInstantScalar(ctx, api, q)
}

// heartbeatSystemPodsUnhealthy returns the number of system pods in the target namespaces
// that are not in Running or Succeeded phase. Workloads always run in project namespaces,
// so every non-healthy pod here is a genuine system component failure.
func heartbeatSystemPodsUnhealthy(ctx context.Context, api v1.API, namespaces []string) (float64, bool, error) {
	nsRegex := namespacesRegex(namespaces)
	q := fmt.Sprintf(
		`count(kube_pod_status_phase{namespace=~"%s", phase!~"Running|Succeeded"} == 1) or vector(0)`,
		nsRegex,
	)
	return queryInstantScalar(ctx, api, q)
}

// heartbeatSchedulerCycleLatency returns the latest scheduler allocate cycle latency in ms
// (avg and peak over the last 5m). Returns available=false if no scheduler data yet.
func heartbeatSchedulerCycleLatency(ctx context.Context, api v1.API) (avg, peak float64, available bool, err error) {
	avgQ := `avg(runai_action_scheduling_latency_milliseconds{action="allocate"})`
	peakQ := `max(runai_action_scheduling_latency_milliseconds{action="allocate"})`
	avg, available, err = queryInstantScalar(ctx, api, avgQ)
	if err != nil || !available {
		return
	}
	peak, available, err = queryInstantScalar(ctx, api, peakQ)
	return
}

// heartbeatWorkloadSubmissionLatencyP99 returns the current P99 workload submission latency
// over the last 5m from the run:ai backend HTTP metrics.
func heartbeatWorkloadSubmissionLatencyP99(ctx context.Context, api v1.API) (float64, bool, error) {
	q := `histogram_quantile(0.99, sum(rate(runai_http_request_duration_seconds_bucket{method="POST",path=~"/api/v[12]/workloads/(trainings|workspaces|inferences|distributed|distributed-inferences|unified)"}[5m])) by (le))`
	return queryInstantScalar(ctx, api, q)
}

// heartbeatReconciliationLatencyP99 returns the current P99 reconciliation latency per
// controller (instant query over 5m rate window). Returns nil map if no data.
func heartbeatReconciliationLatencyP99(ctx context.Context, api v1.API) (map[string]float64, error) {
	q := `histogram_quantile(0.99, sum(rate(controller_runtime_reconcile_time_seconds_bucket[5m])) by (le, controller))`
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("reconciliation latency query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		out[string(s.Metric["controller"])] = float64(s.Value)
	}
	return out, nil
}

// heartbeatWorkQueueDepth returns the current work queue depth per controller.
// Only returns controllers with depth > 0.
// heartbeatWorkQueueLatency returns the current queue wait at the given quantile per
// controller, for queues that are actually waiting.
func heartbeatWorkQueueLatency(ctx context.Context, api v1.API, quantile float64) (map[string]float64, error) {
	q := fmt.Sprintf(
		`histogram_quantile(%.2f, sum by (name, namespace, pod, le) (rate(workqueue_queue_duration_seconds_bucket[5m]))) > 0`,
		quantile)
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("workqueue latency query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		if math.IsNaN(float64(s.Value)) {
			continue
		}
		out[workQueueKey(s.Metric)] = float64(s.Value)
	}
	return out, nil
}

func heartbeatWorkQueueDepth(ctx context.Context, api v1.API) (map[string]float64, error) {
	result, _, _, err := api.Query(ctx, `workqueue_depth > 0`, time.Now())
	if err != nil {
		return nil, fmt.Errorf("workqueue depth query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		out[workQueueKey(s.Metric)] = float64(s.Value)
	}
	return out, nil
}

// heartbeatNodeCPUUtilization returns current CPU utilization per node (1 - idle fraction).
func heartbeatNodeCPUUtilization(ctx context.Context, api v1.API) (map[string]float64, error) {
	q := `1 - avg by(instance)(rate(node_cpu_seconds_total{mode="idle"}[2m]))`
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("node CPU query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		out[string(s.Metric["instance"])] = float64(s.Value)
	}
	return out, nil
}

// workloadNamespaceRegex matches the project namespaces workload pods land in.
func workloadNamespaceRegex(projectPrefix string) string {
	return fmt.Sprintf("runai-%s-.*", projectPrefix)
}

func workloadPhaseCount(nsRegex, phase string) string {
	return fmt.Sprintf(`count(kube_pod_status_phase{namespace=~%q, phase=%q} == 1) or vector(0)`,
		nsRegex, phase)
}

// WorkloadLevels is the workload population over the run. min matters most: a mid-run
// collapse is invisible in the peak and barely moves the average.
type WorkloadLevels struct {
	Min, Avg, Max float64
}

func finalWorkloadLevels(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time, nsRegex, phase string) (WorkloadLevels, error) {
	dur := promDuration(elapsed)
	base := workloadPhaseCount(nsRegex, phase)
	var out WorkloadLevels
	for _, s := range []struct {
		q   string
		dst *float64
	}{
		{fmt.Sprintf(`min_over_time((%s)[%s:30s])`, base, dur), &out.Min},
		{fmt.Sprintf(`avg_over_time((%s)[%s:30s])`, base, dur), &out.Avg},
		{fmt.Sprintf(`max_over_time((%s)[%s:30s])`, base, dur), &out.Max},
	} {
		v, _, err := queryInstantScalarAt(ctx, api, s.q, endTime)
		if err != nil {
			return out, fmt.Errorf("workload %s levels: %w", phase, err)
		}
		*s.dst = v
	}
	return out, nil
}

func heartbeatWorkloadPhase(ctx context.Context, api v1.API, nsRegex, phase string) (float64, error) {
	v, _, err := queryInstantScalar(ctx, api, workloadPhaseCount(nsRegex, phase))
	return v, err
}

// heartbeatCompletionsInWindow counts pods that finished inside the last window and are
// still present to be seen. Reading a count over a short window rather than a running
// total keeps this correct whether or not completed workloads are being deleted — the only
// requirement is that the TTL outlives one heartbeat.
func heartbeatCompletionsInWindow(ctx context.Context, api v1.API, nsRegex string, window time.Duration) (float64, error) {
	q := fmt.Sprintf(`count(kube_pod_completion_time{namespace=~%q} > (time() - %d)) or vector(0)`,
		nsRegex, int(window.Seconds()))
	v, _, err := queryInstantScalar(ctx, api, q)
	return v, err
}

// submissionSelector matches the exact creation endpoints the submitter posts to, one per
// workload type. Non-2xx excluded — a rejection is not offered load.
const submissionSelector = `runai_http_request_duration_seconds_count{` +
	`job="runai-backend-workloads",method="POST",` +
	`path=~"/api/v1/workloads/(trainings|distributed)",status_code=~"2.."}`

func submissionRatePerMin(ctx context.Context, api v1.API) (float64, bool, error) {
	return queryInstantScalar(ctx, api,
		fmt.Sprintf(`sum(rate(%s[5m])) * 60`, submissionSelector))
}

// Averaged over the steady window, not the whole run: the ramp submits faster than R by
// design, so including it reports a rate the run never held.
func finalSubmissionRatePerMin(ctx context.Context, api v1.API, steady time.Duration, endTime time.Time) (float64, bool, error) {
	q := fmt.Sprintf(`avg_over_time((sum(rate(%s[5m])) * 60)[%s:1m])`,
		submissionSelector, promDuration(steady))
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// submissionSeriesPresent separates "no submissions" from "the selector stopped matching",
// which an empty rate query alone cannot.
func submissionSeriesPresent(ctx context.Context, api v1.API, window time.Duration, endTime time.Time) (bool, error) {
	q := fmt.Sprintf(`count(count_over_time(%s[%s])) or vector(0)`,
		submissionSelector, promDuration(window))
	v, _, err := queryInstantScalarAt(ctx, api, q, endTime)
	return v > 0, err
}

// heartbeatNodeMemoryUtilization mirrors the final node memory gate.
func heartbeatNodeMemoryUtilization(ctx context.Context, api v1.API) (map[string]float64, error) {
	q := `1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)`
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("node memory query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		out[string(s.Metric["instance"])] = float64(s.Value)
	}
	return out, nil
}

// heartbeatMetricSourcesDown returns the subset of CriticalMetricSources that are currently
// down (up==0) or absent (no scrape target). Called every heartbeat so any source failure
// is logged within one heartbeat interval.
func heartbeatMetricSourcesDown(ctx context.Context, api v1.API) ([]string, error) {
	var down []string
	for _, job := range CriticalMetricSources {
		q := fmt.Sprintf(`up{job=%q}`, job)
		result, _, _, err := api.Query(ctx, q, time.Now())
		if err != nil {
			return nil, fmt.Errorf("up query for %q failed: %w", job, err)
		}
		vec, ok := result.(model.Vector)
		if !ok || len(vec) == 0 {
			// No series at all — job is not scraped.
			down = append(down, job+" (no scrape target)")
			continue
		}
		for _, s := range vec {
			if float64(s.Value) != 1 {
				down = append(down, job)
				break
			}
		}
	}
	return down, nil
}

// --- final report (range) queries ---
// elapsed is the total duration of the completed test run.

// finalApiserverLatency returns worst 5m-window P50/P95/P99 apiserver latency over the run.
// WATCH and CONNECT are excluded: they are long-polls (60s and ~10s) that swamp any
// quantile. All remaining verbs are included — the previous LIST|GET filter covered only
// ~7% of traffic and hid PUT/PATCH, which are the ones that reach etcd.
func finalApiserverLatency(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (p50, p95, p99 float64, available bool, err error) {
	q := func(quantile string) string {
		return fmt.Sprintf(
			`max_over_time(histogram_quantile(%s, sum(rate(apiserver_request_duration_seconds_bucket{verb!~"WATCH|CONNECT"}[5m])) by (le))[%s:5m])`,
			quantile, promDuration(elapsed),
		)
	}
	if p50, available, err = queryInstantScalarAt(ctx, api, q("0.50"), endTime); err != nil || !available {
		return
	}
	if p95, available, err = queryInstantScalarAt(ctx, api, q("0.95"), endTime); err != nil || !available {
		return
	}
	p99, available, err = queryInstantScalarAt(ctx, api, q("0.99"), endTime)
	return
}

// heartbeatPodUIDs returns pod UID → "namespace/pod" for the target namespaces.
// A pod that is deleted and recreated by its owner gets a new UID and a restartCount of 0,
// so pod_restarts cannot see it — comparing UIDs between heartbeats is what detects it.
func heartbeatPodUIDs(ctx context.Context, api v1.API, namespaces []string) (map[string]string, error) {
	q := fmt.Sprintf(`kube_pod_info{namespace=~"%s"}`, namespacesRegex(namespaces))
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("pod uid query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, nil
	}
	out := make(map[string]string, len(vec))
	for _, s := range vec {
		uid := string(s.Metric["uid"])
		if uid == "" {
			continue
		}
		out[uid] = fmt.Sprintf("%s/%s", s.Metric["namespace"], s.Metric["pod"])
	}
	return out, nil
}

// EssentialService is one whose outage invalidates the run rather than degrading it.
type EssentialService struct {
	Name string
	// Down is non-empty exactly while the service is unusable.
	Down string
	// Present detects that the service is scraped at all, so a renamed job surfaces as
	// UNAVAILABLE rather than silently reporting zero downtime.
	Present string
}

// EssentialServices covers the cluster control plane plus the control-plane services the
// submitter calls, where an outage is indistinguishable from a scale limit.
// etcd uses its own has-leader metric, not up{job="etcd-metrics-proxy"}: the proxy only
// forwards metrics, so its scrape health says nothing about whether etcd serves reads and
// writes. Without a leader etcd cannot serve writes; if etcd is gone the series disappears,
// hence the absent() arm. job="kube-etcd" is deliberately excluded — it is the built-in
// target that never works, which is why the proxy exists.
// Every Down expression carries an absent() arm. `up{...} == 0` is a filter: it yields nothing
// when the series is missing, so without absent() an outage severe enough to also stop the
// scrape — service discovery dropping the target, or Prometheus itself falling behind — counts
// as zero downtime and the gate passes. Observed on an xl run where the apiserver was
// OOM-killed and this gate still reported 0s.
//
// etcd needs a second arm as well: quota exhaustion puts it in NOSPACE and it stops accepting
// writes while keeping its leader, so has_leader alone reports a wedged cluster as healthy.
var EssentialServices = []EssentialService{
	{"apiserver",
		`(up{job="apiserver"} == 0) or absent(up{job="apiserver"})`,
		`up{job="apiserver"}`},
	{"etcd",
		`(etcd_server_has_leader == 0) or absent(etcd_server_has_leader) ` +
			`or (etcd_mvcc_db_total_size_in_bytes >= etcd_server_quota_backend_bytes)`,
		`etcd_server_has_leader`},
	{"runai-backend-cluster-service",
		`(up{job="runai-backend-cluster-service"} == 0) or absent(up{job="runai-backend-cluster-service"})`,
		`up{job="runai-backend-cluster-service"}`},
	{"runai-backend-workloads",
		`(up{job="runai-backend-workloads"} == 0) or absent(up{job="runai-backend-workloads"})`,
		`up{job="runai-backend-workloads"}`},
}

// finalServiceUnavailableSecs returns seconds each essential service spent unusable.
// Seconds rather than a boolean because one missed scrape is noise while the ~60s outage a
// control-plane restart causes invalidates the run. Services absent from the map were not
// scraped at all.
func finalServiceUnavailableSecs(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (map[string]float64, error) {
	const sampleSecs = 15
	out := map[string]float64{}
	for _, svc := range EssentialServices {
		// Presence is asked of the whole run, not of endTime: a service that is still down
		// when the run ends has no series at that instant, and asking only there would
		// report the outage as "never scraped" — the same blind spot the Down expressions
		// guard against.
		present, _, err := queryInstantScalarAt(ctx, api,
			fmt.Sprintf(`count(count_over_time((%s)[%s:%ds])) or vector(0)`,
				svc.Present, promDuration(elapsed), sampleSecs), endTime)
		if err != nil {
			return nil, fmt.Errorf("presence query for %s: %w", svc.Name, err)
		}
		if present == 0 {
			continue
		}
		secs, available, err := queryInstantScalarAt(ctx, api, fmt.Sprintf(
			`(count_over_time((%s)[%s:%ds]) * %d) or vector(0)`,
			svc.Down, promDuration(elapsed), sampleSecs, sampleSecs), endTime)
		if err != nil {
			return nil, fmt.Errorf("unavailability query for %s: %w", svc.Name, err)
		}
		if available {
			out[svc.Name] = secs
		}
	}
	return out, nil
}

// finalEtcdWALFsyncP99 returns the worst 5-minute-window P99 etcd WAL fsync latency over the run.
func finalEtcdWALFsyncP99(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (float64, bool, error) {
	q := fmt.Sprintf(
		`max_over_time(histogram_quantile(0.99, sum(rate(etcd_disk_wal_fsync_duration_seconds_bucket[5m])) by (le))[%s:5m])`,
		promDuration(elapsed),
	)
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// finalEtcdProposalsPeakAndSustained returns:
//   - peak: max value of etcd_server_proposals_pending during the run
//   - sustained: whether it was > 0 for any continuous sustainedSecs window
func finalEtcdProposalsPeakAndSustained(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time, sustainedSecs, depthThreshold int) (peak float64, sustained bool, available bool, err error) {
	dur := promDuration(elapsed)

	peakQ := fmt.Sprintf(`max_over_time(etcd_server_proposals_pending[%s])`, dur)
	peak, available, err = queryInstantScalarAt(ctx, api, peakQ, endTime)
	if err != nil || !available {
		return
	}

	// Degraded means proposals stayed above the threshold for a whole sustainedSecs
	// window. min_over_time exceeds the threshold only if the metric never dropped below
	// it across that window; the outer max_over_time looks for any such window.
	sustainedQ := fmt.Sprintf(
		`count(max_over_time((min_over_time(etcd_server_proposals_pending[%ds]) > %d)[%s:%ds]) > 0) > 0`,
		sustainedSecs, depthThreshold, dur, sustainedSecs,
	)
	result, _, _, qErr := api.Query(ctx, sustainedQ, endTime)
	if qErr != nil {
		err = fmt.Errorf("sustained proposals query failed: %w", qErr)
		return
	}
	vec, ok := result.(model.Vector)
	sustained = ok && len(vec) > 0
	return
}

// finalEtcdDBSizeRatioPeak returns the peak etcd DB usage as a fraction of quota over the run.
func finalEtcdDBSizeRatioPeak(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (float64, bool, error) {
	q := fmt.Sprintf(
		`max_over_time((etcd_mvcc_db_total_size_in_bytes / etcd_server_quota_backend_bytes)[%s:30s])`,
		promDuration(elapsed),
	)
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// heartbeatPodRestartDetails returns one line per container that restarted in the last 5m.
// Called immediately when the heartbeat detects restarts > 0 so logs can be collected
// while they are still fresh on the node.
func heartbeatPodRestartDetails(ctx context.Context, api v1.API, namespaces []string) ([]string, error) {
	nsRegex := namespacesRegex(namespaces)
	q := fmt.Sprintf(
		`increase(kube_pod_container_status_restarts_total{namespace=~"%s"}[5m]) > 0`,
		nsRegex,
	)
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("heartbeat pod restart details query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	details := make([]string, 0, len(vec))
	for _, s := range vec {
		ns := string(s.Metric["namespace"])
		pod := string(s.Metric["pod"])
		container := string(s.Metric["container"])
		details = append(details, fmt.Sprintf("%s/%s/%s: %.0f restarts", ns, pod, container, float64(s.Value)))
	}
	return details, nil
}

// heartbeatPodRestartTotals returns the cumulative restart count per container, keyed
// "namespace/pod/container". The 5m window behind heartbeatPodRestartDetails reports the
// same crash on two consecutive heartbeats; the cumulative counter tells a repeat apart
// from a new one.
func heartbeatPodRestartTotals(ctx context.Context, api v1.API, namespaces []string) (map[string]float64, error) {
	q := fmt.Sprintf(`kube_pod_container_status_restarts_total{namespace=~"%s"}`,
		namespacesRegex(namespaces))
	result, _, _, err := api.Query(ctx, q, time.Now())
	if err != nil {
		return nil, fmt.Errorf("heartbeat pod restart totals query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		key := fmt.Sprintf("%s/%s/%s", s.Metric["namespace"], s.Metric["pod"], s.Metric["container"])
		out[key] = float64(s.Value)
	}
	return out, nil
}

// finalPodRestartDetails returns one line per container that restarted during the run,
// formatted as "namespace/pod/container: N restarts". Only called when the aggregate
// restart count is > 0 so the caller can attach it to the FAIL result for post-run investigation.
func finalPodRestartDetails(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time, namespaces []string) ([]string, error) {
	nsRegex := namespacesRegex(namespaces)
	q := fmt.Sprintf(
		`increase(kube_pod_container_status_restarts_total{namespace=~"%s"}[%s]) > 0`,
		nsRegex, promDuration(elapsed),
	)
	result, _, _, err := api.Query(ctx, q, endTime)
	if err != nil {
		return nil, fmt.Errorf("pod restart details query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok || len(vec) == 0 {
		return nil, nil
	}
	details := make([]string, 0, len(vec))
	for _, s := range vec {
		ns := string(s.Metric["namespace"])
		pod := string(s.Metric["pod"])
		container := string(s.Metric["container"])
		details = append(details, fmt.Sprintf("%s/%s/%s: %.0f restarts", ns, pod, container, float64(s.Value)))
	}
	return details, nil
}

// finalPodRestarts returns total restarts in target namespaces over the run.
func finalPodRestarts(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time, namespaces []string) (float64, bool, error) {
	nsRegex := namespacesRegex(namespaces)
	q := fmt.Sprintf(
		`sum(increase(kube_pod_container_status_restarts_total{namespace=~"%s"}[%s])) or vector(0)`,
		nsRegex, promDuration(elapsed),
	)
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// finalNodeNotReadyPeak returns the peak count of NotReady nodes during the run.
func finalNodeNotReadyPeak(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (float64, bool, error) {
	q := fmt.Sprintf(
		`max_over_time((count(kube_node_status_condition{condition="Ready",status="false"} == 1) or vector(0))[%s:30s])`,
		promDuration(elapsed),
	)
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// finalSystemPodsUnhealthyPeak returns the peak count of system pods not in Running or Succeeded
// phase across the target namespaces over the full run.
func finalSystemPodsUnhealthyPeak(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time, namespaces []string) (float64, bool, error) {
	nsRegex := namespacesRegex(namespaces)
	q := fmt.Sprintf(
		`max_over_time((count(kube_pod_status_phase{namespace=~"%s", phase!~"Running|Succeeded"} == 1) or vector(0))[%s:30s])`,
		nsRegex, promDuration(elapsed),
	)
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// finalPrometheusScrapGapSecs returns the total seconds during the run where any source in
// CriticalMetricSources had up==0. Uses the same job list as the heartbeat source-health check
// so both give consistent signal: heartbeats flag gaps in real time, final report totals them.
func finalPrometheusScrapGapSecs(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (float64, bool, error) {
	dur := promDuration(elapsed)
	jobRegex := strings.Join(CriticalMetricSources, "|")
	// `or vector(0)` ensures the query returns 0 (not empty) when all sources are healthy
	// throughout the run, so the caller sees "0s gap" (PASS) rather than UNAVAILABLE.
	q := fmt.Sprintf(
		`(count_over_time((min(up{job=~%q}) == 0)[%s:30s]) * 30) or vector(0)`,
		jobRegex, dur,
	)
	return queryInstantScalarAt(ctx, api, q, endTime)
}

// --- QoS metric final queries ---

// finalWorkloadSubmissionLatency returns P50 and P99 of workload submission HTTP request duration.
// Filters to POST requests on top-level workload type paths (e.g. POST /api/v1/workloads/trainings)
// and excludes internal sub-paths like /updates, /events, /sync which are background polling calls.
func finalWorkloadSubmissionLatency(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (p50, p99 float64, available bool, err error) {
	dur := promDuration(elapsed)
	base := `runai_http_request_duration_seconds_bucket{method="POST",path=~"/api/v[12]/workloads/(trainings|workspaces|inferences|distributed|distributed-inferences|unified)"}[5m]`

	p50Q := fmt.Sprintf(`max_over_time(histogram_quantile(0.50, sum(rate(%s)) by (le))[%s:5m])`, base, dur)
	p99Q := fmt.Sprintf(`max_over_time(histogram_quantile(0.99, sum(rate(%s)) by (le))[%s:5m])`, base, dur)

	p50, available, err = queryInstantScalarAt(ctx, api, p50Q, endTime)
	if err != nil || !available {
		return
	}
	p99, available, err = queryInstantScalarAt(ctx, api, p99Q, endTime)
	return
}

// finalSchedulerCycleLatency returns the avg and peak duration of the scheduler's allocate
// cycle (runai_action_scheduling_latency_milliseconds{action="allocate"}) over the run.
// This is a gauge set once per scheduling loop iteration covering all pending workloads,
// so it measures algorithm execution time — not per-workload queue wait time.
// Values are in milliseconds.
// SchedulerCycleStat is one scheduler's allocate cycle latency over the run.
type SchedulerCycleStat struct {
	Avg  float64
	Peak float64
}

// finalSchedulerCycleLatencyByScheduler returns allocate cycle latency keyed by `job`
// (= runai-scheduler-<nodepool>; the metric carries no nodepool label). Per scheduler
// because idle nodepool schedulers sit at zero: an unaggregated query can return the idle
// one and report 0ms for the whole run.
func finalSchedulerCycleLatencyByScheduler(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (map[string]SchedulerCycleStat, error) {
	dur := promDuration(elapsed)
	base := `runai_action_scheduling_latency_milliseconds{action="allocate"}`

	out := map[string]SchedulerCycleStat{}
	for _, q := range []struct {
		query string
		set   func(*SchedulerCycleStat, float64)
	}{
		{fmt.Sprintf(`avg_over_time(%s[%s])`, base, dur), func(s *SchedulerCycleStat, v float64) { s.Avg = v }},
		{fmt.Sprintf(`max_over_time(%s[%s])`, base, dur), func(s *SchedulerCycleStat, v float64) { s.Peak = v }},
	} {
		res, _, _, err := api.Query(ctx, q.query, endTime)
		if err != nil {
			return nil, fmt.Errorf("querying scheduler cycle latency: %w", err)
		}
		vec, ok := res.(model.Vector)
		if !ok {
			continue
		}
		for _, s := range vec {
			key := string(s.Metric["job"])
			if key == "" {
				key = string(s.Metric["pod"])
			}
			stat := out[key]
			q.set(&stat, float64(s.Value))
			out[key] = stat
		}
	}
	return out, nil
}

// finalReconciliationLatencyP99ByController returns a map of controller→peak P99 reconciliation latency.
func finalReconciliationLatencyP99ByController(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (map[string]float64, error) {
	q := fmt.Sprintf(
		`max_over_time(histogram_quantile(0.99, sum(rate(controller_runtime_reconcile_time_seconds_bucket[5m])) by (le, controller))[%s:5m])`,
		promDuration(elapsed),
	)
	result, _, _, err := api.Query(ctx, q, endTime)
	if err != nil {
		return nil, fmt.Errorf("reconciliation latency query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		controller := string(s.Metric["controller"])
		if controller == "" {
			controller = "unknown"
		}
		out[controller] = float64(s.Value)
	}
	return out, nil
}

// WorkQueueStat is one queue's depth and wait time over the run.
type WorkQueueStat struct {
	Peak float64
	// Sustained is true if depth stayed above the threshold for a whole sustainedSecs
	// window at any point — a backlog rather than a transient spike.
	Sustained bool
	// Wait* is how long an item sat in the queue before being reconciled — how stale the
	// controller's view is, where depth only says how many items wait. Buckets are decades
	// (…, 1, 10, 100, 1000, +Inf), so each value carries about one significant digit and
	// exactly 1000 means the quantile fell in +Inf and is unbounded.
	WaitP50 float64
	WaitP95 float64
	WaitP99 float64
}

// finalWorkQueueDepthByController returns a map of controller→depth stats.
func finalWorkQueueDepthByController(
	ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time,
	threshold int, sustainedSecs int,
) (map[string]WorkQueueStat, error) {
	dur := promDuration(elapsed)

	peaks, err := workQueueVector(ctx, api, endTime,
		fmt.Sprintf(`max_over_time(workqueue_depth[%s])`, dur))
	if err != nil {
		return nil, err
	}
	if peaks == nil {
		return nil, nil
	}

	// min_over_time over a sustainedSecs window is above the threshold only if the depth
	// never dropped below it for that whole window.
	sustained, err := workQueueVector(ctx, api, endTime, fmt.Sprintf(
		`max_over_time((min_over_time(workqueue_depth[%ds]) > %d)[%s:%ds])`,
		sustainedSecs, threshold, dur, sustainedSecs))
	if err != nil {
		return nil, err
	}

	// Worst 5m-window wait over the run, per quantile. Reported alongside depth because depth
	// alone does not say how far behind the controller actually fell.
	waits := map[float64]map[string]float64{}
	for _, q := range []float64{0.50, 0.95, 0.99} {
		w, err := workQueueVector(ctx, api, endTime, fmt.Sprintf(
			`max_over_time(histogram_quantile(%.2f, sum by (name, namespace, pod, le) (rate(workqueue_queue_duration_seconds_bucket[5m])))[%s:5m])`,
			q, dur))
		if err != nil {
			return nil, err
		}
		waits[q] = w
	}

	out := make(map[string]WorkQueueStat, len(peaks))
	for k, peak := range peaks {
		_, isSustained := sustained[k]
		out[k] = WorkQueueStat{
			Peak:      peak,
			Sustained: isSustained,
			WaitP50:   waits[0.50][k],
			WaitP95:   waits[0.95][k],
			WaitP99:   waits[0.99][k],
		}
	}
	return out, nil
}

func workQueueVector(ctx context.Context, api v1.API, endTime time.Time, q string) (map[string]float64, error) {
	result, _, _, err := api.Query(ctx, q, endTime)
	if err != nil {
		return nil, fmt.Errorf("workqueue depth query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, nil
	}
	out := make(map[string]float64)
	for _, s := range vec {
		v := float64(s.Value)
		// NaN comes from histogram_quantile over a queue with no observations in the
		// window; zero means the queue never filled. Neither is worth a row.
		if v == 0 || math.IsNaN(v) {
			continue
		}
		out[workQueueKey(s.Metric)] = v
	}
	return out, nil
}

// workQueueKey identifies which component a queue belongs to. The name label alone is
// ambiguous: "queue", "project", "pod", "nodepool" and "department" each appear on more
// than one series. Pods identify Run:ai controllers; kubelet and apiserver series carry
// only a job.
func workQueueKey(m model.Metric) string {
	name := string(m["name"])
	if name == "" {
		name = "unknown"
	}
	switch {
	case m["pod"] != "":
		return fmt.Sprintf("%s/%s %s", m["namespace"], m["pod"], name)
	case m["job"] != "":
		return fmt.Sprintf("%s %s", m["job"], name)
	default:
		return name
	}
}

// heartbeatWorkloadListLatencyP95 returns current P95 workload list API latency (seconds)
// over the last 5m. Uses path="/api/v1/workloads" to select only the list endpoint,
// not submission calls (which go to /api/v1/workloads/trainings).
//
// TODO: If this metric is absent (no workload-list traffic during the run), fall back to
// histogram_quantile(0.95, ...) over the api-exerciser's own histogram metric once that
// story lands and Prometheus is configured to scrape the exerciser's /metrics endpoint.
func heartbeatWorkloadListLatencyP95(ctx context.Context, api v1.API) (float64, bool, error) {
	q := `histogram_quantile(0.95, sum(rate(runai_http_request_duration_seconds_bucket{method="GET",path="/api/v1/workloads"}[5m])) by (le))`
	return queryInstantScalar(ctx, api, q)
}

// finalWorkloadListLatency returns the worst 5-minute-window P50 and P95 workload list API
// latency over the full run. P50 is reported alongside P95 so a heavy tail can be told
// apart from a uniformly slow list, as submission latency already does.
func finalWorkloadListLatency(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (p50, p95 float64, available bool, err error) {
	dur := promDuration(elapsed)
	// method="GET" must match heartbeatWorkloadListLatencyP95: without it the final number
	// mixes in non-GET traffic on the same path and disagrees with the live heartbeats.
	q := func(quantile string) string {
		return fmt.Sprintf(
			`max_over_time(histogram_quantile(%s, sum(rate(runai_http_request_duration_seconds_bucket{method="GET",path="/api/v1/workloads"}[5m])) by (le))[%s:5m])`,
			quantile, dur,
		)
	}
	if p50, available, err = queryInstantScalarAt(ctx, api, q("0.50"), endTime); err != nil || !available {
		return
	}
	p95, available, err = queryInstantScalarAt(ctx, api, q("0.95"), endTime)
	return
}

// --- hardware sufficiency ---

// finalNodeCPUUtilizationPeak returns, for each node, the peak CPU utilization fraction (0–1)
// over the run. Uses node_exporter's node_cpu_seconds_total{mode="idle"} so it reflects
// actual host-level saturation regardless of namespace or cgroup boundaries.
func finalNodeCPUUtilizationPeak(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (map[string]float64, error) {
	dur := promDuration(elapsed)
	// Peak of the per-node utilization over the run, sampled at 30s resolution.
	q := fmt.Sprintf(
		`max_over_time((1 - avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])))[%s:30s])`,
		dur,
	)
	result, _, _, err := api.Query(ctx, q, endTime)
	if err != nil {
		return nil, fmt.Errorf("node CPU utilization query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		instance := string(s.Metric["instance"])
		if instance == "" {
			instance = "unknown"
		}
		out[instance] = float64(s.Value)
	}
	return out, nil
}

// finalNodeMemoryUtilizationPeak returns peak used-memory fraction per node.
// Uses MemAvailable rather than MemFree so page cache is not counted as used.
func finalNodeMemoryUtilizationPeak(ctx context.Context, api v1.API, elapsed time.Duration, endTime time.Time) (map[string]float64, error) {
	dur := promDuration(elapsed)
	q := fmt.Sprintf(
		`max_over_time((1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes))[%s:30s])`,
		dur,
	)
	result, _, _, err := api.Query(ctx, q, endTime)
	if err != nil {
		return nil, fmt.Errorf("node memory utilization query failed: %w", err)
	}
	vec, ok := result.(model.Vector)
	if !ok {
		return nil, nil
	}
	out := make(map[string]float64, len(vec))
	for _, s := range vec {
		instance := string(s.Metric["instance"])
		if instance == "" {
			instance = "unknown"
		}
		out[instance] = float64(s.Value)
	}
	return out, nil
}

// --- helpers ---

func promDuration(d time.Duration) string {
	// Prometheus duration strings use integer seconds.
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

func namespacesRegex(namespaces []string) string {
	if len(namespaces) == 0 {
		return ".*"
	}
	result := ""
	for i, ns := range namespaces {
		if i > 0 {
			result += "|"
		}
		result += ns
	}
	return result
}
