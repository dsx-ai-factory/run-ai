/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package burst

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

type trainingFlow struct {
	clients     *client.Clients
	project     submit.ProjectRef
	concurrency int

	mu              sync.Mutex
	created         map[string]string
	deleteRequested map[string]bool
}

func (*trainingFlow) name() string {
	return "training_to_workload_runner"
}

func (f *trainingFlow) watch(ctx context.Context, namespace string) (*measure.Watcher, error) {
	return measure.WatchWorkloads(ctx, f.clients.Kube, namespace)
}

func (f *trainingFlow) create(ctx context.Context, workload item) (string, error) {
	ref, err := submit.Training(ctx, f.clients, submit.WorkloadSpec{
		Name: workload.name, Project: f.project, Image: workload.image, Labels: workload.labels,
	})
	if err != nil {
		return "", err
	}
	f.mu.Lock()
	f.created[ref.ID] = ref.Name
	f.mu.Unlock()
	return ref.Name, nil
}

func (f *trainingFlow) teardown(ctx context.Context, namespace string, _ map[string]string) (int, error) {
	listed, err := submit.ProjectWorkloadIDs(ctx, f.clients, f.project.ID, trainingType)
	if err != nil {
		return 0, err
	}
	f.mu.Lock()
	if f.deleteRequested == nil {
		f.deleteRequested = map[string]bool{}
	}
	for _, id := range listed {
		if _, known := f.created[id]; !known {
			f.created[id] = ""
		}
	}
	var pending []string
	for id := range f.created {
		if !f.deleteRequested[id] {
			pending = append(pending, id)
		}
	}
	f.mu.Unlock()

	errs := f.deleteTrainings(ctx, pending)
	f.mu.Lock()
	remaining := len(f.created) - len(f.deleteRequested)
	f.mu.Unlock()
	if len(errs) > 0 {
		return remaining, errors.Join(errs...)
	}
	runners, err := countAllInNamespace(ctx, f.clients.Kube, measure.WorkloadRunnerKind, namespace)
	if err != nil {
		return remaining, err
	}
	return remaining + runners, nil
}

func (f *trainingFlow) deleteTrainings(ctx context.Context, ids []string) []error {
	var errs []error
	var errsMu sync.Mutex
	jobs := make(chan string)
	var wg sync.WaitGroup
	for range f.concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				_, response, err := f.clients.Workloads.TrainingsAPI.DeleteTraining(ctx, id).Execute()
				if err != nil && (response == nil || response.StatusCode != http.StatusNotFound) {
					errsMu.Lock()
					errs = append(errs, fmt.Errorf("deleting training %s: %w", id, err))
					errsMu.Unlock()
					continue
				}
				f.mu.Lock()
				f.deleteRequested[id] = true
				f.mu.Unlock()
			}
		}()
	}
	for _, id := range ids {
		jobs <- id
	}
	close(jobs)
	wg.Wait()
	return errs
}
