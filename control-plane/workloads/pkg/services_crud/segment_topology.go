/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

type segmentTopologyConfiguration struct {
	preferred    string
	preferredSet bool
	required     string
	requiredSet  bool
}

// segmentResponseSanitizationState preserves the effective native annotations
// before topology enrichment. The submitted manifest needs the enriched KAI
// annotations, while the native response and snapshot must expose only the
// caller's intent plus the canonical typed segmentSize field.
type segmentResponseSanitizationState struct {
	workerAnnotations []workloadsApi.Annotation
	masterAnnotations []workloadsApi.Annotation
	hasMaster         bool
}

func captureSegmentResponseSanitizationState(
	spec *workloadsApi.SupersetSpec,
	masterSpec *workloadsApi.MasterSpec,
) *segmentResponseSanitizationState {
	if !hasDistributedSegmentIntent(spec) {
		return nil
	}

	state := &segmentResponseSanitizationState{
		workerAnnotations: cloneWorkloadAnnotations(spec.GetAnnotations()),
		hasMaster:         masterSpec != nil,
	}
	if masterSpec != nil {
		state.masterAnnotations = cloneWorkloadAnnotations(masterSpec.GetAnnotations())
	}
	return state
}

func sanitizeSegmentedNativeResponse(
	spec *workloadsApi.SupersetSpec,
	masterSpec *workloadsApi.MasterSpec,
	state *segmentResponseSanitizationState,
) {
	if spec == nil || state == nil {
		return
	}

	// Restoring the pre-enrichment annotation list removes the topology name
	// and segment preference that Run:ai derived for the manifest. Explicit
	// placement and unrelated annotations are retained. Segment size is always
	// removed from annotations because spec.segmentSize is its canonical native
	// representation.
	spec.SetAnnotations(withoutSegmentSizeAnnotation(state.workerAnnotations))
	if state.hasMaster && masterSpec != nil {
		masterSpec.SetAnnotations(withoutSegmentSizeAnnotation(state.masterAnnotations))
	}
}

func cloneWorkloadAnnotations(annotations []workloadsApi.Annotation) []workloadsApi.Annotation {
	if annotations == nil {
		return nil
	}
	return append([]workloadsApi.Annotation(nil), annotations...)
}

func withoutSegmentSizeAnnotation(annotations []workloadsApi.Annotation) []workloadsApi.Annotation {
	if annotations == nil {
		return nil
	}

	result := make([]workloadsApi.Annotation, 0, len(annotations))
	for _, annotation := range annotations {
		if annotation.GetName() == topology.KAISegmentSizeAnnotationKey {
			continue
		}
		result = append(result, annotation)
	}
	return result
}

func normalizeDistributedSegmentTopology(spec *workloadsApi.SupersetSpec) (*segmentTopologyConfiguration, *workloadsApi.Error) {
	if spec == nil {
		return nil, nil
	}

	annotationSize, annotationSizeProvided := topology.ActiveAnnotationValue(spec.GetAnnotations(), topology.KAISegmentSizeAnnotationKey)
	preferred, preferredProvided := topology.ActiveAnnotationValue(spec.GetAnnotations(), topology.KAISegmentTopologyPreferredAnnotationKey)
	required, requiredProvided := topology.ActiveAnnotationValue(spec.GetAnnotations(), topology.KAISegmentTopologyRequiredAnnotationKey)

	// An active segment-size annotation, including an empty value, is explicit
	// intent. An empty size is therefore validated as malformed below. Placement
	// annotations are optional, so an empty value means unset and is defaulted.
	preferredSet := preferredProvided && preferred != ""
	requiredSet := requiredProvided && required != ""

	typedSize, typedSizeSet := spec.GetSegmentSizeOk()
	typedSizeSet = typedSizeSet && typedSize != nil
	hasSegmentSize := typedSizeSet || annotationSizeProvided
	hasSegmentIntent := hasSegmentSize || preferredSet || requiredSet
	if !hasSegmentIntent {
		return nil, nil
	}

	if spec.GetDistributedFramework() != workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH {
		return nil, newSegmentTopologyError("segment topology-aware scheduling is supported only for PyTorch distributed workloads")
	}

	if !hasSegmentSize {
		return nil, newSegmentTopologyError(fmt.Sprintf(
			"%s or spec.segmentSize must be provided when segment placement is configured",
			topology.KAISegmentSizeAnnotationKey,
		))
	}

	var annotationSizeValue int64
	if annotationSizeProvided {
		var parseErr error
		annotationSizeValue, parseErr = topology.ParseSegmentSize(annotationSize)
		if parseErr != nil {
			return nil, newSegmentTopologyError(fmt.Sprintf(
				"%s %s",
				topology.KAISegmentSizeAnnotationKey,
				parseErr,
			))
		}
	}

	if typedSizeSet && annotationSizeProvided && int64(*typedSize) != annotationSizeValue {
		return nil, newSegmentTopologyError(fmt.Sprintf(
			"spec.segmentSize (%d) must match %s (%s)",
			*typedSize,
			topology.KAISegmentSizeAnnotationKey,
			annotationSize,
		))
	}

	segmentSize := annotationSizeValue
	if typedSizeSet {
		segmentSize = int64(*typedSize)
	}

	numWorkers := int64(1)
	if configuredWorkers, configured := spec.GetNumWorkersOk(); configured && configuredWorkers != nil {
		numWorkers = int64(*configuredWorkers)
	}

	if _, countErr := topology.SegmentCount(numWorkers, segmentSize); countErr != nil {
		return nil, newSegmentTopologyError(countErr.Error())
	}

	normalizedSize := int32(segmentSize)
	spec.SetSegmentSize(normalizedSize)
	setActiveAnnotation(spec, topology.KAISegmentSizeAnnotationKey, strconv.FormatInt(segmentSize, 10))

	return &segmentTopologyConfiguration{
		preferred:    preferred,
		preferredSet: preferredSet,
		required:     required,
		requiredSet:  requiredSet,
	}, nil
}

func validateAndApplySegmentTopology(
	spec *workloadsApi.SupersetSpec,
	configuration *segmentTopologyConfiguration,
	topologyName string,
	levels []string,
) *workloadsApi.Error {
	if configuration == nil {
		return nil
	}

	if requestedTopology, set := topology.ActiveAnnotationValue(spec.GetAnnotations(), topology.KAITopologyAnnotationKey); set && requestedTopology != "" && requestedTopology != topologyName {
		return newSegmentTopologyError(fmt.Sprintf(
			"%s %q does not match the topology %q assigned to the selected node pools",
			topology.KAITopologyAnnotationKey,
			requestedTopology,
			topologyName,
		))
	}

	levelIndexes := make(map[string]int, len(levels))
	for index, level := range levels {
		levelIndexes[level] = index
	}

	preferredIndex, err := validateSegmentPlacementLevel(
		topology.KAISegmentTopologyPreferredAnnotationKey,
		configuration.preferred,
		configuration.preferredSet,
		levelIndexes,
	)
	if err != nil {
		return err
	}
	requiredIndex, err := validateSegmentPlacementLevel(
		topology.KAISegmentTopologyRequiredAnnotationKey,
		configuration.required,
		configuration.requiredSet,
		levelIndexes,
	)
	if err != nil {
		return err
	}

	// Topology levels are ordered from broadest to narrowest. A preferred
	// placement may be equal to or narrower than the required placement, but
	// never broader.
	if configuration.preferredSet && configuration.requiredSet && preferredIndex < requiredIndex {
		return newSegmentTopologyError(fmt.Sprintf(
			"%s must be the same as or lower than %s in topology %q",
			topology.KAISegmentTopologyPreferredAnnotationKey,
			topology.KAISegmentTopologyRequiredAnnotationKey,
			topologyName,
		))
	}

	setActiveAnnotation(spec, topology.KAITopologyAnnotationKey, topologyName)
	if !configuration.preferredSet && !configuration.requiredSet {
		setActiveAnnotation(spec, topology.KAISegmentTopologyPreferredAnnotationKey, levels[len(levels)-1])
	}

	// Keep the workload-level default tied to the same selected node pools as
	// the segment default. The composer projects this placement onto the
	// PyTorchJob root while keeping segment placement on the Worker template.
	// An explicit workload preferred or required placement remains authoritative.
	workloadPreferred, workloadPreferredProvided := topology.ActiveAnnotationValue(
		spec.GetAnnotations(),
		topology.KAITopologyAnnotationPreferredKey,
	)
	workloadRequired, workloadRequiredProvided := topology.ActiveAnnotationValue(
		spec.GetAnnotations(),
		topology.KAITopologyAnnotationRequiredKey,
	)
	if (!workloadPreferredProvided || workloadPreferred == "") &&
		(!workloadRequiredProvided || workloadRequired == "") {
		setActiveAnnotation(spec, topology.KAITopologyAnnotationPreferredKey, levels[len(levels)-1])
	}
	return nil
}

func validateSegmentPlacementLevel(
	annotationName string,
	level string,
	set bool,
	levelIndexes map[string]int,
) (int, *workloadsApi.Error) {
	if !set {
		return 0, nil
	}

	index, exists := levelIndexes[level]
	if !exists {
		return 0, newSegmentTopologyError(fmt.Sprintf(
			"%s %q is not a level in the selected node-pool topology",
			annotationName,
			level,
		))
	}
	return index, nil
}

func validateSeparateMasterSegmentAnnotations(masterSpec *workloadsApi.MasterSpec, masterSameAsWorker bool) *workloadsApi.Error {
	if masterSameAsWorker || masterSpec == nil || !hasSegmentAnnotationIntent(masterSpec) {
		return nil
	}

	return newSegmentTopologyError("segment annotations must be configured on the worker spec, not on a separate master spec")
}

func hasSegmentAnnotationIntent(annotations Annotater) bool {
	if annotations == nil {
		return false
	}
	// Empty segment size is still explicit (malformed) intent. Empty optional
	// placement values mean unset and do not enable the feature.
	if _, provided := topology.ActiveAnnotationValue(annotations.GetAnnotations(), topology.KAISegmentSizeAnnotationKey); provided {
		return true
	}
	for _, name := range []string{
		topology.KAISegmentTopologyPreferredAnnotationKey,
		topology.KAISegmentTopologyRequiredAnnotationKey,
	} {
		if value, found := topology.ActiveAnnotationValue(annotations.GetAnnotations(), name); found && value != "" {
			return true
		}
	}
	return false
}

func hasDistributedSegmentIntent(spec *workloadsApi.SupersetSpec) bool {
	if spec == nil {
		return false
	}
	if segmentSize, set := spec.GetSegmentSizeOk(); set && segmentSize != nil {
		return true
	}
	return hasSegmentAnnotationIntent(spec)
}

func validateSegmentTopologyScope(
	workloadType string,
	spec *workloadsApi.SupersetSpec,
	masterSpec *workloadsApi.MasterSpec,
) *workloadsApi.Error {
	if !hasWorkloadSegmentIntent(spec, masterSpec) {
		return nil
	}
	if workloadType != Distributed || spec == nil || spec.GetDistributedFramework() != workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH {
		return newSegmentTopologyError("segment topology-aware scheduling is supported only for PyTorch distributed workloads")
	}
	return nil
}

func hasWorkloadSegmentIntent(spec *workloadsApi.SupersetSpec, masterSpec *workloadsApi.MasterSpec) bool {
	if hasDistributedSegmentIntent(spec) || hasSegmentIntentInAnnotaters(getSpecAnnotations(spec)) {
		return true
	}
	if worker, set := spec.GetWorkerOk(); set && worker != nil && hasSegmentAnnotationIntent(worker) {
		return true
	}
	return hasSegmentAnnotationIntent(masterSpec)
}

func hasSegmentIntentInAnnotaters(annotationList []Annotater) bool {
	for _, annotations := range annotationList {
		if hasSegmentAnnotationIntent(annotations) {
			return true
		}
	}
	return false
}

func newSegmentTopologyError(message string) *workloadsApi.Error {
	return workloadsApi.NewError(http.StatusBadRequest, message)
}
