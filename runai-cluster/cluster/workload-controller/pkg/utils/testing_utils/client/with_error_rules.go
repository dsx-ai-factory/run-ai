/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/resources"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FakeClientWithErrorRules allows returning custom error per resource
type FakeClientWithErrorRules struct {
	client.Client
	rulesList []ResourceFailingRule
}

func NewFakeClientWithErrorRules(scheme *runtime.Scheme, statusObjects []client.Object, rules []ResourceFailingRule, objects ...client.Object) *FakeClientWithErrorRules {
	if rules == nil {
		rules = []ResourceFailingRule{}
	}

	return &FakeClientWithErrorRules{Client: InitFakeClientWithObjectWithStatus(scheme, statusObjects, objects...), rulesList: rules}
}

func (c *FakeClientWithErrorRules) UpdateRules(rules []ResourceFailingRule) {
	c.rulesList = rules
}

func (c *FakeClientWithErrorRules) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	err := c.tryGenerateErrorFromRules(obj, Create)
	if err != nil {
		return err
	}

	return c.Client.Create(ctx, obj, opts...)
}

func (c *FakeClientWithErrorRules) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	err := c.tryGenerateErrorFromRules(obj, Update)
	if err != nil {
		return err
	}

	return c.Client.Update(ctx, obj, opts...)
}

func (c *FakeClientWithErrorRules) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	err := c.tryGenerateErrorFromRules(obj, Delete)
	if err != nil {
		return err
	}

	return c.Client.Delete(ctx, obj, opts...)
}

func (c *FakeClientWithErrorRules) tryGenerateErrorFromRules(obj client.Object, action ClientAction) error {
	for _, rule := range c.rulesList {
		if resources.SafeGetKind(obj) == rule.ResourceKind && obj.GetName() == rule.ResourceName && (rule.ClientAction == Unspecified || rule.ClientAction == action) {
			if rule.StatusError != nil {
				return rule.StatusError
			}
			return rule.GenerateDefaultError(obj.GetName())
		}
	}

	return nil
}
