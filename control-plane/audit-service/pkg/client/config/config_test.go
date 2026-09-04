/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
)

func TestInitConfig_PassedConfigIsNotNil_SetsPassedConfig(t *testing.T) {
	err := InitConfig(&Config{
		Enabled:                      false,
		EnablePrintAuditLogsToStdout: true,
		AuditServiceUrl:              "custom-url",
	})
	assert.NoError(t, err)

	conf := config.Get[*Config]()
	assert.False(t, conf.Enabled)
	assert.True(t, conf.EnablePrintAuditLogsToStdout)
	assert.Equal(t, "custom-url", conf.AuditServiceUrl)
}

func TestInitConfig_PassedConfigIsNil_RegistersDefaults(t *testing.T) {
	err := InitConfig(nil)
	assert.NoError(t, err)

	conf := config.Get[*Config]()
	assert.True(t, conf.Enabled)
	assert.False(t, conf.EnablePrintAuditLogsToStdout)
	assert.Equal(t, "http://runai-backend-audit-service:8080", conf.AuditServiceUrl)
}

func TestInitConfig_PassedConfigIsNil_RegistersCustomValues(t *testing.T) {
	t.Setenv("AUDIT_ENABLED", "false")
	t.Setenv("AUDIT_ENABLE_PRINT_TO_STDOUT", "true")
	t.Setenv("AUDIT_SERVICE_URL", "custom-url")

	err := InitConfig(nil)
	assert.NoError(t, err)

	conf := config.Get[*Config]()
	assert.False(t, conf.Enabled)
	assert.True(t, conf.EnablePrintAuditLogsToStdout)
	assert.Equal(t, "custom-url", conf.AuditServiceUrl)
}
