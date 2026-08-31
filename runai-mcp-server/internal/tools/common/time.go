/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"fmt"
	"strings"
	"time"
)

func TimestampString(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func ParseRFC3339IfSet(field, value string) (t time.Time, set bool, err error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false, nil
	}
	if t, err = time.Parse(time.RFC3339, value); err != nil {
		return time.Time{}, false, fmt.Errorf("%s must be RFC3339 (got %q)", field, value)
	}
	return t, true, nil
}

func ValidateTimeWindow(start, end string) error {
	startT, startSet, err := ParseRFC3339IfSet("start", start)
	if err != nil {
		return err
	}
	endT, endSet, err := ParseRFC3339IfSet("end", end)
	if err != nil {
		return err
	}
	if startSet && endSet && !endT.After(startT) {
		return fmt.Errorf("end must be after start")
	}
	return nil
}
