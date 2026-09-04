/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package attach

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/metrics"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket"
	"k8s.io/client-go/rest"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	workloadDomain "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/rs/zerolog/log"
)

var sessionId atomic.Uint64

type PodService interface {
	GenerateRequest(
		ctx context.Context, workload *workloadDomain.Workload, options stream.AttachOptions,
	) (*rest.Request, error)
	AttachRequest(ctx context.Context, req *rest.Request, handler *stream.Handler) error
}

type Handler struct {
	service  PodService
	upgrader websocket.Upgrader
	timeout  time.Duration
}

func NewHandler(service PodService, upgrader websocket.Upgrader, timeout time.Duration) *Handler {
	return &Handler{
		service:  service,
		upgrader: upgrader,
		timeout:  timeout,
	}
}

// Handle is a method of the Handler struct that handles HTTP requests.
// It starts by creating a context with a timeout using the Handler's timeout value.
// It then retrieves the workload from the context.
// If the workload is not found, it returns an internal server error.
// It attempts to create an Options struct from the HTTP request.
// If this fails, it returns a bad request error.
// It then calls the GenerateRequest method of the Handler's service with the context,
// workload, and options. If this fails, it returns an internal server error.
// It upgrades the HTTP connection to a websocket connection using the Handler's upgrader.
// If this fails, it returns an internal server error.
// It uses the service AttachRequest to handle the execute command to the pod via the K8S API.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Starting request handling")
	// Retrieve the workload from the context
	workload := workloadDomain.FromContext(r.Context())
	if workload == nil {
		// should never happen
		log.Error().Msg("Workload not found in context")
		http.Error(w, "workload not found in context", http.StatusInternalServerError)

		return
	}

	log.Info().Str("workloadName", workload.Name).
		Str("project", workload.Project).
		Str("workloadType", workload.Type.String()).
		Str("job framework", workload.JobType.String()).Msg("Handling attach request")

	log.Debug().Str("workloadName", workload.Name).Str("namespace", workload.Namespace).
		Msg("Retrieved workload from context")
	// Create a context with a cancel
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	options, err := stream.AttachOptionsFromRequest(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Options from request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		metrics.SetReasonIntoHeader(w, metrics.ErrOptionCreationFailed)

		return
	}

	log.Debug().Interface("options", options).Msg("Created Options from request")
	log.Debug().Msgf("creating attach request for workload:%s, ns:%s, pod:%s",
		workload.Name, workload.Namespace, options.Pod)

	request, err := h.service.GenerateRequest(ctx, workload, *options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get attach request")
		message, code := api.ToHttpError(err)
		http.Error(w, message, code)
		// limited to specific error strings
		metrics.SetReasonIntoHeader(w, metrics.ReasonType(message))

		return
	}

	defer log.Debug().Msgf("closing attach request for workload:%s, ns:%s, pod:%s",
		workload.Name, workload.Namespace, options.Pod)

	log.Debug().Msg("Successfully obtained attach request")

	// Upgrade the HTTP connection to a websocket connection
	conn, err := h.upgrader.Upgrade(ctx, w, r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade connection to websocket")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		metrics.SetReasonIntoHeader(w, metrics.ErrUpgradeConnection)

		return
	}

	log.Debug().Msg("Successfully upgraded connection to websocket")

	handler := stream.NewHandler(conn, options.Options)

	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close the websocket connection")
		}
	}()

	sessionId.Add(1)
	log.Info().Msgf("attaching request for workload:%s, ns:%s, pod:%s, session:%d",
		workload.Name, workload.Namespace, options.Pod, sessionId.Load())

	err = h.service.AttachRequest(ctx, request, handler)

	log.Info().Msgf("attach finished for workload:%s, ns:%s, pod:%s, session:%d",
		workload.Name, workload.Namespace, options.Pod, sessionId.Load())

	if err != nil {
		log.Error().Err(err).Msg("Failed to start proxy")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		metrics.SetReasonIntoHeader(w, metrics.ErrStartProxy)

		return
	}

	log.Debug().Msg("Attach was successful")

	metrics.SetReasonIntoHeader(w, metrics.MessageOK)
	w.WriteHeader(http.StatusOK)
	log.Info().Msg("Request handled successfully")
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling GET request")
	h.Handle(w, r)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling POST request")
	h.Handle(w, r)
}
