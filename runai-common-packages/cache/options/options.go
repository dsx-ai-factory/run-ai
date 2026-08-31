/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package options

import (
	"time"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/config"
)

const (
	MinStoreSizeMBLimit  = 20
	MaxStoreSizeMBLimit  = 4000
	MaxRecordSizeKBLimit = 30000
)

type Option func(*StoreOptions)

func NewStoreOptions(options ...Option) *StoreOptions {
	s := &StoreOptions{}

	for _, opt := range options {
		opt(s)
	}

	applyDefaultStoreOptions(s)

	return s
}

type StoreOptions struct {
	Ttl             time.Duration
	MaxStoreSizeMB  int
	MaxRecordSizeKB int
	ReadOnly        bool
}

func Ttl(t time.Duration) Option {
	return func(so *StoreOptions) {
		so.Ttl = t
	}
}

func MaxStoreSizeMB(size int) Option {
	return func(so *StoreOptions) {
		so.MaxStoreSizeMB = max(MinStoreSizeMBLimit, min(size, MaxStoreSizeMBLimit))
	}
}

func MaxRecordSizeKB(size int) Option {
	return func(so *StoreOptions) {
		so.MaxRecordSizeKB = min(size, MaxRecordSizeKBLimit)
	}
}

func ReadOnly() Option {
	return func(so *StoreOptions) {
		so.ReadOnly = true
	}
}

func applyDefaultStoreOptions(so *StoreOptions) {
	cacheConf := config.Get[*cache_config.Cache]()
	if so.MaxStoreSizeMB == 0 {
		so.MaxStoreSizeMB = cacheConf.MaxStoreSizeMB
	}
	if so.MaxRecordSizeKB == 0 {
		so.MaxRecordSizeKB = cacheConf.MaxRecordSizeKB
	}
	if so.Ttl == 0 {
		so.Ttl = time.Duration(cacheConf.TtlSeconds) * time.Second
	}
}
