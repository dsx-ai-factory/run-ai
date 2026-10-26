/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type WorkloadGenerator struct {
	workload openapi.Workload
}

func (wg *WorkloadGenerator) SetTenant(id int32) *WorkloadGenerator {
	wg.workload.TenantId = id
	return wg
}

func (wg *WorkloadGenerator) SetType(wlType api.WorkloadType) *WorkloadGenerator {
	wg.workload.Type = wlType.String()
	return wg
}

func (wg *WorkloadGenerator) SetPods(running int32, requested, parallelism, completions *int32) *WorkloadGenerator {
	wg.workload.RunningPods = running
	requestedPods := &openapi.RequestedPods{}
	if requested != nil {
		requestedPods.Number.Set(requested)
	}
	if parallelism != nil {
		requestedPods.Parallelism.Set(parallelism)
	}
	if completions != nil {
		requestedPods.Completions.Set(completions)
	}
	wg.workload.RequestedPods = requestedPods
	return wg
}

func (wg *WorkloadGenerator) SetClusterID(uuid string) *WorkloadGenerator {
	wg.workload.ClusterId = uuid
	return wg
}

func (wg *WorkloadGenerator) SetNameAndUUID(name string) *WorkloadGenerator {
	wg.workload.Name = name
	wg.workload.Id = fmt.Sprintf("%s-id", name)
	return wg
}

func (wg *WorkloadGenerator) SetProject(id string) *WorkloadGenerator {
	wg.workload.ProjectId = id
	wg.workload.ProjectName = fmt.Sprintf("p%s", id)
	wg.workload.Namespace = fmt.Sprintf("ns-%s", id)
	return wg
}

func (wg *WorkloadGenerator) SetPhase(phase openapi.Phase) *WorkloadGenerator {
	wg.workload.Phase = phase
	return wg
}

func (wg *WorkloadGenerator) Build() openapi.Workload {
	wg.workload.CreatedAt = time.Now()
	return wg.workload
}
