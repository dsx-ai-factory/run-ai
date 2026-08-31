/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"net/http"
)

type NimServicesService struct {
	impl *NimServicesImpl
}

func NewNimServicesService() api.NVIDIANIMAPIServicer {
	return &NimServicesService{
		impl: NewNimServicesImpl(NewWorkloadsV2Impl()),
	}
}

// CreateNimService - Create a NIM service workload
func (nss *NimServicesService) CreateNimService(ctx context.Context, request api.NimServiceCreateRequest) (api.ImplResponse, error) {
	payload, err := nss.impl.CreateNimService(ctx, request)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create NIM service %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}

func (nss *NimServicesService) GetNimServiceById(ctx context.Context, id string) (api.ImplResponse, error) {
	payload, err := nss.impl.GetNimService(ctx, id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get NIM service %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusOK, *payload), nil
}

func (nss *NimServicesService) UpdateNimServiceSpec(ctx context.Context, id string, request api.NimServiceUpdateRequest) (api.ImplResponse, error) {
	payload, err := nss.impl.UpdateNimServiceSpec(ctx, id, request)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to update NIM service %s: %v", err.Message, err.Details)
		return api.Response(int(err.Code), err), nil
	}

	return api.Response(http.StatusAccepted, payload), nil
}
