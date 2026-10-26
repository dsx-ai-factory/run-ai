/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	fields "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
)

// RunnerConverter projects a WorkloadRunner onto the unified JobWorkload shape
// the runtime-rules enforcement pipeline operates on. Two pieces of submission-time
// metadata — auto-deletion delay and GPU presence — are not carried on the
// WorkloadRunner itself; they live as annotations on the workload manifest
// embedded in Spec.Template, which we read directly (no cluster lookup of the
// materialized child).
type RunnerConverter struct {
	Client client.Client
}

func NewRunnerConverter(c client.Client) *RunnerConverter {
	return &RunnerConverter{Client: c}
}

// Convert populates jw from runner. The Convert method never fails because
// of a missing or unreachable wrapped child — those just leave the affected
// fields unset, which is the parity behavior with a legacy CRD that hasn't
// set those fields.
func (rc *RunnerConverter) Convert(runner *dromosv1alpha1.WorkloadRunner, jw *jobs.JobWorkload) error {
	jw.TypeMeta = v1.TypeMeta{
		Kind:       commonconstants.KindWorkloadRunner,
		APIVersion: dromosv1alpha1.GroupVersion.String(),
	}
	jw.ObjectMeta = *runner.ObjectMeta.DeepCopy()

	jw.Spec.Active = fields.NewOptionalBoolField(ptr.To(isRunnerActivePhase(runner.Status.Phase)))

	// Phase passthrough — the runtime-rules downstream code (e.g.
	// common.IsFinitePhase used by the auto-delete handler) reads this string.
	// The WorkloadRunner phase enum uses the same string values as the legacy
	// JobWorkload phase strings ("Completed" / "Failed" / ...).
	jw.Status.WorkloadCommonStatus.Phase = string(runner.Status.Phase)

	jw.Status.StartTime = resolveStartTime(runner)
	jw.Status.CompletionTime = parseAnnotationTime(runner.Annotations, commonconstants.CompletionTimeAnnotation)

	jw.Status.OwnedResources = []common.ObservedResource{
		{
			Resource: common.Resource{
				Kind: v1.GroupVersionKind{
					Group:   dromosv1alpha1.GroupVersion.Group,
					Version: dromosv1alpha1.GroupVersion.Version,
					Kind:    dromosv1alpha1.WorkloadRunnerKind,
				},
				Name:      runner.Name,
				Namespace: runner.Namespace,
			},
			Uid: runner.UID,
		},
	}

	template, err := parseTemplate(runner.Spec.Template.Raw)
	if err != nil {
		zap.S().Warnf("runner %s/%s: failed to parse spec.template (%v); skipping GPU + auto-delete population", runner.Namespace, runner.Name, err)
		return nil
	}

	if templateHasGpu(template) {
		jw.Spec.GpuDevices = fields.NewOptionalIntField(ptr.To(1))
	}

	if seconds := templateAutoDeletionSeconds(template); seconds != nil {
		jw.Spec.AutoDeletionTimeAfterCompletionSeconds = fields.NewOptionalIntField(new(int(*seconds)))
	}

	return nil
}

// resolveStartTime prefers the runtime-rules-maintained runai/start-time
// annotation (stamped on Running/Degraded transitions). It falls back to the
// runner's CreationTimestamp so a runner observed before its first Running
// transition still gets a non-zero StartTime — which the validation gate in
// GetWorkloadValues requires.
func resolveStartTime(runner *dromosv1alpha1.WorkloadRunner) *v1.Time {
	if t := parseAnnotationTime(runner.Annotations, commonconstants.StartTimeAnnotation); t != nil {
		return t
	}
	if !runner.CreationTimestamp.IsZero() {
		return runner.CreationTimestamp.DeepCopy()
	}
	return nil
}

func parseAnnotationTime(annotations map[string]string, key string) *v1.Time {
	raw, ok := annotations[key]
	if !ok || raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		zap.S().Warnf("annotation %s has invalid RFC3339 value %q: %v", key, raw, err)
		return nil
	}
	return &v1.Time{Time: parsed}
}

// parseTemplate decodes the workload manifest embedded in a WorkloadRunner's
// Spec.Template into an Unstructured. The GPU and auto-deletion metadata the
// runtime-rules needs are stamped as annotations on this template by the
// control-plane at submission, so there's no need to read the materialized
// child from the cluster.
func parseTemplate(raw []byte) (*unstructured.Unstructured, error) {
	if len(raw) == 0 {
		return nil, errors.New("template is empty")
	}
	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(raw, &obj.Object); err != nil {
		return nil, fmt.Errorf("unmarshal template: %w", err)
	}
	return obj, nil
}

// templateHasGpu detects GPU presence on the embedded template. Fractional
// GPUs are stamped as annotations on the pod template metadata (e.g.
// spec.template.metadata.annotations for a Job), not on the top-level object;
// whole GPUs carry no such annotation and appear only as a resource request in
// the pod spec. The pod template sits at a different path per Kind, so both
// signals are found by a single Kind-agnostic walk.
func templateHasGpu(obj *unstructured.Unstructured) bool {
	return templateRequestsGpu(obj.Object)
}

// partialGpuAnnotations mark a fractional or GPU-memory request.
var partialGpuAnnotations = []string{
	commonconstants.AnnotationGpuFraction,
	commonconstants.AnnotationGpuFractionDevices,
	commonconstants.AnnotationGpuMemory,
}

// gpuResourceNames are whole-GPU resource requests that carry no partial annotation.
var gpuResourceNames = []string{
	commonconstants.NvidiaGpuResourceName,
	commonconstants.AmdGpuResourceName,
	commonconstants.GpuResourceName,
}

// templateRequestsGpu walks the template for a partial-GPU annotation in any
// metadata.annotations map, or a positive GPU resource request (e.g.
// nvidia.com/gpu) in any container resource map.
func templateRequestsGpu(node any) bool {
	switch v := node.(type) {
	case map[string]any:
		if hasPartialGpuAnnotation(v["annotations"]) {
			return true
		}
		for _, name := range gpuResourceNames {
			if isPositiveQuantity(v[name]) {
				return true
			}
		}
		for _, child := range v {
			if templateRequestsGpu(child) {
				return true
			}
		}
	case []any:
		for _, child := range v {
			if templateRequestsGpu(child) {
				return true
			}
		}
	}
	return false
}

func hasPartialGpuAnnotation(val any) bool {
	annotations, ok := val.(map[string]any)
	if !ok {
		return false
	}
	for _, key := range partialGpuAnnotations {
		if s, ok := annotations[key].(string); ok && s != "" {
			return true
		}
	}
	return false
}

func isPositiveQuantity(val any) bool {
	switch q := val.(type) {
	case string:
		if quantity, err := resource.ParseQuantity(q); err == nil {
			return quantity.Sign() > 0
		}
	case int64:
		return q > 0
	case float64:
		return q > 0
	}
	return false
}

// templateAutoDeletionSeconds reads the auto-deletion delay (in seconds) the
// control-plane stamps on the embedded template at submission. Missing, empty,
// or non-numeric → nil (no auto-delete), matching the legacy unset case.
func templateAutoDeletionSeconds(obj *unstructured.Unstructured) *int32 {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return nil
	}
	raw, ok := annotations[commonconstants.AutoDeletionTimeAfterCompletionAnnotation]
	if !ok || raw == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(raw, 10, 32) //nolint:staticcheck
	if err != nil {
		zap.S().Warnf("annotation %s has invalid int value %q: %v", commonconstants.AutoDeletionTimeAfterCompletionAnnotation, raw, err)
		return nil
	}
	return new(int32(parsed))
}
