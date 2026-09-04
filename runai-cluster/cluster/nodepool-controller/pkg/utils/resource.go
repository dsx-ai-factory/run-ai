/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/operands"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateOrUpdateResources creates or updates a slice of resources
func CreateOrUpdateResources(k8sClient client.Client, scheme *runtime.Scheme,
	ctx context.Context, logger logr.Logger,
	nodePool *v1alpha1.NodePool, resources []operands.ResourceOld) error {
	for _, resource := range resources {
		err := SetControllerReference(nodePool, resource.Object, scheme)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("handling [%s] operand resource [%s/%s/%s]",
			resource.OperandName, resource.Object.GetNamespace(),
			reflect.TypeOf(resource.Object).Elem().Name(),
			resource.Object.GetName()))

		err = CreateOrUpdateIfNeeded(k8sClient, ctx, logger, resource.Object)
		if err != nil && !errors.IsAlreadyExists(err) {
			return err
		}
	}

	return nil
}
