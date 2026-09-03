/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpworkloadmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/golang/mock/gomock"
)

type WorkloadPaginationMocker struct {
	client *cpworkloadmocks.MockPaginator
}

func (m *WorkloadPaginationMocker) Set(fb *filter.Builder, limit, offset int32, err error, next *int32, workloads ...openapi.Workload) {
	var response *openapi.GetWorkloads200Response
	if next != nil || len(workloads) > 0 {
		response = &openapi.GetWorkloads200Response{Workloads: workloads, Next: next}
	}
	m.client.EXPECT().List(gomock.Any(), limit, offset, fb.StringList(), false).Return(response, err)
}

func (m *WorkloadPaginationMocker) SetAny(fb *filter.Builder, limit, offset int32, err error, next *int32, workloads ...openapi.Workload) {
	var response *openapi.GetWorkloads200Response
	if next != nil || len(workloads) > 0 {
		response = &openapi.GetWorkloads200Response{Workloads: workloads, Next: next}
	}
	m.client.EXPECT().List(gomock.Any(), limit, offset, fb.StringList(), false).Return(response, err).AnyTimes()
}

func NewWorkloadPaginationMocker(mockCtrl *gomock.Controller) *WorkloadPaginationMocker {
	client := cpworkloadmocks.NewMockPaginator(mockCtrl)
	workload.SetPaginator(client)
	return &WorkloadPaginationMocker{client: client}
}
