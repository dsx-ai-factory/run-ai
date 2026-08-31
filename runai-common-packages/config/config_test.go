/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config_test

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	TestVal1 string   `mapstructure:"test_val1"`
	TestVal2 string   `mapstructure:"test_val2"`
	TestVal3 []string `mapstructure:"test_val3"`
}

func (t *TestConfig) SetDefaults() {
	t.TestVal1 = "test"
}

func Test_GetWithoutRegistering(t *testing.T) {
	config.FlushAll()
	c := config.Get[*TestConfig]()

	assert.NotNil(t, c)
	assert.Equal(t, "", c.TestVal1)
}

func TestRegister_Registers(t *testing.T) {
	config.Register[*TestConfig]()
	c := config.Get[*TestConfig]()

	assert.NotNil(t, c)
}

func TestGet_WithSlice(t *testing.T) {
	config.FlushAll()
	t.Setenv("TEST_VAL3", "val1,val2,val3")
	config.Register[*TestConfig]()
	c := config.Get[*TestConfig]()

	assert.NotNil(t, c)
	assert.Equal(t, []string{"val1", "val2", "val3"}, c.TestVal3)
}

func TestGet_WithEnvVarsUpper(t *testing.T) {
	config.FlushAll()
	t.Setenv("TEST_VAL1", "bla")
	config.Register[*TestConfig]()
	c := config.Get[*TestConfig]()

	assert.NotNil(t, c)
	assert.Equal(t, "bla", c.TestVal1)
}

func TestGet_WithDefaultValues(t *testing.T) {
	config.Register[*TestConfig]()
	c := config.Get[*TestConfig]()

	assert.NotNil(t, c)
	assert.Equal(t, "test", c.TestVal1)
}

type TestConfigNoMaps struct {
	TestVal1 string
}

func (t *TestConfigNoMaps) SetDefaults() {
	t.TestVal1 = "test"
}

func TestGet_ConfigWithoutMapstructure(t *testing.T) {
	config.FlushAll()
	t.Setenv("TESTVAL1", "val")
	config.Register[*TestConfigNoMaps]()
	c := config.Get[*TestConfigNoMaps]()

	assert.Equal(t, "val", c.TestVal1)
}

func TestGetR_GetAndRegisters(t *testing.T) {
	config.FlushAll()
	t.Setenv("TESTVAL1", "val")
	config, err := config.GetR[*TestConfigNoMaps]()

	assert.Nil(t, err)
	assert.Equal(t, "val", config.TestVal1)
}

func TestGetR_GetAndRegistersDefault(t *testing.T) {
	config.FlushAll()
	c, err := config.GetR[*TestConfigNoMaps]()

	assert.Nil(t, err)
	assert.Equal(t, "test", c.TestVal1)
}

func TestGetR_GetsRegistered(t *testing.T) {
	config.FlushAll()
	t.Setenv("TESTVAL1", "val")
	config.Register[*TestConfigNoMaps]()
	c, err := config.GetR[*TestConfigNoMaps]()

	assert.Nil(t, err)
	assert.Equal(t, "val", c.TestVal1)
}
