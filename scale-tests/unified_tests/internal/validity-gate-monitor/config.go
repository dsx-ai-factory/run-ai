package validity_gate_monitor

import (
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type MonitorConfig struct {
	// HeartbeatInterval controls how often to run the sanity heartbeat checks during the run.
	HeartbeatInterval string `mapstructure:"heartbeat_interval"`

	// RunDuration is the total expected test duration. When set, the monitor exits
	// after this duration. If empty it runs until SIGTERM.
	RunDuration string `mapstructure:"run_duration"`

	OutputDir string `mapstructure:"output_dir"`

	// Prometheus pod to port-forward to for metric queries.
	PrometheusPodName   string `mapstructure:"prometheus_pod_name"`
	PrometheusNamespace string `mapstructure:"prometheus_namespace"`
	PrometheusPort      int    `mapstructure:"prometheus_port"`

	// TargetNamespaces are the Run:AI system namespaces checked for pod restarts
	// and system pod health. Workloads run in project namespaces and are never
	// scheduled into these namespaces, so no exclusion filter is needed.
	TargetNamespaces []string `mapstructure:"target_namespaces"`

	// EtcdProposalSustainedSecs is how long etcd_server_proposals_pending must stay above
	// EtcdProposalDepthThreshold before it counts as a validity gate failure.
	EtcdProposalSustainedSecs int `mapstructure:"etcd_proposal_sustained_secs"`

	// EtcdProposalDepthThreshold is the depth above which pending proposals mean etcd is
	// not keeping up. A small backlog under load is normal.
	EtcdProposalDepthThreshold int `mapstructure:"etcd_proposal_depth_threshold"`

	// Workload progress. Passed in by the orchestrator rather than read from another
	// binary's output, so the monitor stays a pure measurement tool.
	//
	// ProjectPrefix selects the namespaces workload pods land in; Run:ai names them
	// runai-<project>.
	ProjectPrefix string `mapstructure:"project_prefix"`
	// SubmissionRatePerMin is the steady-state arrival rate R, used to report how much of
	// the offered load actually reached the API.
	SubmissionRatePerMin float64 `mapstructure:"submission_rate_per_min"`
	// WorkloadDurationSecs is D, derived by the submitter as W_cap/R. Completions arrive in
	// cohorts of this period, so it is also the longest a healthy run may go with nothing
	// completing.
	WorkloadDurationSecs int `mapstructure:"workload_duration_secs"`
	// AutoDeletionSecs is the workload TTL; 0 means completed workloads are never removed.
	AutoDeletionSecs int `mapstructure:"auto_deletion_secs"`
	// CompletionGapThresholdSecs is how long the run may go with nothing completing before it
	// counts as stalled. Depends on the submission shape, so the orchestrator sets it.
	CompletionGapThresholdSecs int `mapstructure:"completion_gap_threshold_secs"`
	// PodsPerWorkload converts the pod counts these gates measure into workloads, since
	// kube_pod_status_phase counts pods and a distributed workload owns several.
	PodsPerWorkload int `mapstructure:"pods_per_workload"`

	// Degraded-metric thresholds. Loose placeholders — retune once a Large run shows
	// what normal looks like.
	SubmissionLatencyP99Threshold     float64 `mapstructure:"submission_latency_p99_threshold"`
	ReconciliationLatencyP99Threshold float64 `mapstructure:"reconciliation_latency_p99_threshold"`
	SchedulerCyclePeakThresholdMs     float64 `mapstructure:"scheduler_cycle_peak_threshold_ms"`
}

func (c *MonitorConfig) SetDefaults() {
	c.HeartbeatInterval = "5m"
	c.RunDuration = ""
	c.OutputDir = "/tmp/validity-gate"

	c.PrometheusPodName = "prometheus-runai-0"
	c.PrometheusNamespace = "runai"
	c.PrometheusPort = 9090

	c.TargetNamespaces = []string{"runai", "runai-backend"}
	c.EtcdProposalSustainedSecs = 120
	c.EtcdProposalDepthThreshold = 10

	c.ProjectPrefix = "load-project"
	c.SubmissionRatePerMin = 0
	c.WorkloadDurationSecs = 0
	c.AutoDeletionSecs = 0
	c.CompletionGapThresholdSecs = 0
	c.PodsPerWorkload = 1

	c.SubmissionLatencyP99Threshold = 2.0
	c.ReconciliationLatencyP99Threshold = 1.0
	c.SchedulerCyclePeakThresholdMs = 30000
}

func ConstructConfig() *MonitorConfig {
	cfg, err := config.GetR[*MonitorConfig]()
	if err != nil {
		logging.Fatalf("failed to register validity gate monitor config: %v", err)
	}
	return cfg
}
