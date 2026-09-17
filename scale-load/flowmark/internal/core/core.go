/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package core

import (
	"context"
	"fmt"
	"time"
)

type Scenario interface {
	Name() string
	Run(ctx context.Context, env Env, cfg Config) (Section, error)
}

type Env struct {
	KubeconfigPath  string
	ControlPlaneURL string
	Username        string
	Password        string
	Token           string
}

type Config struct {
	Suite string
	Knobs map[string]any
}

func (c Config) Duration(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := c.Knobs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	text, ok := raw.(string)
	if !ok {
		return 0, fmt.Errorf("knob %q must be a duration string like \"2m\", got %T", key, raw)
	}
	parsed, err := time.ParseDuration(text)
	if err != nil {
		return 0, fmt.Errorf("knob %q: %w", key, err)
	}
	if parsed <= 0 {
		return 0, fmt.Errorf("knob %q must be a positive duration, got %s", key, parsed)
	}
	return parsed, nil
}

func (c Config) Int(key string, fallback int) (int, error) {
	raw, ok := c.Knobs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	value, ok := raw.(int)
	if !ok {
		return 0, fmt.Errorf("knob %q must be a whole number, got %T", key, raw)
	}
	return value, nil
}

func (c Config) String(key string, fallback string) (string, error) {
	raw, ok := c.Knobs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("knob %q must be a string, got %T", key, raw)
	}
	return value, nil
}

type Category string

const (
	CategoryScale Category = "scale"
	CategoryLoad  Category = "load"
)

type Section struct {
	Category Category       `json:"category"`
	Metrics  map[string]any `json:"metrics"`
	Failed   bool           `json:"failed"`
	Error    string         `json:"error,omitempty"`
}
