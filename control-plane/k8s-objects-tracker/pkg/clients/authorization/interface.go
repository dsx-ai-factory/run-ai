/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization

import (
	"context"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
)

type ClientInterface interface {
	// CheckPermittedScopes returns the scopes in which actions are permitted on a given resource type
	// action is optional, if given - checking will be done only for the specific action.
	CheckPermittedScopes(ctx context.Context, resourceType authorizationClient.ResourceType, action *authorizationClient.Action) (*authorizationClient.PermittedScopesActions, *api.Error)
}
