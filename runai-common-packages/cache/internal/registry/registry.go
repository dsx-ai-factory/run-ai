/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package registry

import (
	"sync"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
)

type storeRegistry struct {
	storeConfigurations sync.Map
	defaultStore        *options.Store
}

func NewStoreRegistry(defaultStore *options.Store) *storeRegistry {
	return &storeRegistry{
		storeConfigurations: sync.Map{},
		defaultStore:        defaultStore,
	}
}

func (s *storeRegistry) Get(key string) *options.Store {
	if store, ok := s.storeConfigurations.Load(key); ok {
		return store.(*options.Store)
	}

	return nil
}

func (s *storeRegistry) GetOrDefault(key string) *options.Store {
	if store, ok := s.storeConfigurations.Load(key); ok {
		return store.(*options.Store)
	}

	return s.defaultStore
}

func (s *storeRegistry) Set(key string, store *options.Store) {
	s.storeConfigurations.Store(key, store)
}

func (s *storeRegistry) List() []*options.Store {
	var configs []*options.Store
	if s.defaultStore != nil {
		configs = append(configs, s.defaultStore)
	}

	s.storeConfigurations.Range(func(_, value any) bool {
		configs = append(configs, value.(*options.Store))
		return true
	})
	return configs
}
