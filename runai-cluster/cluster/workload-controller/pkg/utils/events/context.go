/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events

import (
	"context"
)

const (
	ContextRecorderName string = "workloadRecorder"
)

type EmptyRecorder struct {
}

func NewEmptyRecorder() *EmptyRecorder {
	return &EmptyRecorder{}
}

func (r *EmptyRecorder) Event(_, _, _ string) {}

func (r *EmptyRecorder) Eventf(_, _, _ string, _ ...interface{}) {}

func ToContext(ctx context.Context, recorder Recorder) context.Context {
	return context.WithValue(ctx, ContextRecorderName, recorder)
}

func FromContext(ctx context.Context) Recorder {
	recorder, ok := ctx.Value(ContextRecorderName).(Recorder)
	if !ok {
		return NewEmptyRecorder()
	}

	return recorder
}
