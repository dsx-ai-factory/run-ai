/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCronJobSchedule_Seconds(t *testing.T) {
	expression := getCronJobSchedule(30)
	assert.Equal(t, "30 seconds", expression)
}

func TestGetCronJobSchedule_Minutes(t *testing.T) {
	expression := getCronJobSchedule(20 * 60)
	assert.Equal(t, "*/20 * * * *", expression)
}

func TestGetCronJobSchedule_Hours(t *testing.T) {
	expression := getCronJobSchedule(2 * 60 * 60)
	assert.Equal(t, "0 */2 * * *", expression)
}

func TestGetInitialDelay_Zero(t *testing.T) {
	delay := getInitialDelay(time.Time{})
	assert.Equal(t, time.Duration(0), delay)
}

func TestGetInitialDelay_NonZero(t *testing.T) {
	now := time.Now()
	threeAM := time.Date(now.Year(), now.Month(), now.Day(), 3, 0, 0, 0, now.Location())
	delay := getInitialDelay(threeAM)
	assert.NotEqual(t, time.Duration(0), delay)
}
