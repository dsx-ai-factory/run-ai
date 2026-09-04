/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package setting

import (
	"errors"
	"fmt"
	"time"

	errHandler "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/error"
	manager "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting/manager"
)

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

func (t TestSetting1) Stage() manager.Stage {
	return manager.Alpha
}

func (t TestSetting1) Validate(value interface{}) bool {
	if ts, ok := value.(*TestSetting1); ok {
		return len(ts.Name) > 0 && ts.Count >= 0
	}
	if ts, ok := value.(TestSetting1); ok {
		return len(ts.Name) > 0 && ts.Count >= 0
	}
	return false
}

func (t TestSetting1) TransformForMigration(value interface{}) interface{} {
	return value
}

////////////////////////////////////////////////////////////////////////////////////////////////

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

func (t TestSetting2) Stage() manager.Stage {
	return manager.Beta
}

func (t TestSetting2) Validate(value interface{}) bool {
	_, ok := value.(*TestSetting2)
	if ok {
		return true
	}
	_, ok = value.(TestSetting2)
	return ok
}

func (t TestSetting2) TransformForMigration(value interface{}) interface{} {
	return value
}

///////////////////////////////////////////////////////////////////////////////////////////////

type TestSetting3 struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

func (t TestSetting3) Key() string {
	return "test-setting3"
}

func (t *TestSetting3) SetDefault() {
	t.Enabled = true
	t.Message = "default message"
}

func (t TestSetting3) DisplayName() string {
	return "Test Setting 3"
}

func (t TestSetting3) Description() string {
	return "A test setting without transformation support"
}

func (t TestSetting3) Category() string {
	return "other"
}

func (t TestSetting3) Stage() manager.Stage {
	return manager.Beta
}

func (t TestSetting3) Validate(value interface{}) bool {
	_, ok := value.(*TestSetting3)
	if ok {
		return true
	}
	_, ok = value.(TestSetting3)
	return ok
}

////////////////////////////////////////////////////////////////////////////////////////////////

func generateUniqueTenantID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func isBadRequest(err error) bool {
	var badReqErr errHandler.ErrorBadRequest
	return errors.As(err, &badReqErr)
}

func isInternalServerError(err error) bool {
	var internalErr errHandler.ErrorInternal
	return errors.As(err, &internalErr)
}
