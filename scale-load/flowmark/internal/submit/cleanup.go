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
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/client"
)

var clusterProjectKind = schema.GroupVersionKind{Group: "kai.resources", Version: "v1alpha1", Kind: "Project"}

func CleanupProject(ctx context.Context, clients *client.Clients, clusterID, name string, timeout time.Duration) error {
	project, err := managedProject(ctx, clients, clusterID, name)
	if errors.Is(err, errProjectNotFound) {
		return removeOrphanNamespace(ctx, clients.Kube, name)
	}
	if err != nil {
		return err
	}

	response, err := clients.OrgUnits.ProjectsAPI.DeleteProject(ctx, project.Id).Execute()
	if err != nil {
		return fmt.Errorf("deleting project %q: %w", name, err)
	}
	if response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("deleting project %q: %s", name, response.Status)
	}

	if err := waitClusterProjectGone(ctx, clients.Kube, name, timeout); err != nil {
		return err
	}
	reported, ok := project.Status.GetNamespaceOk()
	if !ok || reported == nil || *reported == "" {
		return removeOrphanNamespace(ctx, clients.Kube, name)
	}
	return deleteProjectNamespace(ctx, clients.Kube, *reported, name)
}

func clusterProject(name string) *unstructured.Unstructured {
	cr := &unstructured.Unstructured{}
	cr.SetGroupVersionKind(clusterProjectKind)
	cr.SetName(name)
	return cr
}

func waitClusterProjectGone(ctx context.Context, kube k8sclient.Client, name string, timeout time.Duration) error {
	var last error
	err := wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, func(ctx context.Context) (bool, error) {
		last = kube.Get(ctx, k8sclient.ObjectKey{Name: name}, clusterProject(name))
		if meta.IsNoMatchError(last) {
			return false, fmt.Errorf("cluster has no %s kind; is --kubeconfig pointing at the cluster of project %q? its namespace was left alone: %w", clusterProjectKind.Kind, name, last)
		}
		return apierrors.IsNotFound(last), nil
	})
	if err == nil {
		return nil
	}
	if meta.IsNoMatchError(err) {
		return err
	}
	if last != nil {
		return fmt.Errorf("cluster project %q still present after %s (last: %v): %w", name, timeout, last, err)
	}
	return fmt.Errorf("cluster project %q still present after %s: %w", name, timeout, err)
}

func deleteProjectNamespace(ctx context.Context, kube k8sclient.Client, namespace, name string) error {
	var ns corev1.Namespace
	err := kube.Get(ctx, k8sclient.ObjectKey{Name: namespace}, &ns)
	if apierrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading namespace %q: %w", namespace, err)
	}
	if ns.Labels[projectNamespaceLabel] != name {
		return fmt.Errorf("namespace %q is not labelled %s=%s; refusing to delete it", namespace, projectNamespaceLabel, name)
	}
	if err := k8sclient.IgnoreNotFound(kube.Delete(ctx, &ns)); err != nil {
		return fmt.Errorf("deleting namespace %q: %w", namespace, err)
	}
	return nil
}

func removeOrphanNamespace(ctx context.Context, kube k8sclient.Client, name string) error {
	var namespaces corev1.NamespaceList
	if err := kube.List(ctx, &namespaces, k8sclient.MatchingLabels{projectNamespaceLabel: name}); err != nil {
		return fmt.Errorf("looking for the namespace of project %q: %w", name, err)
	}
	if len(namespaces.Items) == 0 {
		return nil
	}
	err := kube.Get(ctx, k8sclient.ObjectKey{Name: name}, clusterProject(name))
	switch {
	case err == nil:
		return fmt.Errorf("project %q is gone from the control plane but still present on the cluster; not touching it", name)
	case !apierrors.IsNotFound(err):
		return fmt.Errorf("checking project %q on the cluster before removing its namespace: %w", name, err)
	}
	for i := range namespaces.Items {
		if err := k8sclient.IgnoreNotFound(kube.Delete(ctx, &namespaces.Items[i])); err != nil {
			return fmt.Errorf("deleting orphan namespace %q: %w", namespaces.Items[i].Name, err)
		}
	}
	return nil
}
