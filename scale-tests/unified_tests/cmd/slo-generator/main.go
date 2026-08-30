package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
)

// MetricData represents a single metric measurement
type MetricData struct {
	Value float64 `json:"value"`
	Time  string  `json:"time"`
}

// LatencyData represents latency metrics
type LatencyData struct {
	Mean   float64 `json:"mean"`
	Max    float64 `json:"max"`
	Min    float64 `json:"min"`
	P95    float64 `json:"p95"`
	P50    float64 `json:"p50"`
	P75    float64 `json:"p75"`
	StdDev float64 `json:"std_dev"`
}

// PerformanceMetric represents a single performance metric
type PerformanceMetric struct {
	PendingRate float64     `json:"pending_rate"`
	LatencyMs   LatencyData `json:"latency_ms"`
	SampleSize  int         `json:"sample_size"`
}

// MetricsReport represents the structure of metrics data
type MetricsReport struct {
	ID       string `json:"id"`
	Metadata struct {
		ScaleGroup string `json:"scale_group"`
	} `json:"metadata"`
	PerformanceMetrics struct {
		WorkloadFirstPodCreation   PerformanceMetric `json:"workload_first_pod_creation"`
		WorkloadPodScheduling      PerformanceMetric `json:"workload_pod_scheduling"`
		WorkloadSyncToCluster      PerformanceMetric `json:"workload_sync_to_cluster"`
		WorkloadSyncToControlPlane PerformanceMetric `json:"workload_sync_to_control_plane"`
		WorkloadWholeLifecycle     PerformanceMetric `json:"workload_whole_lifecycle"`
	} `json:"performance_metrics"`
}

// MetricThreshold represents the SLO configuration for a single metric
type MetricThreshold struct {
	Target     float64 `json:"target"`      // Target value (e.g., p95 latency)
	Deviation  float64 `json:"deviation"`   // Allowed deviation as percentage
	MetricName string  `json:"metric_name"` // Name of the metric
	Stage      string  `json:"stage"`       // Lifecycle stage
}

// ScaleGroupSLOs contains all SLOs for a specific scale group
type ScaleGroupSLOs struct {
	GroupName string            `json:"group_name"`
	Metrics   []MetricThreshold `json:"metrics"`
}

// SLOConfig represents the complete SLO configuration
type SLOConfig struct {
	ScaleGroups []ScaleGroupSLOs `json:"scale_groups"`
}

// LokiEntry represents a single entry from Grafana Loki
type LokiEntry struct {
	Line      string                 `json:"line"`
	Timestamp string                 `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
}

// MetricValue represents a single metric's p95 value
type MetricValue struct {
	MetricName string  `json:"metric_name"`
	P95Value   float64 `json:"p95_value_ms"`
}

// ScaleGroupMetrics contains all metrics for a specific scale group
type ScaleGroupMetrics struct {
	GroupName string        `json:"group_name"`
	Metrics   []MetricValue `json:"metrics"`
}

// MetricsConfig represents the complete metrics configuration
type MetricsConfig struct {
	ScaleGroups []ScaleGroupMetrics `json:"scale_groups"`
}

// calculateP95 calculates the 95th percentile of a slice of float64 values
func calculateP95(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Float64s(values)
	index := int(math.Ceil(float64(len(values))*0.95)) - 1
	return values[index]
}

// extractMetrics extracts relevant metrics from a report
func extractMetrics(report MetricsReport) map[string]float64 {
	metrics := make(map[string]float64)
	perf := report.PerformanceMetrics

	// Helper function to extract metric value
	extractValue := func(metric PerformanceMetric) float64 {
		if metric.LatencyMs.P95 > 0 {
			return metric.LatencyMs.P95
		}
		return metric.LatencyMs.Mean
	}

	// Extract metrics
	metrics["workload_first_pod_creation"] = extractValue(perf.WorkloadFirstPodCreation)
	metrics["workload_pod_scheduling"] = extractValue(perf.WorkloadPodScheduling)
	metrics["workload_sync_to_cluster"] = extractValue(perf.WorkloadSyncToCluster)
	metrics["workload_sync_to_control_plane"] = extractValue(perf.WorkloadSyncToControlPlane)
	metrics["workload_whole_lifecycle"] = extractValue(perf.WorkloadWholeLifecycle)

	return metrics
}

// generateMetricsConfig generates the metrics configuration from the processed metrics
func generateMetricsConfig(metricsByGroup map[string][]map[string]float64) MetricsConfig {
	config := MetricsConfig{
		ScaleGroups: make([]ScaleGroupMetrics, 0),
	}

	for groupName, metricsList := range metricsByGroup {
		groupConfig := ScaleGroupMetrics{
			GroupName: groupName,
			Metrics:   make([]MetricValue, 0),
		}

		// Calculate average p95 for each metric
		metricValues := make(map[string][]float64)
		for _, metrics := range metricsList {
			for metricName, value := range metrics {
				metricValues[metricName] = append(metricValues[metricName], value)
			}
		}

		// Create metric entries
		for metricName, values := range metricValues {
			// Calculate P95 of all P95 values (instead of average)
			p95Value := calculateP95(values)

			if p95Value > 0 { // Only include metrics with valid values
				metricConfig := MetricValue{
					MetricName: metricName,
					P95Value:   math.Round(p95Value*100) / 100, // Round to 2 decimal places
				}
				groupConfig.Metrics = append(groupConfig.Metrics, metricConfig)
			}
		}

		config.ScaleGroups = append(config.ScaleGroups, groupConfig)
	}

	return config
}

func main() {
	// Parse command line flags
	metricsDir := flag.String("metrics-dir", "", "Directory containing metrics data")
	outputPath := flag.String("output", "metrics_p95.json", "Output path for generated metrics configuration")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if *metricsDir == "" {
		log.Fatal("--metrics-dir is required")
	}

	// Ensure metrics directory exists
	if _, err := os.Stat(*metricsDir); os.IsNotExist(err) {
		log.Fatalf("Metrics directory %s does not exist", *metricsDir)
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(*outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Read and process metrics files
	metricsByGroup := make(map[string][]map[string]float64)
	err := filepath.Walk(*metricsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		if *debug {
			log.Printf("Processing file: %s", path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read metrics file %s: %w", path, err)
		}

		// Parse the Loki entries
		var entries []LokiEntry
		if err := json.Unmarshal(data, &entries); err != nil {
			if *debug {
				log.Printf("Failed to parse Loki entries: %v", err)
			}
			return fmt.Errorf("failed to parse metrics file %s: %w", path, err)
		}

		if *debug {
			log.Printf("Found %d entries in file", len(entries))
		}

		// Process each entry
		for i, entry := range entries {
			// Parse the metrics data from the line field
			var report MetricsReport
			if err := json.Unmarshal([]byte(entry.Line), &report); err != nil {
				if *debug {
					log.Printf("Failed to parse metrics data for entry %d: %v", i, err)
				}
				continue
			}

			scaleGroup := report.Metadata.ScaleGroup
			if scaleGroup == "" {
				if *debug {
					log.Printf("Skipping entry %d: empty scale group", i)
				}
				continue
			}

			if *debug {
				log.Printf("Processing entry %d for scale group: %s", i, scaleGroup)
			}

			metrics := extractMetrics(report)
			if *debug {
				log.Printf("Extracted metrics: %+v", metrics)
			}
			metricsByGroup[scaleGroup] = append(metricsByGroup[scaleGroup], metrics)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("Failed to process metrics directory: %v", err)
	}

	if *debug {
		log.Printf("Processed metrics by group: %+v", metricsByGroup)
	}

	// Generate metrics configuration
	metricsConfig := generateMetricsConfig(metricsByGroup)

	// Write output file
	outputData, err := json.MarshalIndent(metricsConfig, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics config: %v", err)
	}

	if err := os.WriteFile(*outputPath, outputData, 0644); err != nil {
		log.Fatalf("Failed to write metrics config: %v", err)
	}

	fmt.Printf("Successfully generated metrics configuration at %s\n", *outputPath)
}
