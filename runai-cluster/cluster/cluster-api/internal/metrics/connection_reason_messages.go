/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import "net/http"

type ReasonType string

const (
	ErrOptionCreationFailed  ReasonType = "option_creation_failed"
	ErrGenerateRequest       ReasonType = "generate_request_failed"
	ErrUpgradeConnection     ReasonType = "upgrade_connection_failed"
	ErrGetWorkloadAndOptions ReasonType = "get_workload_and_options_failed"
	ErrOptionNotSupported    ReasonType = "option_not_supported"
	ErrStartProxy            ReasonType = "start_proxy_failed"
	MessageOK                ReasonType = "ok"

	reasonHeader = "X-Runai-Reason"
)

// SetReasonIntoHeader middleware can only move context higher in the chain,
// so we use headers to be able to extract reason
func SetReasonIntoHeader(w http.ResponseWriter, reason ReasonType) {
	w.Header().Set(reasonHeader, string(reason))
}

func GetReasonFromHeader(w http.ResponseWriter) string {
	return w.Header().Get(reasonHeader)
}
