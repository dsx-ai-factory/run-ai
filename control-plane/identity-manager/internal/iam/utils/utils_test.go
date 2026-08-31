/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEpochStringToTime_HappyFlow(t *testing.T) {
	res, _ := EpochStringToTime("1638316800")
	assert.Equal(t, time.Unix(1638316800, 0), res)
}

func TestEpochStringToTime_Error(t *testing.T) {
	res, err := EpochStringToTime("invalid")
	assert.Empty(t, res)
	assert.NotNil(t, err)
}

func TestFormatDate_HappyFlow(t *testing.T) {
	res := FormatDate("1638316800")
	assert.Equal(t, time.Unix(1638316800, 0), *res)
	assert.NotNil(t, res)
}

func TestFormatDate_EmptyDate(t *testing.T) {
	res := FormatDate("")
	assert.Nil(t, res)
}

func TestFormatDate_EpochStringToTime_Error(t *testing.T) {
	res := FormatDate("invalid")
	assert.Nil(t, res)
}

func TestStringToBool_HappyFlow_true(t *testing.T) {
	res := StringToBool("true")
	assert.True(t, *res)
}

func TestStringToBool_HappyFlow_false(t *testing.T) {
	res := StringToBool("false")
	assert.False(t, *res)
}

func TestStringToBool_EmptyString(t *testing.T) {
	res := StringToBool("")
	assert.Nil(t, res)
}

func TestStringToBool_default(t *testing.T) {
	res := StringToBool("invalid")
	assert.Nil(t, res)
}

func TestSplitCommaSeparated_StringWithComma(t *testing.T) {
	res := SplitCommaSeparated("groupA,groupB,groupC")
	assert.Equal(t, "groupA", (*res)[0])
	assert.Equal(t, "groupB", (*res)[1])
	assert.Equal(t, "groupC", (*res)[2])
	assert.Equal(t, 3, len(*res))
}

func TestSplitCommaSeparated_StringWithDots(t *testing.T) {
	res := SplitCommaSeparated("groupA.groupB.groupC")
	assert.Equal(t, "groupA.groupB.groupC", (*res)[0])
	assert.Equal(t, 1, len(*res))
}

func TestSplitCommaSeparated_EmptyString(t *testing.T) {
	res := SplitCommaSeparated("")
	assert.Nil(t, res)
}

func TestStringToTimeDuration_HappyFlow(t *testing.T) {
	str := "3600" // 1 hour in seconds
	res := StringToTimeDuration(&str)
	expected := time.Hour
	assert.NotNil(t, res)
	assert.Equal(t, expected, *res)
}

func TestStringToTimeDuration_ZeroSeconds(t *testing.T) {
	str := "0"
	res := StringToTimeDuration(&str)
	expected := time.Duration(0)
	assert.NotNil(t, res)
	assert.Equal(t, expected, *res)
}

func TestStringToTimeDuration_SmallValue(t *testing.T) {
	str := "30" // 30 seconds
	res := StringToTimeDuration(&str)
	expected := 30 * time.Second
	assert.NotNil(t, res)
	assert.Equal(t, expected, *res)
}

func TestStringToTimeDuration_LargeValue(t *testing.T) {
	str := "86400" // 24 hours in seconds
	res := StringToTimeDuration(&str)
	expected := 24 * time.Hour
	assert.NotNil(t, res)
	assert.Equal(t, expected, *res)
}

func TestStringToTimeDuration_NilInput(t *testing.T) {
	res := StringToTimeDuration(nil)
	assert.Nil(t, res)
}

func TestStringToTimeDuration_EmptyString(t *testing.T) {
	str := ""
	res := StringToTimeDuration(&str)
	assert.Nil(t, res)
}

func TestStringToTimeDuration_InvalidString(t *testing.T) {
	str := "invalid"
	res := StringToTimeDuration(&str)
	assert.Nil(t, res)
}

func TestStringToTimeDuration_FloatString(t *testing.T) {
	str := "3600.5"
	res := StringToTimeDuration(&str)
	assert.Nil(t, res)
}

func TestStringToTimeDuration_NegativeValue(t *testing.T) {
	str := "-3600"
	res := StringToTimeDuration(&str)
	expected := -time.Hour
	assert.NotNil(t, res)
	assert.Equal(t, expected, *res)
}

func TestParseEpochToTime_HappyFlow(t *testing.T) {
	res := ParseEpochToTime("1638316800")
	assert.Equal(t, time.Unix(1638316800, 0), *res)
}

func TestParseEpochToTime_EmptyString(t *testing.T) {
	res := ParseEpochToTime("")
	assert.Nil(t, res)
}

func TestParseEpochToTime_InvalidString(t *testing.T) {
	res := ParseEpochToTime("invalid")
	assert.Nil(t, res)
}
