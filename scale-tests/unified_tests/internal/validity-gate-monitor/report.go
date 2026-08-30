package validity_gate_monitor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	StatusPass        = "PASS"
	StatusFail        = "FAIL"
	StatusUnavailable = "UNAVAILABLE"
	// StatusDegraded marks a metric that breached its threshold but does not invalidate
	// the run. The infrastructure is stable and run data is usable; however, some aspect
	// of system performance (API latency, scheduler throughput, controller backlogs) was
	// outside the expected range. Sets Degraded=true without affecting OverallStatus or InfraLimited.
	StatusDegraded = "DEGRADED"
)

// MetricResult is a single metric entry in the validity gate report.
// Matches the structured format from the ticket spec:
//
//	{ "metric": "etcd_wal_fsync_p99", "value": "45ms", "threshold": "100ms", "headroom": "55%", "status": "PASS" }
type MetricResult struct {
	Metric    string   `json:"metric"`
	Value     string   `json:"value"`
	Threshold string   `json:"threshold,omitempty"`
	Headroom  string   `json:"headroom,omitempty"`
	Status    string   `json:"status"`
	Detail    []string `json:"detail,omitempty"`
}

// ValidityGateReport is the top-level output written to disk after the run.
//
// OverallStatus / InfraLimited: FAIL when any ValidityGate or HardwareSufficiency
// metric breaches its threshold — the infrastructure cannot be trusted as a fair
// baseline for resource recommendation measurements.
//
// Degraded: true when any DegradedMetric shows StatusDegraded. Infrastructure is
// stable and run data is usable, but some aspect of system performance was outside
// the expected range during the run. Does not set OverallStatus=FAIL or InfraLimited=true.
type ValidityGateReport struct {
	OverallStatus       string         `json:"overall_status"`
	InfraLimited        bool           `json:"infra_limited"`
	Degraded            bool           `json:"degraded"`
	ValidityGates       []MetricResult `json:"validity_gates"`
	HardwareSufficiency []MetricResult `json:"hardware_sufficiency"`
	DegradedMetrics     []MetricResult `json:"degraded_metrics"`
}

func (r *ValidityGateReport) computeOverallStatus() {
	r.OverallStatus = StatusPass

	for _, m := range r.ValidityGates {
		if m.Status == StatusFail {
			r.OverallStatus = StatusFail
			r.InfraLimited = true
		}
	}
	for _, m := range r.HardwareSufficiency {
		if m.Status == StatusFail {
			r.OverallStatus = StatusFail
			r.InfraLimited = true
		}
	}
	for _, m := range r.DegradedMetrics {
		if m.Status == StatusDegraded {
			r.Degraded = true
		}
	}
}

// WriteReport marshals the report to JSON and writes it to outputDir/validity-gate-report.json.
func WriteReport(report *ValidityGateReport, outputDir string) error {
	report.computeOverallStatus()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir %s: %w", outputDir, err)
	}

	path := filepath.Join(outputDir, "validity-gate-report.json")
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("failed to write report to %s: %w", path, err)
	}
	return nil
}
