/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	distinfwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	infwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const unknownTypeErrorMsg = "unknown workload type %s"

func GetK8sResource(workloadType api.WorkloadType) (client.Object, error) {
	log.Debug().Str("workloadType", string(workloadType)).Msg("Starting")

	var resource client.Object

	switch workloadType {
	case api.Workspace:
		log.Debug().Msg("Creating InteractiveWorkload")

		resource = &iwv2.InteractiveWorkload{}
	case api.Training:
		log.Debug().Msg("Creating TrainingWorkload")

		resource = &twv2.TrainingWorkload{}
	case api.Distributed:
		log.Debug().Msg("Creating DistributedWorkload")

		resource = &dwv2.DistributedWorkload{}
	case api.Inference:
		log.Debug().Msg("Creating InferenceWorkload")

		resource = &infwv2.InferenceWorkload{}
	case api.DistributedInference:
		log.Debug().Msg("Creating DistributedInferenceWorkload")

		resource = &distinfwv2.DistributedInferenceWorkload{}
	case api.External:
		log.Debug().Msg("Creating ExternalWorkload")

		resource = &v1alpha1.ExternalWorkload{}
	default:
		err := fmt.Errorf(unknownTypeErrorMsg, workloadType)
		log.Error().Err(err).Str("workloadType", string(workloadType)).Msg("Unknown workload type")

		return nil, err
	}

	log.Debug().Str("resourceType", fmt.Sprintf("%T", resource)).Msg("Resource created successfully")

	return resource, nil
}

type K8sLocator struct {
	k8sClient client.Client
}

func NewK8sLocator(k8sClient client.Client) *K8sLocator {
	log.Debug().Msg("NewK8sLocator: Creating new K8sLocator")

	locator := &K8sLocator{
		k8sClient: k8sClient,
	}

	log.Debug().Interface("locator", locator).Msg("NewK8sLocator: K8sLocator created")

	return locator
}

func (kl *K8sLocator) Locate(
	ctx context.Context, workloadType api.WorkloadType, workloadName, namespace string,
) (client.Object, error) {
	log.Debug().
		Str("workloadType", workloadType.String()).
		Str("workloadName", workloadName).
		Str("namespace", namespace).
		Msg("Starting")

	resource, err := GetK8sResource(workloadType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get K8s resource")
		return nil, err
	}

	log.Debug().Str("resourceType", fmt.Sprintf("%T", resource)).Msg("K8s resource obtained")

	resource.SetName(workloadName)
	resource.SetNamespace(namespace)
	log.Debug().
		Str("workloadName", workloadName).
		Str("namespace", namespace).
		Msg("Name and namespace set on resource")

	log.Debug().Msg("Attempting to get resource from K8s client")

	err = kl.k8sClient.Get(ctx,
		client.ObjectKey{Namespace: namespace, Name: workloadName}, resource)
	if err != nil {
		log.Error().Err(err).
			Str("workloadName", workloadName).
			Str("namespace", namespace).
			Msg("Failed to get K8s resource from client")

		return nil, err
	}

	log.Info().
		Str("workloadType", workloadType.String()).
		Str("workloadName", workloadName).
		Str("workloadUID", string(resource.GetUID())).
		Str("namespace", namespace).
		Str("resourceType", fmt.Sprintf("%T", resource)).
		Msg("Successfully located K8s resource")

	return resource, nil
}
