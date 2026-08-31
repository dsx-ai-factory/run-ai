/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package response

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/metrics"
)

const (
	StatusOkMessage = "Request handled successfully"
)

func ErrorWithReason(w http.ResponseWriter, logMessage string, err error) {
	log.Error().Err(err).Msg(logMessage)
	message, code := api.ToHttpError(err)
	http.Error(w, message, code)
	// limited to specific error strings
	metrics.SetReasonIntoHeader(w, metrics.ReasonType(message))
}

func ErrorBadRequestWithSetReason(w http.ResponseWriter, message string, reason metrics.ReasonType, err error) {
	log.Error().Err(err).Msg(message)
	http.Error(w, message, http.StatusBadRequest)
	metrics.SetReasonIntoHeader(w, reason)
}

func ErrorInternalServerWithSetReason(w http.ResponseWriter, message string, reason metrics.ReasonType, err error) {
	log.Error().Err(err).Msg(message)
	http.Error(w, message, http.StatusInternalServerError)
	metrics.SetReasonIntoHeader(w, reason)
}

func OK(w http.ResponseWriter) {
	metrics.SetReasonIntoHeader(w, metrics.MessageOK)
	w.WriteHeader(http.StatusOK)
	log.Info().Msg(StatusOkMessage)
}
