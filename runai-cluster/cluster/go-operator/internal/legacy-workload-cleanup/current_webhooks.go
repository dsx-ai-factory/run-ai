/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"context"
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// currentWebhookConfigurations are live admission-controller webhooks that older cluster versions
// registered with retired resources in their rules. The Operator's init-ca eventually rewrites them
// without those resources, but the cleanup Job may run before that happens. Stripping the retired
// resources here makes validateLegacyWebhookState independent of that ordering; init-ca converges
// to the same rule set afterwards.
var currentWebhookConfigurations = []namedResourceSet{
	{gvr: schema.GroupVersionResource{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}, kind: "MutatingWebhookConfiguration",
		names: []string{"runai-pod-controller-mutation"}},
}

func stripRetiredResourcesFromCurrentWebhooks(ctx context.Context, k8sClient client.Client) error {
	for _, resourceSet := range currentWebhookConfigurations {
		for _, name := range resourceSet.names {
			if err := stripRetiredResourcesFromMutatingWebhook(ctx, k8sClient, name); err != nil {
				return err
			}
		}
	}
	return nil
}

func stripRetiredResourcesFromMutatingWebhook(ctx context.Context, k8sClient client.Client, name string) error {
	configuration := &admissionv1.MutatingWebhookConfiguration{}
	if err := k8sClient.Get(ctx, client.ObjectKey{Name: name}, configuration); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get current MutatingWebhookConfiguration %s: %w", name, err)
	}
	changed := false
	for i := range configuration.Webhooks {
		rules, rulesChanged := withoutRetiredResources(configuration.Webhooks[i].Rules)
		if rulesChanged {
			configuration.Webhooks[i].Rules = rules
			changed = true
		}
	}
	if !changed {
		return nil
	}
	if err := k8sClient.Update(ctx, configuration); err != nil {
		return fmt.Errorf("strip retired resources from MutatingWebhookConfiguration %s: %w", name, err)
	}
	return nil
}

// withoutRetiredResources removes retired resources from every rule and drops rules left empty.
func withoutRetiredResources(rules []admissionv1.RuleWithOperations) ([]admissionv1.RuleWithOperations, bool) {
	remaining := make([]admissionv1.RuleWithOperations, 0, len(rules))
	changed := false
	for _, rule := range rules {
		resources := make([]string, 0, len(rule.Resources))
		for _, resource := range rule.Resources {
			baseResource, _, _ := strings.Cut(resource, "/")
			if _, retired := retiredResources[baseResource]; retired {
				changed = true
				continue
			}
			resources = append(resources, resource)
		}
		if len(resources) == 0 {
			continue
		}
		rule.Resources = resources
		remaining = append(remaining, rule)
	}
	return remaining, changed
}
