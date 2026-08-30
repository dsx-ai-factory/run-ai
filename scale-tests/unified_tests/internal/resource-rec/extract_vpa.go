package resource_rec

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"context"

	corev1 "k8s.io/api/core/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RecommendationRow is one output record: per-container VPA recommendation
// for a single workload container.
type RecommendationRow struct {
	Namespace     string `json:"namespace"`
	Workload      string `json:"workload"`
	Container     string `json:"container"`
	Source        string `json:"source"`
	CPULowerBound string `json:"cpu_lower_bound"`
	CPUTarget     string `json:"cpu_target"`
	CPUUpperBound string `json:"cpu_upper_bound"`
	MemLowerBound string `json:"memory_lower_bound"`
	MemTarget     string `json:"memory_target"`
	MemUpperBound string `json:"memory_upper_bound"`
}

// ExtractVPARecommendations reads the current VPA status across the given namespaces
// and writes per-container recommendation rows as both CSV and JSON under outputDir.
// VPA objects with an empty recommendation do not abort the extraction; they are recorded
// in vpa-skipped.json so the caller can surface them.
func ExtractVPARecommendations(ctx context.Context, k8sClient client.Client, namespaces []string, outputDir string) error {
	var vpas []vpav1.VerticalPodAutoscaler
	for _, ns := range namespaces {
		list := &vpav1.VerticalPodAutoscalerList{}
		if err := k8sClient.List(ctx, list, client.InNamespace(ns)); err != nil {
			return fmt.Errorf("listing VPAs in namespace %s: %w", ns, err)
		}
		vpas = append(vpas, list.Items...)
	}
	fmt.Printf("found %d VPA objects across %v\n", len(vpas), namespaces)

	rows, skipped := buildRows(vpas)
	fmt.Printf("extracted %d recommendation rows, %d VPAs skipped\n", len(rows), len(skipped))

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir %s: %w", outputDir, err)
	}
	if err := writeJSON(rows, filepath.Join(outputDir, "vpa-recommendations.json")); err != nil {
		return err
	}
	// Written even when empty so the caller can always read it.
	if err := writeJSON(skipped, filepath.Join(outputDir, "vpa-skipped.json")); err != nil {
		return err
	}
	return writeCSV(rows, filepath.Join(outputDir, "vpa-recommendations.csv"))
}

// SkippedVPA is a VPA that produced no recommendation, so its workload is absent from the
// table. Emitted as a file rather than logged so a caller can surface it however it likes.
type SkippedVPA struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Target    string `json:"target"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
}

const sourceVPA = "vpa"

// skipReason prefers ConfigUnsupported over RecommendationProvided: an unresolvable target
// also reports NoPodsMatched, but that is the symptom rather than the cause.
func skipReason(vpa vpav1.VerticalPodAutoscaler) (reason, message string) {
	var fallback *vpav1.VerticalPodAutoscalerCondition
	for i, c := range vpa.Status.Conditions {
		switch c.Type {
		case vpav1.ConfigUnsupported:
			if c.Status == corev1.ConditionTrue {
				r := c.Reason
				if r == "" {
					r = string(vpav1.ConfigUnsupported)
				}
				return r, c.Message
			}
		case vpav1.RecommendationProvided:
			fallback = &vpa.Status.Conditions[i]
		}
	}
	if fallback != nil {
		return fallback.Reason, fallback.Message
	}
	return "Unknown", "VPA reported no RecommendationProvided condition"
}

func buildRows(vpas []vpav1.VerticalPodAutoscaler) ([]RecommendationRow, []SkippedVPA) {
	rows := []RecommendationRow{}
	skipped := []SkippedVPA{}
	for _, vpa := range vpas {
		if vpa.Status.Recommendation == nil || len(vpa.Status.Recommendation.ContainerRecommendations) == 0 {
			// A target VPA cannot resolve (e.g. a StatefulSet owned by a CR such as
			// Prometheus) drops silently out of the table otherwise.
			target := ""
			if vpa.Spec.TargetRef != nil {
				target = fmt.Sprintf("%s/%s", vpa.Spec.TargetRef.Kind, vpa.Spec.TargetRef.Name)
			}
			reason, message := skipReason(vpa)
			skipped = append(skipped, SkippedVPA{
				Namespace: vpa.Namespace, Name: vpa.Name, Target: target,
				Reason: reason, Message: message,
			})
			continue
		}
		workload := ""
		if vpa.Spec.TargetRef != nil {
			workload = vpa.Spec.TargetRef.Name
		}
		for _, cr := range vpa.Status.Recommendation.ContainerRecommendations {
			row := RecommendationRow{
				Namespace: vpa.Namespace,
				Workload:  workload,
				Container: cr.ContainerName,
				Source:    sourceVPA,
			}
			if cr.LowerBound != nil {
				row.CPULowerBound = cr.LowerBound.Cpu().String()
				row.MemLowerBound = cr.LowerBound.Memory().String()
			}
			if cr.Target != nil {
				row.CPUTarget = cr.Target.Cpu().String()
				row.MemTarget = cr.Target.Memory().String()
			}
			if cr.UpperBound != nil {
				row.CPUUpperBound = cr.UpperBound.Cpu().String()
				row.MemUpperBound = cr.UpperBound.Memory().String()
			}
			rows = append(rows, row)
		}
	}
	return rows, skipped
}

func writeJSON(rows any, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rows); err != nil {
		return fmt.Errorf("writing JSON to %s: %w", path, err)
	}
	fmt.Printf("wrote %s\n", path)
	return nil
}

var csvHeader = []string{
	"namespace", "workload", "container", "source",
	"cpu_lower_bound", "cpu_target", "cpu_upper_bound",
	"memory_lower_bound", "memory_target", "memory_upper_bound",
}

func writeCSV(rows []RecommendationRow, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(csvHeader); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}
	for _, r := range rows {
		if err := w.Write([]string{
			r.Namespace, r.Workload, r.Container, r.Source,
			r.CPULowerBound, r.CPUTarget, r.CPUUpperBound,
			r.MemLowerBound, r.MemTarget, r.MemUpperBound,
		}); err != nil {
			return fmt.Errorf("writing CSV row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flushing CSV: %w", err)
	}
	fmt.Printf("wrote %s\n", path)
	return nil
}
