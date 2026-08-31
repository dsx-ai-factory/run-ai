/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sync

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
)

type WorkerPool[T any] struct {
	numberOfWorkers int
	handler         func(context.Context, T) error
	taskList        []T
}

func NewWorkerPool[T any](numberOfWorkers int, handler func(context.Context, T) error, taskList []T) WorkerPool[T] {
	return WorkerPool[T]{
		numberOfWorkers: numberOfWorkers,
		handler:         handler,
		taskList:        taskList,
	}
}

func (w WorkerPool[T]) Execute(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(w.numberOfWorkers)
	for _, task := range w.taskList {
		if groupCtx.Err() != nil {
			break
		}
		taskCp := task
		group.Go(func() error {
			return w.handler(groupCtx, taskCp)
		})
	}
	err := group.Wait()
	if err != nil {
		return fmt.Errorf("failed to execute all tasks: %w", err)
	}
	if ctx.Err() != nil {
		return fmt.Errorf("paranet context is cancelled: %w", groupCtx.Err())
	}

	return nil
}
