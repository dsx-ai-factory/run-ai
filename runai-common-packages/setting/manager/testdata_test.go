/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manager

import (
	"errors"

	errHandler "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/error"
)

// ============================================================================
// SHARED TEST MOCK DATA
// ============================================================================
type TestSetting1 struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func (t TestSetting1) Key() string {
	return "test-setting1"
}

func (t *TestSetting1) SetDefault() {
	t.Name = "default name"
	t.Count = 1
}

func (t TestSetting1) DisplayName() string {
	return "Test Setting 1"
}

func (t TestSetting1) Description() string {
	return "A test setting for unit tests"
}

func (t TestSetting1) Category() string {
	return "test"
}

func (t TestSetting1) Stage() Stage {
	return Alpha
}

func (t TestSetting1) Validate(value interface{}) bool {
	if ts, ok := value.(TestSetting1); ok {
		return len(ts.Name) > 0 && ts.Count >= 0
	}
	return false
}

func (t TestSetting1) TransformForMigration(value interface{}) interface{} {
	return value
}

// ============================================================================

type TestSetting2 struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

func (t TestSetting2) Key() string {
	return "test-setting2"
}

func (t *TestSetting2) SetDefault() {
	t.Enabled = true
	t.Message = "default message"
}

func (t TestSetting2) DisplayName() string {
	return "Test Setting 2"
}

func (t TestSetting2) Description() string {
	return "Another test setting for unit tests"
}

func (t TestSetting2) Category() string {
	return "other"
}

func (t TestSetting2) Stage() Stage {
	return Beta
}

func (t TestSetting2) Validate(value interface{}) bool {
	_, ok := value.(TestSetting2)
	return ok
}

// ============================================================================
type TestSettingWithTransform struct {
	Value string `json:"value"`
}

func (t TestSettingWithTransform) Key() string {
	return "test-setting-transform"
}

func (t *TestSettingWithTransform) SetDefault() {
	t.Value = "default"
}

func (t TestSettingWithTransform) DisplayName() string {
	return "Test Setting With Transform"
}

func (t TestSettingWithTransform) Description() string {
	return "A test setting with transformation"
}

func (t TestSettingWithTransform) Category() string {
	return "test"
}

func (t TestSettingWithTransform) Stage() Stage {
	return Alpha
}

func (t TestSettingWithTransform) Validate(value interface{}) bool {
	_, ok := value.(TestSettingWithTransform)
	return ok
}

func (t TestSettingWithTransform) TransformForMigration(value interface{}) interface{} {
	if ts, ok := value.(TestSettingWithTransform); ok {
		return TestSettingWithTransform{Value: "transformed-" + ts.Value}
	}
	return value
}

// ============================================================================
// SHARED TEST UTILITIES
// ============================================================================

func isBadRequest(err error) bool {
	var badReqErr errHandler.ErrorBadRequest
	return errors.As(err, &badReqErr)
}

func isInternalServerError(err error) bool {
	var internalErr errHandler.ErrorInternal
	return errors.As(err, &internalErr)
}
