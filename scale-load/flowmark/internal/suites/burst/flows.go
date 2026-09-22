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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/measure"
	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/submit"
)

type deploymentFlow struct {
	kube        k8sclient.WithWatch
	concurrency int
}

func (*deploymentFlow) name() string {
	return "deployment_to_external_workload"
}

func (f *deploymentFlow) watch(ctx context.Context, namespace string) (*measure.Watcher, error) {
	return measure.WatchExternalWorkloads(ctx, f.kube, namespace)
}

func (f *deploymentFlow) create(ctx context.Context, workload item) (string, error) {
	deployment, err := submit.K8sDeployment(ctx, f.kube, submit.DeploymentSpec{
		Name: workload.name, Namespace: workload.namespace, Image: workload.image, Replicas: workload.pods, Labels: workload.labels,
	})
	if err != nil {
		return "", err
	}
	return string(deployment.UID), nil
}

func (f *deploymentFlow) teardown(ctx context.Context, namespace string, labels map[string]string) (int, error) {
	var deployments appsv1.DeploymentList
	if err := f.kube.List(ctx, &deployments, k8sclient.InNamespace(namespace), k8sclient.MatchingLabels(labels)); err != nil {
		return 0, fmt.Errorf("listing the round's deployments: %w", err)
	}
	targets := make([]k8sclient.Object, 0, len(deployments.Items))
	for i := range deployments.Items {
		if deployments.Items[i].DeletionTimestamp == nil {
			targets = append(targets, &deployments.Items[i])
		}
	}
	if err := deleteAll(ctx, f.concurrency, targets, func(ctx context.Context, obj k8sclient.Object) error {
		return k8sclient.IgnoreNotFound(f.kube.Delete(ctx, obj))
	}); err != nil {
		return len(deployments.Items), err
	}
	external, err := countAllInNamespace(ctx, f.kube, measure.ExternalWorkloadKind, namespace)
	if err != nil {
		return len(deployments.Items), err
	}
	return len(deployments.Items) + external, nil
}

type trainingFlow struct {
	clients     *client.Clients
	project     submit.ProjectRef
	concurrency int

	mu      sync.Mutex
	created map[string]string
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
	f.mu.Lock()
	ids := make([]string, 0, len(f.created))
	for id := range f.created {
		ids = append(ids, id)
	}
	f.mu.Unlock()

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
				delete(f.created, id)
				f.mu.Unlock()
			}
		}()
	}
	for _, id := range ids {
		jobs <- id
	}
	close(jobs)
	wg.Wait()

	f.mu.Lock()
	remaining := len(f.created)
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

func deleteAll(ctx context.Context, concurrency int, targets []k8sclient.Object, remove func(context.Context, k8sclient.Object) error) error {
	jobs := make(chan k8sclient.Object)
	errs := make(chan error, len(targets))
	var wg sync.WaitGroup
	for range concurrency {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for obj := range jobs {
				if err := remove(ctx, obj); err != nil {
					errs <- fmt.Errorf("deleting %s: %w", obj.GetName(), err)
				}
			}
		}()
	}
	for _, obj := range targets {
		jobs <- obj
	}
	close(jobs)
	wg.Wait()
	close(errs)
	var failures []error
	for err := range errs {
		failures = append(failures, err)
	}
	return errors.Join(failures...)
}

func countAllInNamespace(ctx context.Context, kube k8sclient.Client, kind schema.GroupVersionKind, namespace string) (int, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(schema.GroupVersionKind{Group: kind.Group, Version: kind.Version, Kind: kind.Kind + "List"})
	if err := kube.List(ctx, list, k8sclient.InNamespace(namespace)); err != nil {
		return 0, fmt.Errorf("listing %s objects: %w", kind.Kind, err)
	}
	return len(list.Items), nil
}
