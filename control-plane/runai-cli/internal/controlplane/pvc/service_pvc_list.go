/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	"context"
	"fmt"

	openapi "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
)

const (
	ListAPIError = "list API call error: %w"
)

type Lister interface {
	List(ctx context.Context, projectId string) (*openapi.PVCListResponse, error)
}

type Service struct {
	client Lister
}

func NewListerService(client Lister) *Service {
	return &Service{client: client}
}

func (pls *Service) ListPvcs(ctx context.Context, projectId string) (*openapi.PVCListResponse, error) {
	list, err := pls.client.List(ctx, projectId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (pls *Service) ListPvcsBrief(ctx context.Context, projectId string) (*Output, error) {
	list, err := pls.ListPvcs(ctx, projectId)
	if err != nil {
		return nil, fmt.Errorf(ListAPIError, err)
	}
	return NewOutput(list)
}
