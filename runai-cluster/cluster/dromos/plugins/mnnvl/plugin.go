/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"fmt"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/spf13/viper"
	"sigs.k8s.io/controller-runtime/pkg/builder"
)

const (
	PluginName             = "mnnvl"
	workloadRunnerInitials = "wlr"
)

type Plugin struct {
	clusterElasticCDEnabled bool
	opts                    Options
	usesDefaultAffinity     bool // no PodAffinityProvider was given; using the built-in default
}

// NewPlugin builds the mnnvl plugin from the given Options. Useful for tests.
func NewPlugin(elasticCDEnabled bool, opts ...Option) *Plugin {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	usesDefaultAffinity := o.PodAffinityProvider == nil
	if usesDefaultAffinity {
		o.PodAffinityProvider = defaultPodAffinity
	}
	return &Plugin{
		clusterElasticCDEnabled: elasticCDEnabled,
		opts:                    o,
		usesDefaultAffinity:     usesDefaultAffinity,
	}
}

// NewBuilder returns the mnnvl plugin builder; when ENABLE_MNNVL_SUPPORT is off it
// returns a validator-only sentinel.
func NewBuilder(opts ...Option) framework.PluginBuilder {
	return func(_ context.Context, deps framework.PluginDeps) (framework.Plugin, error) {
		v := viper.New()
		v.AutomaticEnv()
		v.SetDefault("ENABLE_MNNVL_SUPPORT", false)
		v.SetDefault("ENABLE_ELASTIC_COMPUTE_DOMAIN", false)

		var cfg Config
		if err := v.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("parse mnnvl config: %w", err)
		}
		if !cfg.Enabled {
			return framework.DisabledValidator(PluginName, mnnvlInstructionOf, mnnvlInstructionPath()), nil
		}
		p := NewPlugin(cfg.ElasticComputeDomain, opts...)
		if p.usesDefaultAffinity && p.opts.DisableWorkloadLabelStamp {
			deps.Logger.Info("mnnvl: the default pod affinity is used but workload-label stamping is " +
				"disabled; nothing stamps the workload-name label, so GPU pods will not be grouped")
		}
		return p, nil
	}
}

func (p *Plugin) Name() string {
	return PluginName
}

func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	return bldr.Owns(&nvapi.ComputeDomain{}), nil
}

var (
	_ framework.ValidatorPlugin          = &Plugin{}
	_ framework.ResourceMutatePlugin     = &Plugin{}
	_ framework.AdditionalResourcePlugin = &Plugin{}
	_ framework.StatusPlugin             = &Plugin{}
	_ framework.WatcherPlugin            = &Plugin{}
)
