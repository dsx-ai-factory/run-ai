/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package errors

import (
	"context"
	"net/http"

	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type (
	ClientErrorReason int32
	Entity            int32
)

const (
	GeneralError ClientErrorReason = iota
	NotFound
	Conflict
	TenantNotFound
)

const (
	Tenant Entity = iota
	MetricStore
)

const TenantNotFoundMsg = "Tenant not found"

func (e Entity) String() string {
	switch e {
	case Tenant:
		return "Tenant"
	case MetricStore:
		return "MetricStore"
	default:
		return "No such entity"
	}
}

type BusinessError struct {
	Reason ClientErrorReason
	Err    error
}

func (e *BusinessError) Error() string {
	if e == nil || e.Err == nil {
		return ""
	}

	return e.Err.Error()
}

func LogHandler(ctx context.Context, msg string, statusCode int) {
	if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", msg)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", msg)
	}
}
