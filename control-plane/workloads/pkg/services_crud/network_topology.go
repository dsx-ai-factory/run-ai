/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"net/http"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/run-ai/runai/runai-application-packages/datastructs"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

const ErrNodePoolsDiff = "same node pool names are expected in both worker and master"

type Annotater interface {
	GetAnnotations() []workloadsApi.Annotation
	SetAnnotations(annotations []workloadsApi.Annotation)
}

type TopologyMode string

func handleNodePoolsAndTopology(
	ctx context.Context,
	workloadType string,
	actualNodePools []clusterApi.Nodepool,
	spec *workloadsApi.SupersetSpec,
	masterSpec *workloadsApi.MasterSpec,
	masterSameAsWorker bool,
	segmentRouting segmentedDistributedRoutingDecision,
) *workloadsApi.Error {
	if spec == nil {
		return nil
	}

	annotationList := getSpecAnnotations(spec)
	if err := validateSegmentTopologyScope(workloadType, spec, masterSpec); err != nil {
		return err
	}

	switch workloadType {
	case DistributedInference:
		return updateWorkloadTopologyAnnotations(ctx, actualNodePools, spec, annotationList)
	case Distributed:
		if err := validateSeparateMasterSegmentAnnotations(masterSpec, masterSameAsWorker); err != nil {
			return err
		}
		if masterSpec != nil {
			err := compareNodePools(spec.NodePools, masterSpec.NodePools)
			if err != nil {
				return err
			}
		}
		return updateWorkloadTopologyAnnotationsForRouting(ctx, actualNodePools, spec, annotationList, segmentRouting)
	default:
		// topology is not relevant, but we have to check the list of node pools itself
		_, err := topology.CheckNodePoolsNames(ctx, actualNodePools, spec.NodePools)
		return err
	}
}

func compareNodePools(worker []string, master []string) *workloadsApi.Error {
	if len(master) == 0 {
		return nil // not specifying master node-pools is what we actually hope for
	}

	if len(worker) == 0 {
		if len(master) != 0 {
			return workloadsApi.NewError(http.StatusBadRequest, ErrNodePoolsDiff)
		}
		return nil // both zero we are also ok
	}

	// both specified, we expect the lists to contain same element (not necessarily in the same order)
	workerPools := datastructs.NewSet(worker...)
	masterPools := datastructs.NewSet(master...)

	if !workerPools.EqualTo(masterPools) {
		return workloadsApi.NewError(http.StatusBadRequest, ErrNodePoolsDiff)
	}

	return nil
}

func getSpecAnnotations(spec *workloadsApi.SupersetSpec) []Annotater {
	var annotationList []Annotater
	if spec != nil {
		annotationList = append(annotationList, spec)
		if spec.HasLeader() {
			annotationList = append(annotationList, spec.Leader.Get())
		}
	}
	return annotationList
}

func updateWorkloadTopologyAnnotations(ctx context.Context, actualNodePools []clusterApi.Nodepool, spec *workloadsApi.SupersetSpec, annotationList []Annotater) *workloadsApi.Error {
	return updateWorkloadTopologyAnnotationsForRouting(
		ctx,
		actualNodePools,
		spec,
		annotationList,
		segmentedDistributedRoutingDecision{},
	)
}

func updateWorkloadTopologyAnnotationsForRouting(
	ctx context.Context,
	actualNodePools []clusterApi.Nodepool,
	spec *workloadsApi.SupersetSpec,
	annotationList []Annotater,
	segmentRouting segmentedDistributedRoutingDecision,
) *workloadsApi.Error {
	segmentConfiguration, err := normalizeDistributedSegmentTopology(spec)
	if err != nil {
		return err
	}
	if segmentConfiguration != nil && (!segmentRouting.validated || segmentRouting.err != nil) {
		if segmentRouting.err != nil {
			return segmentRouting.err
		}
		return workloadsApi.NewError(
			http.StatusServiceUnavailable,
			"failed to validate cluster support for segment topology-aware scheduling",
		)
	}

	found, topologyName, err := topology.CheckNodePoolsTopology(ctx, actualNodePools, spec.NodePools)
	if err != nil {
		return err
	}

	if !found {
		if segmentConfiguration != nil {
			return newSegmentTopologyError("segment topology-aware scheduling requires the selected node pools to have a network topology")
		}
		return nil
	}

	networkTopology, err := topology.GetByName(ctx, clusterservice.Get(), topologyName)
	if err != nil {
		return err
	}

	if len(networkTopology.Levels) == 0 {
		return topology.ErrTopologyHasNoLevels
	}

	lowestTopologyLevel := networkTopology.Levels[len(networkTopology.Levels)-1]
	if segmentConfiguration != nil {
		// Resolve workload and segment defaults from the same selected node-pool
		// topology. The composer projects workload placement onto the PyTorchJob
		// and keeps segment placement on the Worker PodTemplate.
		return validateAndApplySegmentTopology(spec, segmentConfiguration, topologyName, networkTopology.Levels)
	}

	for _, annotations := range annotationList {
		setAnnotation(annotations, topology.KAITopologyAnnotationKey, topologyName)
		setAnnotation(annotations, topology.KAITopologyAnnotationPreferredKey, lowestTopologyLevel)
	}
	return nil
}

func setAnnotation(annotations Annotater, name, value string) {
	annotationList := annotations.GetAnnotations()
	// annotation is empty
	if annotationList == nil {
		annotations.SetAnnotations([]workloadsApi.Annotation{{Name: *workloadsApi.NewNullableString(&name), Value: *workloadsApi.NewNullableString(&value)}})
		return
	}

	for index := range annotationList {
		// annotation already exist
		if annotationList[index].GetName() == name {
			annotationList[index].SetValue(value)
			annotations.SetAnnotations(annotationList)
			return
		}
	}

	// annotation missing adding
	annotationList = append(annotationList, workloadsApi.Annotation{Name: *workloadsApi.NewNullableString(&name), Value: *workloadsApi.NewNullableString(&value)})
	annotations.SetAnnotations(annotationList)
}

// setActiveAnnotation follows annotation composition's last-entry-wins
// behavior and makes a system-derived annotation active even when the last
// matching policy annotation was explicitly excluded.
func setActiveAnnotation(annotations Annotater, name, value string) {
	annotationList := annotations.GetAnnotations()
	lastMatchingIndex := -1
	for index := range annotationList {
		if annotationList[index].GetName() == name {
			lastMatchingIndex = index
		}
	}

	if lastMatchingIndex >= 0 {
		annotationList[lastMatchingIndex].SetValue(value)
		annotationList[lastMatchingIndex].SetExclude(false)
		annotations.SetAnnotations(annotationList)
		return
	}

	annotationList = append(annotationList, workloadsApi.Annotation{
		Name:  *workloadsApi.NewNullableString(&name),
		Value: *workloadsApi.NewNullableString(&value),
	})
	annotations.SetAnnotations(annotationList)
}
