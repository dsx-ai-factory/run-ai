/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/statusmsg"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	clustercontainer "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/container"
	clusterpod "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/pod"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/service"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type Attacher interface {
	CreateRequest(namespace string, options *stream.AttachOptions) (*rest.Request, error)
}

type Service struct {
	podsGetter service.PodsGetter
	podAttach  Attacher
	config     *rest.Config
}

func NewService(podsGetter service.PodsGetter, podAttach Attacher, config *rest.Config) *Service {
	log.Debug().Msg("Creating new Service")

	s := &Service{
		podsGetter: podsGetter,
		podAttach:  podAttach,
		config:     config,
	}

	log.Info().Msg("Attach service created successfully")

	return s
}

func (s *Service) GenerateRequest(
	ctx context.Context, workload *workload.Workload, options stream.AttachOptions,
) (*rest.Request, error) {
	log.Debug().Str("workloadName", workload.Name).Str("namespace", workload.Namespace).
		Str("podName", options.Pod).Msg("Starting - Getting pod")

	pod, err := s.podsGetter.GetPod(ctx, workload, options.Pod)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pod")
		return nil, err
	}

	container, err := clustercontainer.GetContainer(pod, options.Container)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get container")
		return nil, err
	}

	log.Debug().Str("container", container).Msg("Container selected")

	options.Container = container
	options.Pod = pod.Name

	err = validate(pod, options)
	if err != nil {
		log.Error().Err(err).Msg("Validation failed")
		return nil, err
	}

	request, err := s.getRequest(workload.Namespace, options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get attach request")
		return nil, err
	}

	log.Debug().Msg("Successfully obtained attach request")

	return request, nil
}

func (s *Service) AttachRequest(ctx context.Context, req *rest.Request, handler *stream.Handler) error {
	log.Info().Msg("AttachRequest, status=start")
	defer log.Info().Msg("AttachRequest, status=end")

	keepalive, err := handler.GetKeepAlive()
	if err != nil {
		return err
	}

	log.Info().Msgf("session keepalive type: %s", keepalive.GetType())

	attach, err := remotecommand.NewSPDYExecutor(s.config, http.MethodPost, req.URL())
	if err != nil {
		return err
	}

	so := handler.GetStreamOptions()
	defer handler.Close()

	errChan := make(chan error)

	defer close(errChan)

	var wg sync.WaitGroup

	wg.Add(1)
	// Start a goroutine to copy data from 'from' to 'to'
	go func() {
		defer wg.Done()

		sErr := attach.StreamWithContext(ctx, so)
		if sErr != nil {
			log.Info().Msgf("steam content finished with error: %v", sErr)

			localErr := handler.WriteStatusMsg(statusmsg.GenerateStatus(sErr))
			if localErr != nil {
				log.Warn().Msgf("failed to send error to client: %v", localErr)
			}
		}

		log.Info().Msgf("steam content ended")

		errChan <- sErr
	}()

	defer wg.Wait()

	for {
		select {
		case <-keepalive.Channel():
			iErr := keepalive.Execute()
			if iErr != nil {
				log.Info().Msgf("keepalive limit reached")
				return nil
			}
		case <-errChan:
			log.Info().Msgf("stream ended - keepalive no longer needed")
			return nil
		}
	}
}

func validate(pod *v1.Pod, options stream.AttachOptions) error {
	log.Debug().Str("podName", pod.Name).Msg("Starting validation")

	if err := clusterpod.ValidatePodNotPending(pod); err != nil {
		log.Error().Err(err).Msg("Pod validation failed")
		return err
	}

	if err := ValidateContainer(pod, options); err != nil {
		log.Error().Err(err).Msg("Container validation failed")
		return err
	}

	log.Debug().Msg("Validation successful")

	return nil
}

func (s *Service) getRequest(namespace string, options stream.AttachOptions) (*rest.Request, error) {
	log.Debug().Str("namespace", namespace).Interface("options", options).Msg("Starting to stream attach")

	request, err := s.podAttach.CreateRequest(namespace, &options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to stream attachment")
		return nil, err
	}

	log.Debug().Msg("Successfully started streaming attachment")

	return request, nil
}

func ValidateContainer(pod *v1.Pod, options stream.AttachOptions) error {
	log.Debug().Str("podName", pod.Name).Str("container", options.Container).Msg("Starting container validation")
	// If the user requested a specific container, we will validate that the container exists in the pod
	if options.Container == "" {
		log.Error().Msg("container name empty, should have reached this point...")
		// should not happen
		return fmt.Errorf("container name is required")
	}

	for _, container := range pod.Spec.Containers {
		if container.Name == options.Container {
			err := clustercontainer.ValidateAttachable(&container)
			if err != nil {
				return err
			}

			return clustercontainer.ValidateContainerIsReady(pod.Status.ContainerStatuses, container.Name)
		}
	}

	log.Error().Str("container", options.Container).Msg("Requested container not found")

	return api.NewNotFoundError("the requested container not found")
}
