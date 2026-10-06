/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"

	configmaps "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config-maps"
	. "github.com/onsi/gomega"
	resourcev1 "k8s.io/api/resource/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getRctFromClient(name, namespace string, k8sClient client.Client) *resourcev1.ResourceClaimTemplate {
	rct := &resourcev1.ResourceClaimTemplate{}
	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, rct)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return nil
	}
	return rct
}

func validateRctExistsEventually(name, namespace string, k8sClient client.Client) *resourcev1.ResourceClaimTemplate {
	var rct *resourcev1.ResourceClaimTemplate
	EventuallyWithOffset(1, func() *resourcev1.ResourceClaimTemplate {
		rct = getRctFromClient(name, namespace, k8sClient)
		return rct
	}, DefaultTimeout).ShouldNot(BeNil())
	return rct
}

func validateRctDeletedEventually(name, namespace string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *resourcev1.ResourceClaimTemplate {
		return getRctFromClient(name, namespace, k8sClient)
	}, DefaultTimeout).Should(BeNil())
}

func validateAgentConfigWithK8sObjectsCounter(k8sClient client.Client, counterValue string) {
	validateAgentConfigWithCounter(k8sClient, configmaps.K8sObjectsSyncCounter, counterValue)
}
