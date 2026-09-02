/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type configPool struct {
	mu             sync.RWMutex
	vipers         map[string]*viper.Viper
	onConfigChange func(*viper.Viper)
	keyDelimiter   string
}

func newConfigPool(onConfigChange func(*viper.Viper), delimiter string) *configPool {
	return &configPool{
		vipers:         make(map[string]*viper.Viper),
		onConfigChange: onConfigChange,
		keyDelimiter:   delimiter,
	}
}

func (cp *configPool) AddConfigFiles(paths []string) (*viper.Viper, error) {
	for _, path := range paths {
		err := cp.addConfigFile(path)
		if err != nil {
			return nil, err
		}
	}
	return generateCombinedViper(cp.vipers, cp.keyDelimiter), nil
}

func (cp *configPool) FlushAll() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.vipers = make(map[string]*viper.Viper)
}

func (cp *configPool) addConfigFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to normalize path: %w", err)
	}

	cp.mu.RLock()
	_, exists := cp.vipers[absPath]
	cp.mu.RUnlock()
	if exists {
		return nil
	}

	v := initViperWithOptions(cp.keyDelimiter)
	v.SetConfigFile(absPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file %s: %w", absPath, err)
	}

	cp.mu.Lock()
	cp.vipers[absPath] = v
	cp.mu.Unlock()

	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		cp.onConfigChange(generateCombinedViper(cp.vipers, cp.keyDelimiter))
	})

	return nil
}

func generateCombinedViper(vipers map[string]*viper.Viper, delimiter string) *viper.Viper {
	vip := initViperWithOptions(delimiter)
	for _, v := range vipers {
		for key, value := range v.AllSettings() {
			vip.Set(key, value)
		}
	}
	return vip
}

func initViperWithOptions(delimiter string) *viper.Viper {
	v := viper.NewWithOptions(
		viper.ExperimentalBindStruct(),
		viper.KeyDelimiter(delimiter),
	)
	v.SetEnvKeyReplacer(strings.NewReplacer(delimiter, "_"))
	v.AutomaticEnv()
	return v
}
