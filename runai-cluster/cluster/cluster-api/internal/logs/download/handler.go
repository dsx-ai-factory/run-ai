/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/response"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/logs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/router"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/rs/zerolog/log"
)

const (
	underlineStr  = "_"
	fileExtension = ".log"

	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
	HeaderTransferEncoding   = "Transfer-Encoding"
	HeaderCacheControl       = "Cache-Control"
	HeaderConnection         = "Connection"

	ValueApplication = "application/octet-stream"
	ValueChunked     = "chunked"
	ValueNoCache     = "no-cache"
	ValueKeepAlive   = "keep-alive"

	contentDispositionHeaderTemplate = "attachment; filename=\"%s\""

	errorLogsDownloadOptionsNotSupported = "Logs download does not support option attribute"
	failedToStreamLogsDownloader         = "Failed to stream logs downloader"

	bufferSize = 1024
)

var (
	ErrOptionFollowTrue            = errors.New("option follow true")
	ErrConnectionChunkNotSupported = errors.New("connection does not support chunks response")
)

type Handler struct {
	service logs.PodLogsService
}

func NewHandler(service logs.PodLogsService) router.UrlHandler {
	return &Handler{service: service}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// Create a context with cancel
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	reqWorkload, reqOptions, err := logs.GetWorkloadAndOptions(w, r)
	if err != nil {
		// should never happen
		response.ErrorBadRequestWithSetReason(w, logs.FailedToParseWorkloadOrOptions, metrics.ErrGetWorkloadAndOptions, err)
		return
	}

	if reqOptions.Follow {
		response.ErrorBadRequestWithSetReason(
			w, errorLogsDownloadOptionsNotSupported, metrics.ErrOptionNotSupported, ErrOptionFollowTrue)

		return
	}

	// Call the LogsReader method of the Handler's service
	log.Debug().Msgf("creating log download stream for workload:%s, ns:%s, pod:%s",
		reqWorkload.Name, reqWorkload.Namespace, reqOptions.Pod)

	logsReader, err := h.service.LogsReader(ctx, reqWorkload, reqOptions)
	if err != nil {
		response.ErrorWithReason(w, logs.FailedToGetLogsReader, err)
		return
	}

	defer func() {
		log.Debug().Msgf("closing log download stream for workload:%s, ns:%s, pod:%s",
			reqWorkload.Name, reqWorkload.Namespace, reqOptions.Pod)

		_ = logsReader.Close()
	}()

	log.Debug().Msg("Successfully obtained logs downloader")

	err = streamHandler(ctx, w, logsReader, createLogFileName(reqWorkload, reqOptions))
	if err != nil && err.Error() != "EOF" {
		response.ErrorWithReason(w, failedToStreamLogsDownloader, err)
		return
	}

	response.OK(w)
}

func createLogFileName(workload *workload.Workload, options *logs.Options) string {
	var sb strings.Builder

	sb.WriteString(workload.Name)
	sb.WriteString(underlineStr)
	sb.WriteString(workload.Project)
	// options is validate in the beginning of the handler, this is just for sanitation.
	if options != nil {
		sb.WriteString(underlineStr)
		sb.WriteString(options.Pod)
		sb.WriteString(underlineStr)
		sb.WriteString(options.Container)
	}

	sb.WriteString(underlineStr)
	sb.WriteString(time.Now().Format(time.RFC3339))
	sb.WriteString(fileExtension)

	return sb.String()
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling GET request")
	h.Handle(w, r)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	log.Debug().Msg("Handling POST request")
	h.Handle(w, r)
}

func streamHandler(ctx context.Context, w http.ResponseWriter, readCloser io.ReadCloser, filename string) error {
	var (
		buf  = make([]byte, bufferSize)
		size int
		err  error
	)

	// Set headers for chunked response
	w.Header().Set(HeaderContentType, ValueApplication)
	w.Header().Set(HeaderContentDisposition, getContentDispositionValue(filename))
	w.Header().Set(HeaderTransferEncoding, ValueChunked)
	w.Header().Set(HeaderCacheControl, ValueNoCache)
	w.Header().Set(HeaderConnection, ValueKeepAlive)

	// Ensure the writer supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		return ErrConnectionChunkNotSupported
	}

	// Send a few chunks
	for {
		size, err = readCloser.Read(buf)
		if err != nil {
			return err
		}

		metrics.MessageSentReport(metrics.FromContext(ctx), uint64(size))

		_, err = w.Write(buf[:size])
		if err != nil {
			return err
		}

		flusher.Flush()
	}
}

func getContentDispositionValue(filename string) string {
	return fmt.Sprintf(contentDispositionHeaderTemplate, filename)
}
