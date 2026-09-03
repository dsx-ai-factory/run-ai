/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package filter

import (
	"context"
	"fmt"

	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/diagnostics"
)

type NamespacesFilterClienter interface {
	VerifyAccess(ctx context.Context, sar *authorizationv1.SelfSubjectAccessReview) (bool, error)
	ListNamespaces(ctx context.Context) (*corev1.NamespaceList, error)
}

type NamespacesFilter struct {
	nsFilterClient NamespacesFilterClienter
	namespaces     []string
}

func NewNamespacesFilter(nsFilterClient NamespacesFilterClienter, namespaces []string) *NamespacesFilter {
	return &NamespacesFilter{
		nsFilterClient: nsFilterClient,
		namespaces:     unique(namespaces),
	}
}

func (nf *NamespacesFilter) Filter(ctx context.Context) (*diagnostics.FilterResult, error) {
	if len(nf.namespaces) == 0 {
		return &diagnostics.FilterResult{
			Valid:        []string{},
			NonExisting:  []string{},
			NonPermitted: []string{},
		}, nil
	}

	existingNamespaces, nonExistingNamespaces, err := nf.filterNonExistingNamespaces(ctx)
	if err != nil {
		return nil, err
	}

	nonPermittedNamespaces, err := nf.filterNonPermittedNamespaces(ctx, existingNamespaces)
	if err != nil {
		return nil, err
	}

	validNamespaces := diff(existingNamespaces, nonPermittedNamespaces)

	return &diagnostics.FilterResult{
		Valid:        validNamespaces,
		NonExisting:  nonExistingNamespaces,
		NonPermitted: nonPermittedNamespaces,
	}, nil
}

func (nf *NamespacesFilter) filterNonExistingNamespaces(ctx context.Context) ([]string, []string, error) {
	nsList, err := nf.nsFilterClient.ListNamespaces(ctx)
	if err != nil {
		return nil, nil, err
	}

	nsSet := make(map[string]struct{})
	for _, ns := range nsList.Items {
		nsSet[ns.Name] = struct{}{}
	}

	existingNamespaces := make([]string, 0, len(nf.namespaces))
	nonExistingNamespaces := make([]string, 0, len(nf.namespaces))

	for _, ns := range nf.namespaces {
		if _, ok := nsSet[ns]; ok {
			existingNamespaces = append(existingNamespaces, ns)
			continue
		}

		nonExistingNamespaces = append(nonExistingNamespaces, ns)
	}

	return existingNamespaces, nonExistingNamespaces, nil
}

func (nf *NamespacesFilter) filterNonPermittedNamespaces(ctx context.Context, namespaces []string) ([]string, error) {
	nonPermittedNamespaces := make([]string, 0, len(namespaces))

	for _, ns := range namespaces {
		permitted, err := nf.checkNamespacePermissions(ctx, ns)
		if err != nil {
			return nil, err
		}
		if !permitted {
			nonPermittedNamespaces = append(nonPermittedNamespaces, ns)
		}
	}

	return nonPermittedNamespaces, nil
}

func (nf *NamespacesFilter) checkNamespacePermissions(ctx context.Context, namespace string) (bool, error) {
	checks := []struct {
		desc string
		sar  *authorizationv1.SelfSubjectAccessReview
	}{
		{desc: "list pods", sar: listPodsSAR(namespace)},
		{desc: "get pod logs", sar: podLogsSAR(namespace)},
	}

	for _, c := range checks {
		allowed, err := nf.nsFilterClient.VerifyAccess(ctx, c.sar)
		if err != nil {
			return false, fmt.Errorf("failed to verify %s access permissions in %s: %w", c.desc, namespace, err)
		}
		if !allowed {
			return false, nil
		}
	}

	return true, nil
}

func listPodsSAR(namespace string) *authorizationv1.SelfSubjectAccessReview {
	return &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: namespace,
				Verb:      "list",
				Group:     "",
				Version:   "v1",
				Resource:  "pods",
			},
		},
	}
}

func podLogsSAR(namespace string) *authorizationv1.SelfSubjectAccessReview {
	return &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace:   namespace,
				Verb:        "get",
				Group:       "",
				Version:     "v1",
				Resource:    "pods",
				Subresource: "log",
			},
		},
	}
}

func diff(source, exclude []string) []string {
	excludeSet := make(map[string]struct{})
	for _, item := range exclude {
		excludeSet[item] = struct{}{}
	}

	result := make([]string, 0, len(source))
	for _, item := range source {
		if _, exists := excludeSet[item]; !exists {
			result = append(result, item)
		}
	}
	return result
}

func unique(items []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(items))

	for _, item := range items {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
