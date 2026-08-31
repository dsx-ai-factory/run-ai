/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package personalcreds materializes Run:AI personal-credential Secrets for
// YAML / NIM-API workloads. See designs/personal-credentials-plugin/.
package personalcreds

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PluginName   = "personalcreds"
	FieldManager = "personalcreds-plugin"
)

type Plugin struct {
	reader  client.Reader
	runaiNS string
}

// NewPlugin builds the personalcreds plugin with an explicit reader + runai
// namespace. Useful for tests that want to wire it directly.
func NewPlugin(reader client.Reader, runaiNS string) *Plugin {
	return &Plugin{reader: reader, runaiNS: runaiNS}
}

// New is the framework.PluginBuilder. Pass it to app.RunController in the
// pluginBuilders slice. It loads Config (env vars PERSONALCREDS_ENABLED,
// PERSONALCREDS_RUNAI_NAMESPACE) and returns framework.ErrPluginDisabled
// when the plugin is disabled - the framework skips it.
func New(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
	cfg, err := rcpconfig.GetRootR[*Config]()
	if err != nil {
		return nil, fmt.Errorf("load personalcreds config: %w", err)
	}
	if !cfg.Enabled {
		return nil, framework.ErrPluginDisabled
	}
	if deps.Client == nil {
		return nil, fmt.Errorf("personalcreds plugin requires a non-nil Client in PluginDeps")
	}
	return NewPlugin(deps.Client, cfg.RunaiNamespace), nil
}

func (p *Plugin) Name() string { return PluginName }

var (
	_ framework.ValidatorPlugin          = &Plugin{}
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.StatusPlugin             = &Plugin{}
	_ framework.WatcherPlugin            = &Plugin{}
)
