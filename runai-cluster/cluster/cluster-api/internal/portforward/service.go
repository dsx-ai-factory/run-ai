/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package portforward

import (
	"context"

	clusterpod "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/pod"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/service"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
)

type Forwarder interface {
	Create(namespace string, options *Options) (httpstream.Connection, error)
}

type Service struct {
	podsGetter service.PodsGetter
	forward    Forwarder
}

func NewService(podsGetter service.PodsGetter, forward Forwarder) *Service {
	log.Debug().Msg("Creating new port forward service")

	return &Service{
		podsGetter: podsGetter,
		forward:    forward,
	}
}

func (s *Service) GetConnection(
	ctx context.Context, workload *workload.Workload, options Options,
) (httpstream.Connection, error) {
	log.Info().
		Str("workload", workload.Name).
		Str("namespace", workload.Namespace).
		Int32("port", options.Port).
		Str("pod", options.Pod).
		Msg("Getting connection for port forward")

	pod, err := s.podsGetter.GetPod(ctx, workload, options.Pod)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pod")
		return nil, err
	}

	builder := NewOptionsBuilder().FromOptions(options)
	if options.Pod == "" {
		builder.WithPod(pod.Name)
	}

	finalOption := builder.Build()

	err = validate(pod, finalOption)
	if err != nil {
		log.Error().Err(err).Msg("Validation failed")
		return nil, err
	}

	stream, err := s.createConnector(workload.Namespace, finalOption)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create connector")
		return nil, err
	}

	log.Debug().Msg("Connection for port forward created successfully")

	return stream, nil
}

func validate(pod *v1.Pod, options Options) error {
	log.Debug().Str("pod", pod.Name).Int32("port", options.Port).Msg("Validating pod and port")

	if err := clusterpod.ValidatePodRunning(pod); err != nil {
		return err
	}

	return nil
}

func (s *Service) createConnector(namespace string, options Options) (httpstream.Connection, error) {
	log.Debug().Str("namespace", namespace).Int32("port", options.Port).Str("pod", options.Pod).Msg("Creating connector")
	return s.forward.Create(namespace, &options)
}
