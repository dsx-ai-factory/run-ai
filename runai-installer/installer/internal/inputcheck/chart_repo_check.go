/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inputcheck

import (
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

func init() { checks = append(checks, chartRepoCheck) }

// chartRepoCheck ensures a connected install can locate the Run:ai charts.
// NGC registries have a built-in chart repository; every other registry type
// needs either registry.chartRepository in the config or the chart repository
// base from the environment. Air-gapped installs pull charts from the bundle
// and are exempt.
func chartRepoCheck(s *spec.Spec) []Finding {
	if s.IsAirGapped() || s.Registry == nil || s.Registry.Type == spec.RegistryTypeNGC {
		return nil
	}
	if s.Registry.ChartRepository != "" || step.ChartRepoBase() != "" {
		return nil
	}
	return []Finding{{
		Check:   "chart-repository",
		Field:   "registry.chartRepository",
		Message: "chart repository is required for a non-NGC registry: set registry.chartRepository or the " + step.ChartRepoBaseEnv + " environment variable",
	}}
}
