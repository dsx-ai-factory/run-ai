/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toggles

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/smithy-go/ptr"
	"github.com/run-ai/runai/runai-common-packages/config"
)

var (
	once     sync.Once
	instance Toggler = &dummyToggler{}
	log      Logger  = &dummyLogger{}
	initErr  error
)

func Init(cfg *Config) error {
	once.Do(func() {
		if err := initConfig(cfg); err != nil {
			initErr = fmt.Errorf("toggles: failed to initialize config: %w", err)
			instance = &failedToggler{}
			return
		}
		loadedConf := config.Get[*Config]()
		if !ptr.ToBool(loadedConf.Enabled) {
			return
		}

		ut, err := newUnleashToggler()
		if err != nil {
			initErr = fmt.Errorf("toggles: could not create unleash toggler: %w", err)
			instance = &failedToggler{}
			return
		}
		instance = ut
	})
	return initErr
}

func IsEnabled(ctx context.Context, name string) bool {
	return instance.IsEnabled(ctx, name)
}

func Close() error {
	err := instance.Close()
	instance = &dummyToggler{}
	once = sync.Once{}
	initErr = nil
	return err
}

func SetLogger(l Logger) {
	if l != nil {
		log = l
	}
}

func SetMock(mock Toggler) {
	instance = mock
}

func initConfig(cfg *Config) error {
	if cfg == nil {
		return registerConfig()
	}
	if setErr := config.SetWithDefaults(cfg); setErr != nil {
		return fmt.Errorf("toggles: failed to set config: %w", setErr)
	}
	return nil
}
