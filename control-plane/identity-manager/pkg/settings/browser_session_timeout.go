/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/setting/manager"
)

const DefaultIdleTimeoutInMinutes float32 = 0

type BrowserSessionTimeout struct {
	IdleTimeout *float32 `json:"idleTimeout"`
}

func NewBrowserSessionTimeout() manager.Settinger {
	return &BrowserSessionTimeout{}
}

func (b *BrowserSessionTimeout) Category() string {
	return utils.SecurityCategory
}

func (b *BrowserSessionTimeout) Description() string {
	return "The idle timeout duration in minutes for browser sessions"
}

func (b *BrowserSessionTimeout) DisplayName() string {
	return "Browser Session Timeout"
}

func (b *BrowserSessionTimeout) Key() string {
	return string(api.SETTINGSKEYENUM_BROWSER_SESSION_TIMEOUT)
}

func (b *BrowserSessionTimeout) SetDefault() {
	defaultTimeout := DefaultIdleTimeoutInMinutes
	b.IdleTimeout = &defaultTimeout
}

func (b *BrowserSessionTimeout) Stage() manager.Stage {
	return manager.GA
}

func (b *BrowserSessionTimeout) Validate(value interface{}) bool {
	var valueToUpdate BrowserSessionTimeout

	switch v := value.(type) {
	case BrowserSessionTimeout:
		valueToUpdate = v
	case *BrowserSessionTimeout:
		valueToUpdate = *v
	default:
		return false
	}

	return valueToUpdate.IdleTimeout != nil && *valueToUpdate.IdleTimeout >= 0
}
