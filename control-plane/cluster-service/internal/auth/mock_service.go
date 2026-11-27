/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package auth holds a permissive authorization client for integration tests,
// where no authorization service is deployed.
package auth

import (
	"context"
	"net/http"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	auth_mocks "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

// MockService authorizes every request that carries a subject id. Methods that
// are not overridden come from the embedded mockery mock and fail loudly when
// called, so a test that needs one must add it here.
type MockService struct {
	*auth_mocks.IAuthorizationClient
}

func (m MockService) AuthorizeMiddleware(next http.Handler, _ string, _ string, _ bool, _ string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subjectID, err := runaictx.ParseSubjectId(r.Context())
		if err != nil || subjectID == "" {
			http.Error(w, `{"message": "forbidden"}`, http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), structs.PermittedScopesKey, &authorization_api.PermittedScopes{
			System: ptr.To(runaictx.IsManagementTenant(r.Context())),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (m MockService) GetPermittedScopes(_ context.Context, _ authorization_api.ResourceType, _ *authorization_api.Action) (*authorization_api.PermittedScopesActions, int, error) {
	return &authorization_api.PermittedScopesActions{
		Read: authorization_api.PermittedScopes{System: ptr.To(true)},
	}, 0, nil
}

func (m MockService) GetGroupsForTenant(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (m MockService) SetServiceAccountPermissions(_ context.Context, _ string, _ string, _ authorization_api.InternalCreateAccessRuleRequestScope, _ authorization_api.AllowedInternalCreateRoles) error {
	return nil
}
