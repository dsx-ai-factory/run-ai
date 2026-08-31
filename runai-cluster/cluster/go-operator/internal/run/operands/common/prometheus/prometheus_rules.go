/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions"
)

func SetCommonPrometheusRuleFields(
	rule *monitoringv1.PrometheusRule,
	runaiConfig *runaiv1.RunaiConfig,
) *monitoringv1.PrometheusRule {
	rule.Labels = utils.InitMapIfNil(rule.Labels)
	rule.Labels["app"] = rule.Name

	rule.Spec.Groups = runaiConfig.Spec.Prometheus.RuleGroups

	advancedMetricsEnabled := runaiConfig.Spec.Prometheus.Config.AdvancedMetricsEnabled != nil &&
		*runaiConfig.Spec.Prometheus.Config.AdvancedMetricsEnabled
	if advancedMetricsEnabled {
		rule.Spec.Groups = append(rule.Spec.Groups, definitions.AdvancedRuleGroup())
	}

	return rule
}
