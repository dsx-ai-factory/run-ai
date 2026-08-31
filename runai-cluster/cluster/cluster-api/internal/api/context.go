/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"context"
	"errors"
)

type contextKey int

const (
	workloadTypeContextKey contextKey = iota
	workloadActionContextKey
	workloadNameContextKey
	workloadFrameworkContextKey
)

const (
	notInContextMsg = "key not found in context"
)

func WithWorkloadType(ctx context.Context, workloadType WorkloadType) context.Context {
	return context.WithValue(ctx, workloadTypeContextKey, workloadType)
}

func WithWorkloadAction(ctx context.Context, workloadAction WorkloadAction) context.Context {
	return context.WithValue(ctx, workloadActionContextKey, workloadAction)
}

func WithWorkloadName(ctx context.Context, workloadName string) context.Context {
	return context.WithValue(ctx, workloadNameContextKey, workloadName)
}

func WithWorkloadFramework(ctx context.Context, workloadFramework WorkloadFramework) context.Context {
	return context.WithValue(ctx, workloadFrameworkContextKey, workloadFramework)
}

func WorkloadTypeFromContext(ctx context.Context) (WorkloadType, error) {
	workloadType, ok := ctx.Value(workloadTypeContextKey).(WorkloadType)
	if !ok {
		return workloadType, errors.New(notInContextMsg)
	}

	return workloadType, nil
}

func WorkloadActionFromContext(ctx context.Context) (WorkloadAction, error) {
	workloadAction, ok := ctx.Value(workloadActionContextKey).(WorkloadAction)
	if !ok {
		return workloadAction, errors.New(notInContextMsg)
	}

	return workloadAction, nil
}

func WorkloadNameFromContext(ctx context.Context) string {
	workloadName, _ := ctx.Value(workloadNameContextKey).(string)
	return workloadName
}

func WorkloadFrameworkFromContext(ctx context.Context) (WorkloadFramework, error) {
	workloadFramework, ok := ctx.Value(workloadFrameworkContextKey).(WorkloadFramework)
	if !ok {
		return workloadFramework, errors.New(notInContextMsg)
	}

	return workloadFramework, nil
}
