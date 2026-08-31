/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/pods"
	"github.com/run-ai/runai/runai-cluster/cluster/common/getters"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SelectorFactoryFN func(workload *Workload) Selector

type Pods struct {
	k8sClient          client.Client
	selectorFn         SelectorFactoryFN
	fallbackSelectorFn SelectorFactoryFN
}

func NewPodsSelector(k8sClient client.Client) *Pods {
	log.Debug().Msg("Creating new Pods selector")

	return &Pods{
		k8sClient:          k8sClient,
		selectorFn:         GetSelectorCreator,
		fallbackSelectorFn: FallBackSelectorCreator,
	}
}

func (p *Pods) GetPod(ctx context.Context, workload *Workload, preferredPod string) (*corev1.Pod, error) {
	log.Debug().Str("workload", workload.Name).Str("preferredPod", preferredPod).Msg("Getting pod")

	workloadUID := getWorkloadUID(workload)
	if workloadUID == "" {
		log.Error().Msg("Workload UID is empty")
		return nil, apierrors.NewBadRequest("workload UID is empty")
	}

	pod, err := p.getPod(ctx, workload, preferredPod, workloadUID)
	if err != nil {
		if apierrors.IsNotFound(err) {
			if preferredPod != "" {
				// If the pod isn't found and the user requested a specific pod, return a NotFoundError
				log.Error().Str("podName", preferredPod).Msg("Specific pod not found")
				return nil, api.NewNotFoundError(err.Error())
			}
			// If the pod is not found, we will return a NotReadyError to indicate that the pod is not ready yet
			// but we don't want to return a 404 because the pod might be created later
			log.Warn().Msg("Pod not found, might be created later")

			return nil, api.NewNotReadyError("pod not found")
		}

		log.Error().Err(err).Msg("Failed to get pod")

		return nil, err
	}

	log.Debug().Str("podName", pod.Name).Msg("Pod retrieved successfully")

	return pod, nil
}

func (p *Pods) getPod(ctx context.Context, workload *Workload, preferredPod, workloadUID string) (*corev1.Pod, error) {
	if preferredPod != "" {
		return getters.GetPodOfWorkload(ctx, workloadUID, workload.Namespace, preferredPod, p.k8sClient)
	}

	selectorsFn := []SelectorFactoryFN{p.selectorFn, p.fallbackSelectorFn}
	for _, selectorFn := range selectorsFn {
		workloadSelector := selectorFn(workload)
		if workloadSelector == nil {
			log.Debug().Msg("Selector is nil")
			continue
		}

		selector := workloadSelector.CreateSelector()

		podsList, err := pods.GetPodsWithSelector(ctx, p.k8sClient, selector)
		if client.IgnoreNotFound(err) != nil {
			log.Error().Err(err).Msg("Failed to get pods of workload")
			return nil, err
		}

		if len(podsList) > 0 {
			log.Debug().Int("podCount", len(podsList)).Msg("Pods found for workload")
			return &podsList[0], nil
		}

		log.Debug().Msg("No pods found for workload")
	}

	return nil, apierrors.NewNotFound(corev1.Resource("Pod"), "no pods were found for requested workload")
}
