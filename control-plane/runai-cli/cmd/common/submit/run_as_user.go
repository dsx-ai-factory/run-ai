/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"os"
	"strconv"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"go.uber.org/zap"
)

func calculateSecurityRunAsUserSettings(token string) (int64, int64, []int) {
	logger := instance.GetLogger()
	claims, err := auth.GetClaims(token)
	if err != nil {
		logger.Debug("failed to get claims from token, use local credentials", zap.Error(err))
		return getSecurityRunAsUserSettingsFromOs()
	}
	return getSecurityRunAsUserSettingsFromJWT(claims)
}

func getSecurityRunAsUserSettingsFromJWT(claims *auth.CustomClaims) (int64, int64, []int) {
	logger := instance.GetLogger()
	if claims.Uid == "" && claims.Gid == "" && claims.SupplementaryGroups == nil {
		logger.Debug("claims from token are empty, use local credentials")
		return getSecurityRunAsUserSettingsFromOs()
	}

	logger.Debug("using claims from token",
		zap.String("UID", claims.Uid),
		zap.String("GID", claims.Gid),
		zap.Any("SupplementaryGroups", claims.SupplementaryGroups))

	uid, err := strconv.Atoi(claims.Uid)
	if err != nil {
		// if the uid is not a number, we will not use it
		uid = 0
	}

	gid, err := strconv.Atoi(claims.Gid)
	if err != nil {
		// if the gid is not a number, we will not use it
		gid = 0
	}

	return int64(uid), int64(gid), claims.SupplementaryGroups
}

func getSecurityRunAsUserSettingsFromOs() (int64, int64, []int) {
	logger := instance.GetLogger()
	uid, gid := int64(os.Getuid()), int64(os.Getgid())
	supplementaryGroups, err := os.Getgroups()
	if err != nil {
		logger.Error("failed to get os supplementary groups", zap.Error(err))
	}
	return uid, gid, supplementaryGroups
}

func setSecurityRunAsUserSettings(uid, gid int64, supplementaryGroups []int) {
	Spec.SetSecurityRunAsUserSettings(uid, gid, supplementaryGroups)
}
