/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	riresource "github.com/run-ai/karta/pkg/resource"
	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/kartaparser"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KartaTopOwnerHandler struct {
	cachedClient client.Client
	ewConverter  *kartaparser.EWConverter
}

func NewKartaTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *KartaTopOwnerHandler {
	return &KartaTopOwnerHandler{
		cachedClient: cachedClient,
		ewConverter:  kartaparser.NewEWConverter(cachedClient, draParser),
	}
}

func (h *KartaTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	karta *kartav1alpha1.Karta,
	externalWorkload *v1alpha1.ExternalWorkload) error {

	objectMeta := metav1.ObjectMeta{
		Labels:      object.GetLabels(),
		Annotations: object.GetAnnotations(),
		Namespace:   object.GetNamespace(),
		Name:        object.GetName(),
	}
	gk := object.GetObjectKind().GroupVersionKind().GroupKind()

	externalWorkload.Spec.RequestedResources = corev1.ResourceRequirements{}
	externalWorkload.Spec.RequestedPods = v1alpha1.RequestedPods{}
	externalWorkload.Spec.StorageRefs = nil

	factory := riresource.NewComponentFactoryFromObject(karta, object)

	rootComponent, err := factory.GetRootComponent()
	if err != nil {
		return fmt.Errorf("failed to get root component: %w", err)
	}
	err = h.ewConverter.ConvertComponentToExternalWorkloadSpec(ctx,
		objectMeta, gk, rootComponent, true, externalWorkload)
	if err != nil {
		return fmt.Errorf("failed to parse root component definition: %v", err)
	}

	childComponents, err := factory.GetChildComponents()
	if err != nil {
		return fmt.Errorf("failed to get child components: %w", err)
	}
	for _, childComponent := range childComponents {
		if childComponent.Definition().OwnerRef == nil || *childComponent.Definition().OwnerRef != rootComponent.Name() {
			continue
		}

		err = h.ewConverter.ConvertComponentToExternalWorkloadSpec(ctx,
			objectMeta, gk, childComponent, false, externalWorkload)
		if err != nil {
			return fmt.Errorf("failed to parse child component <%s> definition: %v", childComponent.Name(), err)
		}

		// LWS has a structural group directly below the root, with the pod-bearing
		// leader and worker components one level below it.
		if !childComponent.HasPodDefinition() {
			for _, nestedComponent := range childComponents {
				if nestedComponent.Definition().OwnerRef == nil ||
					*nestedComponent.Definition().OwnerRef != childComponent.Name() ||
					!nestedComponent.HasPodDefinition() {
					continue
				}

				err = h.ewConverter.ConvertComponentToExternalWorkloadSpec(ctx,
					objectMeta, gk, nestedComponent, false, externalWorkload)
				if err != nil {
					return fmt.Errorf("failed to parse child component <%s> definition: %v", nestedComponent.Name(), err)
				}
			}
		}
	}

	h.ewConverter.FillDefaultsAfterAllParsing(ctx, objectMeta, gk, externalWorkload)
	return nil
}
