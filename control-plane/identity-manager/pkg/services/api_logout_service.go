/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type LogoutAPIService struct {
	iamServices iam.IamServiceContainer
}

func NewLogoutAPIService(iamServices *iam.IamServiceContainer) api.LogoutAPIServicer {
	return &LogoutAPIService{
		iamServices: *iamServices,
	}
}

func (s *LogoutAPIService) Logout(ctx context.Context) (api.ImplResponse, error) {
	realm := utils.GetStringFromCtx(ctx, utils.AuthRealmKey)
	sessionId := utils.GetStringFromCtx(ctx, runaictx.CfSId)
	if sessionId == "" {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("logout user failed. missing session id")
		return api.Response(http.StatusInternalServerError, nil), utils.Error{Code: http.StatusInternalServerError, Message: "logout user failed. Missing session id"}
	}

	err := s.iamServices.UserService.LogoutUserBySessionID(ctx, realm, sessionId)
	if err != nil {
		return utils.HandleError(ctx, utils.User, utils.SessionID, sessionId, err, fmt.Sprintf(`failed logging out user with session id: %s`, sessionId))
	}

	cookie := utils.FormatClearCookie(utils.RefreshTokenCookieName, utils.RefreshTokenCookiePath)
	headers := utils.CreateSetCookieHeaders(cookie)

	return api.ResponseWithHeaders(http.StatusOK, headers, nil), nil
}
