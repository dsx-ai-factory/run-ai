/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	t.Setenv(envConfigKeyDelimiter, "::")
	conf := getConfig()
	assert.Equal(t, "::", conf.KeyDelimiter)
}

func TestGetConfig_Defaults(t *testing.T) {
	conf := getConfig()
	assert.Equal(t, defaultKeyDelimiter, conf.KeyDelimiter)
}

func TestInit(t *testing.T) {
	reset()
	conf := &Config{
		KeyDelimiter: "##",
	}

	err := Init(conf)
	assert.NoError(t, err)
}

func TestInit_InvalidKeyDelimiter(t *testing.T) {
	reset()
	conf := &Config{
		KeyDelimiter: " ",
	}

	err := Init(conf)
	assert.Equal(t, err.Error(), "Key: 'Config.KeyDelimiter' Error:Field validation for 'KeyDelimiter' failed on the 'excludesall' tag")
}

func reset() {
	once = sync.Once{}
	globalRegistry = newRegistry()
}
