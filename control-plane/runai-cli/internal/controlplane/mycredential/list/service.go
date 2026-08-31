/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
)

const (
	APIError = "my-credential API call error: %w"
)

type Lister interface {
	ListUserCredentials(ctx context.Context, typeFilter string) ([]client.MyCredentialEntry, error)
}

type Service struct {
	client Lister
}

func NewService(client Lister) *Service {
	return &Service{client: client}
}

func (s *Service) List(ctx context.Context, typeFilter string) (*Output, error) {
	entries, err := s.client.ListUserCredentials(ctx, typeFilter)
	if err != nil {
		return nil, fmt.Errorf(APIError, err)
	}
	return NewOutput(entries), nil
}
