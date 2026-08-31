/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"fmt"
	ocpv1 "github.com/openshift/api/operator/v1"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func openshiftPreInstallOperations(k8s client.Client) error {
	exists, err := ocpIngressOperatorExists(k8s)
	if err != nil {
		return fmt.Errorf("failed to check if ocp ingress operator exists: %w", err)
	}

	if exists {
		if err := addOpenshiftMonitoringToRunaiNamespace(k8s); err != nil {
			return fmt.Errorf("failed to add openshift monitoring to runai namespace: %w", err)
		}
		if err := patchOcpIngressDefaultController(k8s); err != nil {
			return fmt.Errorf("failed to patch ocp ingress default controller: %w", err)
		}
	}

	return nil
}

func patchOcpIngressDefaultController(k8s client.Client) error {
	ocpIngressDefaultController := &ocpv1.IngressController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "default",
			Namespace: "openshift-ingress-operator",
		},
	}
	patchErr := k8s.Patch(context.Background(), ocpIngressDefaultController, client.RawPatch(types.MergePatchType, []byte(`{"spec": {"routeAdmission": {"namespaceOwnership": "InterNamespaceAllowed"}}}`)))
	if patchErr != nil {
		return fmt.Errorf("failed to patch ocp ingress controller: %w", patchErr)
	}
	return nil
}

func ocpIngressOperatorExists(k8s client.Client) (bool, error) {
	ocpIngressOperatorNS := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "openshift-ingress-operator",
		},
	}

	ocpErr := k8s.Get(context.Background(), client.ObjectKeyFromObject(ocpIngressOperatorNS), ocpIngressOperatorNS)
	if k8sErrors.IsNotFound(ocpErr) {
		return false, nil
	}
	if ocpErr != nil {
		return false, fmt.Errorf("failed to check if ingress operator exists: %w", ocpErr)
	}

	return true, nil
}

func addOpenshiftMonitoringToRunaiNamespace(k8s client.Client) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.DefaultNamespace,
		},
	}

	patchErr := k8s.Patch(context.Background(), ns, client.RawPatch(types.MergePatchType, []byte(`{"metadata": {"labels": {"openshift.io/cluster-monitoring": "true"}}}`)))
	if patchErr != nil {
		return fmt.Errorf("failed to patch runai namespace: %w", patchErr)
	}
	return nil
}
