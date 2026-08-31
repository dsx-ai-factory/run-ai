/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils/pointer"
	"k8s.io/utils/ptr"
)

func FormatDate(date string) *time.Time {
	if date == "" {
		return nil
	}
	unixTime, err := EpochStringToTime(date)
	if err != nil {
		return nil
	}
	return &unixTime
}

func EpochStringToTime(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	now := time.Now().Unix()
	if now >= sec {
		return time.Unix(sec, 0), nil
	}
	return time.Time{}, fmt.Errorf("bad format of epoch time %v. time should be in unix seconds", sec)
}

func StringToBool(s string) *bool {
	if s == "" {
		return nil
	}
	switch strings.ToLower(s) {
	case "true":
		return pointer.ToBool(true)
	case "false":
		return pointer.ToBool(false)
	default:
		return nil
	}
}

func SplitCommaSeparated(input string) *[]string {
	if input == "" {
		return nil
	}
	values := strings.Split(input, ",")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	return &values
}

func UnixTimestampToString() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func StringToTimeDuration(str *string) *time.Duration {
	if str == nil {
		return nil
	}
	seconds, err := strconv.ParseInt(*str, 10, 64)
	if err != nil {
		return nil
	}
	duration := time.Duration(seconds) * time.Second
	return ptr.To(duration)
}

func ParseEpochToTime(epochStr string) *time.Time {
	if epochStr == "" {
		return nil
	}
	epoch, err := strconv.ParseInt(epochStr, 10, 64)
	if err != nil {
		return nil
	}
	t := time.Unix(epoch, 0)
	return &t
}
