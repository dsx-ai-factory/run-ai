/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import "github.com/go-resty/resty/v2"

type AuthConfigurator interface {
	Configure(*resty.Request) *resty.Request
}

type EmptyAuthConfigurator struct{}

func (e *EmptyAuthConfigurator) Configure(req *resty.Request) *resty.Request {
	return req
}

type BasicAuthConfigurator struct {
	Username string
	Password string
}

func (b *BasicAuthConfigurator) Configure(req *resty.Request) *resty.Request {
	if b == nil {
		return req
	}
	return req.SetBasicAuth(b.Username, b.Password)
}
