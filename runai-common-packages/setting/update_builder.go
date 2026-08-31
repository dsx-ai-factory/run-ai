/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package setting

import (
	"context"

	manager "github.com/run-ai/runai/runai-common-packages/setting/manager"
)

type updateBuilder[T manager.Settinger] struct {
	tenantId     string
	value        interface{}
	updateByUser bool
}

func (ub *updateBuilder[T]) UpdatedByUser() *updateBuilder[T] {
	ub.updateByUser = true
	return ub
}

func (ub *updateBuilder[T]) UpdatedBySystem() *updateBuilder[T] {
	ub.updateByUser = false
	return ub
}

func (ub *updateBuilder[T]) Execute() error {
	if err := checkInitialized(); err != nil {
		return err
	}
	ctx := context.Background()
	setting := initializSetting[T]()
	settingKey := setting.Key()
	return settingsManager.Update(ctx, ub.tenantId, settingKey, ub.updateByUser, ub.value)
}
