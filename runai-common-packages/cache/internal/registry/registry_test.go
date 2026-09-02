/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package registry

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/stretchr/testify/require"
)

var key1 = "users"
var key2 = "tenants"

func TestGet_NotExist(t *testing.T) {
	s := NewStoreRegistry(nil)
	require.Nil(t, s.Get(key1))
}

func TestSetGet_Exist(t *testing.T) {
	s := NewStoreRegistry(nil)
	s.Set(key1, &options.Store{Name: "service-a"})
	conf := s.Get(key1)
	require.NotNil(t, conf)
	require.Equal(t, conf.Name, "service-a")
}

func TestSetGet_Default(t *testing.T) {
	initConf(t)
	s := NewStoreRegistry(nil)
	s.Set(key1, &options.Store{
		Name:    "some_name",
		Options: options.NewStoreOptions(),
	})
	conf := s.Get(key1)
	require.NotNil(t, conf)
	require.Equal(t, conf.Options.Ttl, time.Hour)
}

func TestList(t *testing.T) {
	s := NewStoreRegistry(nil)
	s.Set(key1, &options.Store{Name: "service-1"})
	s.Set(key2, &options.Store{Name: "service-2"})
	conf := s.List()
	require.Equal(t, len(conf), 2)
	require.ElementsMatch(t, []string{conf[0].Name, conf[1].Name}, []string{"service-2", "service-1"})
}

func initConf(t *testing.T) {
	t.Setenv("SERVICE_NAME", "test-service")
	t.Setenv("SECRETS_MOUNT_PATH", ".")
	t.Setenv("CACHE_TYPE", "inmemory")
	t.Setenv("CACHE_URL", "localhost:1234")
	require.NoError(t, config.InitConfig())
}
