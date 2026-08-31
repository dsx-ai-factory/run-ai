/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package orchestrator2

import "context"

type FailureDecision int

const (
	FailureAbort FailureDecision = iota
	FailureSkip
	FailureRetry
)

type FailureInfo struct {
	Pipeline string
	Flow     string
	Step     string
	Phase    string
	Err      error
}

type FailureHandler interface {
	OnFailure(ctx context.Context, info FailureInfo) FailureDecision
}
