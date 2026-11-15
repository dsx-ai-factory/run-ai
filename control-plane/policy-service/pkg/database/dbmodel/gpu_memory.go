/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodel

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
)

// NormalizeGPUMemory resolves each persisted source before inheritance or patching.
func (p *PolicyWithOptionalEffective) NormalizeGPUMemory() error {
	if p.GvkWorkload {
		return nil
	}
	var err error
	p.Policy.DefaultsV2, err = gpucompat.NormalizeLegacyGPUFields(p.Policy.DefaultsV2)
	if err != nil {
		return fmt.Errorf("policy defaults: %w", err)
	}
	p.Policy.RulesV2, err = gpucompat.NormalizeLegacyGPUPolicy(p.Policy.RulesV2)
	if err != nil {
		return fmt.Errorf("policy rules: %w", err)
	}
	p.EffectivePolicy.DefaultsV2, err = gpucompat.NormalizeLegacyGPUFields(p.EffectivePolicy.DefaultsV2)
	if err != nil {
		return fmt.Errorf("effective defaults: %w", err)
	}
	p.EffectivePolicy.RulesV2, err = gpucompat.NormalizeLegacyGPUPolicy(p.EffectivePolicy.RulesV2)
	if err != nil {
		return fmt.Errorf("effective rules: %w", err)
	}
	return nil
}

func (p Policy) ForDB() (*Policy, error) {
	if p.GvkWorkload {
		return &p, nil
	}
	var err error
	p.DefaultsV2, err = gpucompat.LegacyGPUFields(p.DefaultsV2)
	if err != nil {
		return nil, fmt.Errorf("store policy defaults: %w", err)
	}
	p.RulesV2, err = gpucompat.LegacyGPUPolicy(p.RulesV2)
	if err != nil {
		return nil, fmt.Errorf("store policy rules: %w", err)
	}
	return &p, nil
}

func (p EffectivePolicy) ForDB() (*EffectivePolicy, error) {
	// GVK rules have an array wrapper and no native defaults or dependency order.
	if _, gvk := p.RulesV2[GvkRulesJsonKey]; gvk {
		return &p, nil
	}
	var err error
	p.DefaultsV2, err = gpucompat.LegacyGPUFields(p.DefaultsV2)
	if err != nil {
		return nil, fmt.Errorf("store effective defaults: %w", err)
	}
	p.RulesV2, err = gpucompat.LegacyGPUPolicy(p.RulesV2)
	if err != nil {
		return nil, fmt.Errorf("store effective rules: %w", err)
	}
	p.DefaultFromOrder = defaultFromOrderForDB(p.DefaultFromOrder)
	return &p, nil
}

func defaultFromOrderForDB(src *api.DefaultFromOrder) *api.DefaultFromOrder {
	if src == nil {
		return nil
	}
	return &api.DefaultFromOrder{
		Worker: gpucompat.LegacyGPUDefaultFromOrder(src.Worker),
		Master: gpucompat.LegacyGPUDefaultFromOrder(src.Master),
	}
}
