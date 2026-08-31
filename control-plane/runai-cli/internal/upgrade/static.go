/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package upgrade delete static.go file once MSI is fixed
package upgrade

import (
	"context"
)

type StaticMessage struct {
	msg string
}

func NewStaticMessage(msg string) *StaticMessage {
	return &StaticMessage{msg: msg}
}

func (s StaticMessage) Upgrade(_ context.Context, _ string) (string, error) {
	return s.msg, nil
}
