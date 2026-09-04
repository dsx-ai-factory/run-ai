/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package karta

import (
	"context"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// InstanceMutator mutates one extracted pod instance. The pod spec is addressed
// through pointers, so mutating in place is what gets written back.
type InstanceMutator func(instance kartaresource.ExtractedInstance) error

// MutatePodInstances applies mutate to every pod instance of every component of
// resource - as described by karta - and writes the result back into resource.
func MutatePodInstances(
	ctx context.Context,
	karta *kartav1alpha1.Karta,
	resource *unstructured.Unstructured,
	mutate InstanceMutator,
) error {
	factory := kartaresource.NewComponentFactoryFromObject(karta, resource)

	components, err := GetAllComponents(factory)
	if err != nil {
		return err
	}

	for _, component := range components {
		instances, err := component.GetExtractedInstances(ctx)
		if err != nil {
			return fmt.Errorf("extract instances: %w", err)
		}

		for _, instance := range instances {
			if err := mutate(instance); err != nil {
				return fmt.Errorf("mutate instance: %w", err)
			}
		}

		if err := UpdateComponent(ctx, component, instances); err != nil {
			return fmt.Errorf("update component: %w", err)
		}
	}

	updated, err := factory.GetResource()
	if err != nil {
		return fmt.Errorf("get updated resource: %w", err)
	}
	resource.Object = updated.(*unstructured.Unstructured).Object

	return nil
}
