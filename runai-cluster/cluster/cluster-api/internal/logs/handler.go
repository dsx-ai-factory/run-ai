/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/meta"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/response"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"

	"github.com/rs/zerolog/log"
	workloadDomain "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
)

const (
	FailedToParseWorkloadOrOptions = "Failed to parse workload or options request"
	FailedToGetLogsReader          = "Failed to get logs reader"
	failedToUpgradeConnection      = "Failed to upgrade connection to websocket"
	failedToStartProxy             = "Failed to start proxy"
)

var sessionId atomic.Uint64

type PodLogsService interface {
	LogsReader(ctx context.Context, workload *workloadDomain.Workload, options *Options) (io.ReadCloser, error)
}

type Handler struct {
	service        PodLogsService
	upgrader       websocket.Upgrader
	proxyFactoryFn ProxyFactoryFn
	timeout        time.Duration
	config         *stream.Config
}

func NewHandler(
	service PodLogsService, upgrader websocket.Upgrader, proxyFactoryFn ProxyFactoryFn, timeout time.Duration,
) *Handler {
	if proxyFactoryFn == nil {
		proxyFactoryFn = NewOneWayProxyFactory()
	}

	return &Handler{
		service:        service,
		upgrader:       upgrader,
		proxyFactoryFn: proxyFactoryFn,
		timeout:        timeout,
		config:         stream.GetInstance(),
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
	// Create a context with cancel
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	workload, options, err := GetWorkloadAndOptions(w, r)
	if err != nil {
		response.ErrorBadRequestWithSetReason(w, FailedToParseWorkloadOrOptions, metrics.ErrGetWorkloadAndOptions, err)
		return
	}

	// Call the LogsReader method of the Handler's service
	log.Debug().Msgf("creating log reader stream for workload:%s, ns:%s, pod:%s",
		workload.Name, workload.Namespace, options.Pod)

	logsReader, err := h.service.LogsReader(ctx, workload, options)
	if err != nil {
		response.ErrorWithReason(w, FailedToGetLogsReader, err)
		return
	}

	defer func() {
		log.Debug().Msgf("closing log reader stream for workload:%s, ns:%s, pod:%s",
			workload.Name, workload.Namespace, options.Pod)

		_ = logsReader.Close()
	}()

	log.Debug().Msg("Successfully obtained logs reader")

	// Upgrade the HTTP connection to a websocket connection
	conn, err := h.upgrader.Upgrade(ctx, w, r)
	if err != nil {
		response.ErrorInternalServerWithSetReason(w, failedToUpgradeConnection, metrics.ErrUpgradeConnection, err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			log.Error().Err(err).Msg("Failed to close the websocket connection")
		}
	}()

	log.Debug().Msg("Successfully upgraded connection to websocket")

	sessionId.Add(1)
	log.Info().Msgf("staring log reader proxy for workload:%s, ns:%s, pod:%s, session:%d",
		workload.Name, workload.Namespace, options.Pod, sessionId.Load())

	if meta.GetFeatures().IsMessageEnabled() {
		websocket.SetClientMessageHandler(conn)
	}

	pingKeepAlive := keepalive.NewPing(conn, logsReader, h.config.GetPingTime())
	proxy := h.proxyFactoryFn(logsReader, conn, pingKeepAlive)

	log.Debug().Msg("Created proxy")

	err = proxy.Start(ctx)

	log.Info().Msgf("ending log reader proxy for workload:%s, ns:%s, pod:%s, session:%d",
		workload.Name, workload.Namespace, options.Pod, sessionId.Load())

	if err != nil {
		response.ErrorInternalServerWithSetReason(w, failedToStartProxy, metrics.ErrStartProxy, err)
		return
	}

	log.Debug().Msg("Successfully started proxy")

	response.OK(w)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling GET request")
	h.Handle(w, r)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling POST request")
	h.Handle(w, r)
}

func GetWorkloadAndOptions(w http.ResponseWriter, r *http.Request) (*workloadDomain.Workload, *Options, error) {
	log.Debug().Msg("Starting request handling")
	// Retrieve the workload from the context
	workload := workloadDomain.FromContext(r.Context())
	if workload == nil {
		// should never happen
		log.Error().Msg("Workload not found in context")
		http.Error(w, "workload not found in context", http.StatusInternalServerError)

		return nil, nil, fmt.Errorf("workload not found in context")
	}

	log.Info().Str("workloadName", workload.Name).
		Str("project", workload.Project).
		Str("workloadType", workload.Type.String()).
		Str("job framework", workload.JobType.String()).Msg("Handling logs request")

	log.Debug().Str("workloadName", workload.Name).Str("namespace", workload.Namespace).
		Msg("Retrieved workload from context")

	options, err := FromGetRequest(r)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Options from request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		metrics.SetReasonIntoHeader(w, metrics.ErrOptionCreationFailed)

		return nil, nil, err
	}

	log.Debug().Interface("options", options).Msg("Created Options from request")

	return workload, &options, err
}
