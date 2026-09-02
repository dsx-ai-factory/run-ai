/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package data

import (
	"context"

	"github.com/rs/zerolog/log"
	"go.uber.org/multierr"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/priority"
	v1 "k8s.io/api/scheduling/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	priorityInferenceNumber = 125
	priorityBuildNumber     = 100
	priorityTrainNumber     = 50
	defaultPriority         = priorityTrainNumber
)

type MetricsWorkloadType string

const (
	JobTypeTrain                  MetricsWorkloadType = "Train"
	JobTypeInteractive            MetricsWorkloadType = "Interactive"
	JobTypeInference              MetricsWorkloadType = "Inference"
	JobTypeDistributedInference   MetricsWorkloadType = "Distributed-Inference"
	JobTypeInteractivePreemptible MetricsWorkloadType = "Interactive-Preemptible"
)

func GetWorkloadTypeFromPodGroup(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, readerClient client.Reader) MetricsWorkloadType {
	podGroupPriority := getPodGroupPriority(ctx, podGroup, readerClient)

	workloadType := getWorkloadTypeFromPriority(podGroupPriority, len(podGroup.OwnerReferences) > 0 && podGroup.OwnerReferences[0].Kind == "Deployment")
	if workloadType == "" {
		log.Error().Msgf("unexpected priority class: %s (%d) for pod group <%s/%s>; will use default Train workload type",
			podGroup.Spec.PriorityClassName, podGroupPriority, podGroup.Namespace, podGroup.Name)
		return JobTypeTrain
	}

	return workloadType
}

func GetWorkloadTypeFromPriorityClassName(ctx context.Context, priorityClassName, workloadNamespace, workloadName string,
	readerClient client.Reader) MetricsWorkloadType {
	workloadPriority := getPriorityOfPriorityClassName(ctx, priorityClassName, readerClient)

	workloadType := getWorkloadTypeFromPriority(workloadPriority, false)
	if workloadType == "" {
		log.Error().Msgf("unexpected priority class: %s (%d) for workload <%s/%s>; will use default Train workload type",
			priorityClassName, workloadPriority, workloadNamespace, workloadName)
		return JobTypeTrain
	}

	return workloadType
}

func getWorkloadTypeFromPriority(priority int32, hasDeploymentOwner bool) MetricsWorkloadType {
	// if priority > 125 : Inference
	// else if priority >= 100 Inference or Interactive
	// else if priority > 50 : InteractivePreemptible
	// else Train

	if priority > priorityInferenceNumber {
		return JobTypeInference
	}
	if priority >= priorityBuildNumber {
		if hasDeploymentOwner {
			// For backward compatibility
			return JobTypeInference
		}
		return JobTypeInteractive
	}
	if priority > priorityTrainNumber {
		return JobTypeInteractivePreemptible
	}
	return JobTypeTrain
}

func getPodGroupPriority(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, readerClient client.Reader) int32 {
	return getPriorityOfPriorityClassName(ctx, podGroup.Spec.PriorityClassName, readerClient)
}

func getPriorityOfPriorityClassName(ctx context.Context, priorityClassName string, readerClient client.Reader) int32 {
	var err error

	if priorityClassName != "" {
		var priorityClass *v1.PriorityClass
		priorityClass, err = priority.GetPriorityClass(ctx, priorityClassName, readerClient)
		if err == nil && priorityClass != nil {
			return priorityClass.Value
		}
	}

	clusterDefaultPriority, innerErr := priority.GetDefaultPriority(ctx, readerClient)
	if innerErr != nil {
		err = multierr.Append(err, innerErr)
		clusterDefaultPriority = defaultPriority
	}

	log.Error().Msgf("Couldn't find priority of priorityClassName: %s, error: %v. Using cluster default priority %d",
		priorityClassName, err, clusterDefaultPriority)
	return clusterDefaultPriority
}
