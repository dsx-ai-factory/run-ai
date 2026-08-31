/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package portforward

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/meta"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/metrics"
	"k8s.io/apimachinery/pkg/util/httpstream"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	workloadDomain "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
)

var sessionId atomic.Uint64

type PodPFService interface {
	GetConnection(ctx context.Context, workload *workloadDomain.Workload, options Options) (httpstream.Connection, error)
}

type Handler struct {
	service  PodPFService
	upgrader websocket.Upgrader
	timeout  time.Duration
}

func NewHandler(service PodPFService, upgrader websocket.Upgrader, timeout time.Duration) *Handler {
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
// It then calls the LogsReader method of the Handler's service with the context, workload,
// and options. If this fails, it returns an internal server error.
// It upgrades the HTTP connection to a websocket connection using the Handler's upgrader.
// If this fails, it returns an internal server error.
// It creates a proxy using the Handler's proxyFactoryFn with the logs reader and
// the websocket connection.
// It starts the proxy with the context. If this fails, it returns an internal server error.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
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
		Str("job framework", workload.JobType.String()).Msg("Handling port forward request")
	// Create a context with a timeout
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	// Create an Options struct from the HTTP request
	options, err := FromGetRequest(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create options from request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		metrics.SetReasonIntoHeader(w, metrics.ErrOptionCreationFailed)

		return
	}

	// Call the GetConnection method of the Handler's service
	dial, err := h.service.GetConnection(ctx, workload, options)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get portforward connection")
		message, code := api.ToHttpError(err)
		http.Error(w, message, code)
		metrics.SetReasonIntoHeader(w, metrics.ErrGenerateRequest)

		return
	}

	// Upgrade the HTTP connection to a websocket connection
	conn, err := h.upgrader.Upgrade(ctx, w, r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade connection to websocket")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		metrics.SetReasonIntoHeader(w, metrics.ErrUpgradeConnection)

		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close the connection")
		}
	}()

	log.Info().Msg("Starting connection tunneling")

	// Handle the connection tunneling
	sessionId.Add(1)

	log.Info().Msgf("staring port-forward for workload:%s, ns:%s, pod:%s, port:%d, session:%d",
		workload.Name, workload.Namespace, options.Pod, options.Port, sessionId.Load())
	defer log.Info().Msgf("ending port-forward for workload:%s, ns:%s, pod:%s, port:%d, session:%d",
		workload.Name, workload.Namespace, options.Pod, options.Port, sessionId.Load())

	stream := NewStream(dial, options.Port)

	if meta.GetFeatures().IsMessageEnabled() {
		websocket.SetClientMessageHandler(conn)
	}

	err = stream.TunnelConnection(ctx, conn, options.Port)
	if err != nil {
		log.Error().Err(err).Msg("Failed to tunnel connection")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		metrics.SetReasonIntoHeader(w, metrics.ErrStartProxy)

		return
	}

	log.Info().Msg("Port forwarding completed successfully")
	metrics.SetReasonIntoHeader(w, metrics.MessageOK)
	// return 200
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling GET request for port forwarding")
	h.Handle(w, r)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling POST request for port forwarding")
	h.Handle(w, r)
}
