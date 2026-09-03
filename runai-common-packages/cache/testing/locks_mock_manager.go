/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/internal/lock"

type locksMockManager struct {
	t          TestingT
	mockLocker *MockLocker
}

func newLocksMockManager(t TestingT, mockLocker *MockLocker) *locksMockManager {
	return &locksMockManager{
		t:          t,
		mockLocker: mockLocker,
	}
}

func (m *locksMockManager) CreateLock(name string, shared bool) lock.Locker {
	return m.mockLocker
}
