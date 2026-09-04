/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package container

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	v1 "k8s.io/api/core/v1"
)

func ValidateContainerIsReady(containerStatuses []v1.ContainerStatus, container string) error {
	log.Debug().Str("container", container).Msg("Starting container readiness validation")

	containerStatus, err := GetContainerStatus(containerStatuses, container)
	if err != nil {
		return err
	}

	if containerStatus.State.Running != nil {
		log.Debug().Str("container", container).Msg("Container is running")
		return nil
	}

	if containerStatus.State.Terminated != nil {
		log.Info().Str("container", container).
			Int32("exitCode", containerStatus.State.Terminated.ExitCode).
			Msg("Container is terminated")

		return api.NewNotReadyError(fmt.Sprintf(
			"container is terminated, exit code: %v", containerStatus.State.Terminated.ExitCode))
	}

	if containerStatus.State.Waiting != nil {
		log.Info().Str("container", container).Str("reason", containerStatus.State.Waiting.Reason).Msg("Container is waiting")
		return api.NewNotReadyError(fmt.Sprintf("container is waiting, reason: %v", containerStatus.State.Waiting.Reason))
	}

	log.Warn().Str("container", container).Msg("Container is not ready")

	return api.NewNotReadyError("container is not ready")
}

func ValidateContainerExist(containerStatuses []v1.ContainerStatus, container string) error {
	log.Debug().Str("container", container).Msg("Starting container exists validation")

	containerStatus, err := GetContainerStatus(containerStatuses, container)
	if err != nil {
		return err
	}

	if containerStatus.State.Running != nil || containerStatus.State.Terminated != nil {
		return nil
	}

	if containerStatus.State.Waiting != nil {
		log.Info().Str("container", container).Str("reason", containerStatus.State.Waiting.Reason).Msg("Container is waiting")
		return api.NewNotReadyError(fmt.Sprintf("container is waiting, reason: %v", containerStatus.State.Waiting.Reason))
	}

	log.Warn().Str("container", container).Msg("Container is not ready")

	return api.NewNotReadyError("container is not ready")
}

func GetContainerStatus(containerStatuses []v1.ContainerStatus, container string) (*v1.ContainerStatus, error) {
	for _, containerStatus := range containerStatuses {
		if containerStatus.Name == container {
			return &containerStatus, nil
		}
	}

	log.Error().Str("container", container).Msg("Container does not exist")

	return nil, api.NewNotFoundError("container does not exist")
}

func ValidateAttachable(container *v1.Container) error {
	if !container.TTY {
		return api.NewProtocolError(fmt.Sprintf("container %s, attach is not supported", container.Name))
	}

	return nil
}

func GetContainer(pod *v1.Pod, containerName string) (string, error) {
	log.Debug().Str("podName", pod.Name).Str("requestedContainer", containerName).Msg("Starting container selection")

	if containerName == "" {
		if len(pod.Spec.Containers) > 0 {
			container := pod.Spec.Containers[0].Name
			log.Debug().Str("container", container).Msg("Selected first container")

			return container, nil
		}

		if len(pod.Spec.InitContainers) > 0 {
			container := pod.Spec.InitContainers[0].Name
			log.Debug().Str("container", container).Msg("Selected first init container")

			return container, nil
		}

		if len(pod.Spec.EphemeralContainers) > 0 {
			container := pod.Spec.EphemeralContainers[0].Name
			log.Debug().Str("container", container).Msg("Selected first ephemeral container")

			return container, nil
		}

		log.Error().Msg("No containers found in pod")

		return "", api.NewNotFoundError("no containers found in pod")
	}

	log.Debug().Str("container", containerName).Msg("Using specified container")

	return containerName, nil
}
