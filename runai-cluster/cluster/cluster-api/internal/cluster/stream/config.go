/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"sync"
	"time"
)

const (
	DefaultIdleTimeSec = 300
	DefaultPingTimeSec = 5
)

var lock = &sync.RWMutex{}

type Config struct {
	idleTime time.Duration
	pingTime time.Duration
}

var configInstance *Config

func GetInstance() *Config {
	if configInstance == nil {
		lock.Lock()
		defer lock.Unlock()

		if configInstance == nil {
			configInstance = &Config{
				idleTime: time.Second * DefaultIdleTimeSec,
				pingTime: time.Second * DefaultPingTimeSec,
			}
		}
	}

	return configInstance
}

func (c *Config) SetIdleTime(duration time.Duration) {
	lock.Lock()
	defer lock.Unlock()

	c.idleTime = duration
}

func (c *Config) GetIdleTime() time.Duration {
	lock.RLock()
	defer lock.RUnlock()

	return c.idleTime
}

func (c *Config) SetPingTime(duration time.Duration) {
	lock.Lock()
	defer lock.Unlock()

	c.pingTime = duration
}

func (c *Config) GetPingTime() time.Duration {
	lock.RLock()
	defer lock.RUnlock()

	return c.pingTime
}
