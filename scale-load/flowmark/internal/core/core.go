/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"time"
)

type Suite interface {
	Name() string
	Category() Category
	Run(ctx context.Context, target Target, scenario Scenario) (Section, error)
}

type Target struct {
	KubeconfigPath  string
	ControlPlaneURL string
	Username        string
	Password        string
	Token           string
}

type Scenario struct {
	Suite string
	Knobs map[string]any
}

func (s Scenario) Duration(key string, fallback time.Duration) (time.Duration, error) {
	raw, ok := s.Knobs[key]
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

func (s Scenario) Int(key string, fallback int) (int, error) {
	raw, ok := s.Knobs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	value, ok := raw.(int)
	if !ok {
		return 0, fmt.Errorf("knob %q must be a whole number, got %T", key, raw)
	}
	return value, nil
}

func (s Scenario) UnknownKnobs(known ...string) []string {
	isKnown := make(map[string]bool, len(known))
	for _, key := range known {
		isKnown[key] = true
	}
	var unknown []string
	for key := range s.Knobs {
		if !isKnown[key] {
			unknown = append(unknown, key)
		}
	}
	sort.Strings(unknown)
	return unknown
}

func (s Scenario) Ints(key string, fallback []int) ([]int, error) {
	raw, ok := s.Knobs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("knob %q must be a list of whole numbers, got %T", key, raw)
	}
	values := make([]int, 0, len(items))
	for i, item := range items {
		value, ok := item.(int)
		if !ok {
			return nil, fmt.Errorf("knob %q item %d must be a whole number, got %T", key, i+1, item)
		}
		values = append(values, value)
	}
	return values, nil
}

func (s Scenario) Map(key string) (map[string]any, error) {
	raw, ok := s.Knobs[key]
	if !ok || raw == nil {
		return nil, nil
	}
	value, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("knob %q must be a mapping, got %T", key, raw)
	}
	return value, nil
}

func (s Scenario) String(key string, fallback string) (string, error) {
	raw, ok := s.Knobs[key]
	if !ok || raw == nil {
		return fallback, nil
	}
	value, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("knob %q must be a string, got %T", key, raw)
	}
	return value, nil
}

func NewRunID() (string, error) {
	var raw [4]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate run id: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
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
