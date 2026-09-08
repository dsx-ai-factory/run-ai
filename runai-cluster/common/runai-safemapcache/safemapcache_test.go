/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package safemapcache

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	key   = "key"
	value = "value"
)

func TestStoreValue(t *testing.T) {
	smc := New[string, string]()
	smc.Store(key, value)

	storedValue, found, valid := smc.Load(key)
	assert.Equal(t, true, found, "expected value to be stored")
	assert.Equal(t, true, valid, "expected stored value to be valid")
	assert.Equal(t, storedValue, value, "expected stored value to be equal to original value")
}

func TestDeleteValue(t *testing.T) {
	smc := New[string, string]()
	smc.Store(key, value)
	smc.Delete(key)
	_, found, _ := smc.Load(key)
	assert.Equal(t, false, found, "expected value to not be stored")
}

func TestExpiredValue(t *testing.T) {
	smc := NewWithPeriods[string, string](time.Hour, time.Nanosecond)
	smc.Store(key, value)

	time.Sleep(2 * time.Nanosecond)

	storedValue, found, valid := smc.Load(key)
	assert.Equal(t, true, found, "expected value to be stored")
	assert.Equal(t, false, valid, "expected stored value to be invalid")
	assert.Equal(t, storedValue, value, "expected stored value to be equal to original value")
}

func TestDeletedValue(t *testing.T) {
	smc := New[string, string]()
	smc.Store(key, value)
	smc.Delete(key)
	_, found, _ := smc.Load(key)
	assert.Equal(t, false, found, "expected value to not be stored")
}

func TestClear(t *testing.T) {
	smc := NewWithPeriods[string, string](time.Hour, time.Hour)
	smc.Store(key, value)
	smc.Store(key+"_2", value)
	smc.Store(key+"_3", value)

	smc.Clear()

	_, found, _ := smc.Load(key)
	assert.Equal(t, false, found, fmt.Sprintf("expected value of %s to be deleted", key))
	_, found, _ = smc.Load(key + "_2")
	assert.Equal(t, false, found, fmt.Sprintf("expected value of %s to be deleted", key+"_2"))
	_, found, _ = smc.Load(key + "_3")
	assert.Equal(t, false, found, fmt.Sprintf("expected value of %s to be deleted", key+"_3"))
}

func TestAutomaticCleanup(t *testing.T) {
	smc := NewWithPeriods[string, string](time.Second/2, time.Second/2)
	smc.Store(key, value)

	time.Sleep(2 * time.Second)

	_, found, _ := smc.Load(key)
	assert.Equal(t, false, found, "expected value to not be stored")
}

func TestCannotStoreInDestroyedCache(t *testing.T) {
	defer func(t *testing.T) {
		err := recover()
		assert.Equal(t, ErrStoreInInactive, err)
	}(t)

	smc := New[string, string]()
	smc.Destroy()
	smc.Store(key, value)
}

func TestCannotLoadFromDestroyedCache(t *testing.T) {
	defer func(t *testing.T) {
		err := recover()
		assert.Equal(t, ErrLoadFromInactive, err)
	}(t)

	smc := New[string, string]()
	smc.Store(key, value)
	smc.Destroy()
	_, _, _ = smc.Load(key)
}

func TestCannotDeleteFromDestroyedCache(t *testing.T) {
	defer func(t *testing.T) {
		err := recover()
		assert.Equal(t, ErrDeleteFromInactive, err)
	}(t)

	smc := New[string, string]()
	smc.Store(key, value)
	smc.Destroy()
	smc.Delete(key)
}

func TestCannotDestroyDestroyedCache(t *testing.T) {
	defer func(t *testing.T) {
		err := recover()
		assert.Equal(t, ErrDestroyInactive, err)
	}(t)

	smc := New[string, string]()
	smc.Store(key, value)
	smc.Destroy()
	smc.Destroy()
}
