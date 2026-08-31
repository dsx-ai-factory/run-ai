/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvlutils

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
)

type supportKey string

const (
	SupportContextMNNVL supportKey = "mnnvl"
)

func FromContext(ctx context.Context) Status {

	status, ok := ctx.Value(SupportContextMNNVL).(Status)
	if !ok {
		return Status{
			MasterMode: mnnvl.ModeNone,
			WorkerMode: mnnvl.ModeNone,
		}
	}
	return status
}

func ToContext(ctx context.Context, status Status) context.Context {
	return context.WithValue(ctx, SupportContextMNNVL, status)
}
