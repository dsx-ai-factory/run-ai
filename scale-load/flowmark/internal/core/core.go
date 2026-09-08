/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package core

import "context"

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
