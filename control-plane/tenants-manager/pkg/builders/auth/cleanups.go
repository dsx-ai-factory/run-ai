/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"net/http"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

func (b *Builder) deleteRunaiRealm(ctx context.Context, realmName string) (*http.Response, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Deleting realm %s", realmName)
	resp, err := b.idmClient.DeleteRealm(ctx, realmName, true)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to Delete runai realm %s", realmName)
		return resp, err
	}
	return resp, nil
}

func (b *Builder) deleteAccessRules(ctx context.Context, userEmail string) (*http.Response, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Deleting access rules for user %s", userEmail)
	request := b.authClient.DeleteSubjectAccessRules(ctx)
	request = request.SubjectId(userEmail)
	request = request.SubjectType(string(authorization_api.SUBJECTTYPE_USER))
	resp, err := b.authClient.DeleteSubjectAccessRulesExecute(request)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to delete access rules for user %s. Response: %+v, Error: %+v", userEmail, resp, err)
		return resp, err
	}
	return resp, nil
}
