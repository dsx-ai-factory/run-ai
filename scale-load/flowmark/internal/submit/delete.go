/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

const (
	WorkspaceType        = "Workspace"
	TrainingType         = "Training"
	reportedDeleteErrors = 5
)

func DeleteWorkloads(ctx context.Context, clients *client.Clients, workloadType string, ids []string, concurrency int) (undeleted []string, err error) {
	remove, err := deleter(clients, workloadType)
	if err != nil {
		return ids, err
	}
	var mu sync.Mutex
	var errs []error
	jobs := make(chan string)
	var workers sync.WaitGroup
	for range max(concurrency, 1) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for id := range jobs {
				response, err := remove(ctx, id)
				if err == nil || (response != nil && response.StatusCode == http.StatusNotFound) {
					continue
				}
				mu.Lock()
				undeleted = append(undeleted, id)
				errs = append(errs, fmt.Errorf("deleting %s %s: %w", strings.ToLower(workloadType), id, err))
				mu.Unlock()
			}
		}()
	}
	sent := 0
	for ; sent < len(ids) && ctx.Err() == nil; sent++ {
		jobs <- ids[sent]
	}
	close(jobs)
	workers.Wait()
	if sent < len(ids) {
		undeleted = append(undeleted, ids[sent:]...)
		errs = append(errs, fmt.Errorf("stopped with %d deletes left: %w", len(ids)-sent, ctx.Err()))
	}
	return undeleted, summarize(errs)
}

func deleter(clients *client.Clients, workloadType string) (func(context.Context, string) (*http.Response, error), error) {
	switch workloadType {
	case WorkspaceType:
		return func(ctx context.Context, id string) (*http.Response, error) {
			_, response, err := clients.Workloads.WorkspacesAPI.DeleteWorkspace(ctx, id).Execute()
			return response, err
		}, nil
	case TrainingType:
		return func(ctx context.Context, id string) (*http.Response, error) {
			_, response, err := clients.Workloads.TrainingsAPI.DeleteTraining(ctx, id).Execute()
			return response, err
		}, nil
	default:
		return nil, fmt.Errorf("workload type %q cannot be deleted through the control plane", workloadType)
	}
}

func summarize(errs []error) error {
	if len(errs) <= reportedDeleteErrors {
		return errors.Join(errs...)
	}
	shown := append([]error(nil), errs[:reportedDeleteErrors]...)
	return errors.Join(append(shown, fmt.Errorf("and %d more", len(errs)-reportedDeleteErrors))...)
}
