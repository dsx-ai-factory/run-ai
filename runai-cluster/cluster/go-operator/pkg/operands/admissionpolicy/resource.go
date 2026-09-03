/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package admissionpolicy

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type resourceType string

const (
	policyResource  resourceType = "validating admission policy"
	bindingResource resourceType = "validating admission policy binding"
)

type resource struct {
	operandName  string
	k8sReader    client.Reader
	definition   Definition
	resourceType resourceType
}

var _ operands.Resource = &resource{}

func NewResources(operandName string, k8sReader client.Reader, definition Definition) []operands.Resource {
	definition = definition.deepCopy()
	return []operands.Resource{
		&resource{
			operandName:  operandName,
			k8sReader:    k8sReader,
			definition:   definition,
			resourceType: policyResource,
		},
		&resource{
			operandName:  operandName,
			k8sReader:    k8sReader,
			definition:   definition,
			resourceType: bindingResource,
		},
	}
}

func (r *resource) OperandName() string {
	return r.operandName
}

func (r *resource) Object(_ context.Context) (client.Object, error) {
	if r.operandName == "" {
		return nil, fmt.Errorf("operand name is required for admission policy resources")
	}
	if err := r.definition.validate(); err != nil {
		return nil, err
	}

	switch r.resourceType {
	case policyResource:
		return r.definition.Policy.DeepCopy(), nil
	case bindingResource:
		return r.definition.Binding.DeepCopy(), nil
	default:
		return nil, fmt.Errorf("unsupported admission policy resource type %q", r.resourceType)
	}
}

func (r *resource) Status(ctx context.Context) error {
	if r.k8sReader == nil {
		return fmt.Errorf("kubernetes reader is required for %s status", r.resourceType)
	}

	desired, err := r.Object(ctx)
	if err != nil {
		return err
	}
	key := client.ObjectKeyFromObject(desired)
	if err := r.k8sReader.Get(ctx, key, desired); err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("%s [%s] is missing", r.resourceType, key.Name)
		}
		return err
	}
	return nil
}

func (r *resource) Enabled() bool {
	return r.definition.enabled()
}
