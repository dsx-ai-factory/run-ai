/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	"testing"
)

func TestPrometheusUpgradeRunaiConfigNoRemoteWrite(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	promUpgrader := newPrometheusUpgrader(icd)

	shouldUpgrade := promUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, false)
}

func TestPrometheusUpgradeRunaiConfigWithRunaiRemoteWrite(t *testing.T) {
	promSpec := monitoringv1.PrometheusSpec{
		CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
			RemoteWrite: []monitoringv1.RemoteWriteSpec{
				{
					URL: "https://aaa.example.com/api/v1/receive",
					BasicAuth: &monitoringv1.BasicAuth{
						Password: coreV1.SecretKeySelector{
							Key: "remoteWritePassword",
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "runai-grafana-secret",
							},
						},
						Username: coreV1.SecretKeySelector{
							Key: "remoteWriteUsername",
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "runai-grafana-secret",
							},
						},
					},
				},
			},
		},
	}
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Prometheus = &prometheus.Prometheus{
		Spec: prometheus.Spec(promSpec),
	}
	promUpgrader := newPrometheusUpgrader(icd)

	shouldUpgrade := promUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, true)
	resource, err := promUpgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	res := promUpgrader.shouldPerformUpgrade(resource)
	assert.Equal(t, res, false)

	assert.Equal(t, len(runaiConfig.Spec.Prometheus.Spec.RemoteWrite), 0)
}

func TestPrometheusUpgradeRunaiConfigWithMultipleRemoteWrites(t *testing.T) {
	promSpec := monitoringv1.PrometheusSpec{
		CommonPrometheusFields: monitoringv1.CommonPrometheusFields{
			RemoteWrite: []monitoringv1.RemoteWriteSpec{
				{
					URL: "https://aaa.example.com/api/v1/receive",
					BasicAuth: &monitoringv1.BasicAuth{
						Password: coreV1.SecretKeySelector{
							Key: "remoteWritePassword",
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "runai-grafana-secret",
							},
						},
						Username: coreV1.SecretKeySelector{
							Key: "remoteWriteUsername",
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "runai-grafana-secret",
							},
						},
					},
				},
				{
					URL: "https://my.remotewrite.com/api/v1/receive",
					BasicAuth: &monitoringv1.BasicAuth{
						Password: coreV1.SecretKeySelector{
							Key: "remoteWritePassword",
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "secret",
							},
						},
						Username: coreV1.SecretKeySelector{
							Key: "remoteWriteUsername",
							LocalObjectReference: coreV1.LocalObjectReference{
								Name: "secret",
							},
						},
					},
				},
			},
		},
	}
	runaiConfig := &runaiv1.RunaiConfig{}
	runaiConfig.Spec.Prometheus = &prometheus.Prometheus{
		Spec: prometheus.Spec(promSpec),
	}
	promUpgrader := newPrometheusUpgrader(icd)

	shouldUpgrade := promUpgrader.shouldPerformUpgrade(runaiConfig)
	assert.Equal(t, shouldUpgrade, true)
	resource, err := promUpgrader.updateRunaiConfig(runaiConfig)
	assert.Nil(t, err)
	res := promUpgrader.shouldPerformUpgrade(resource)
	assert.Equal(t, res, false)

	assert.Equal(t, len(runaiConfig.Spec.Prometheus.Spec.RemoteWrite), 1)
	assert.Equal(t, runaiConfig.Spec.Prometheus.Spec.RemoteWrite[0].URL, "https://my.remotewrite.com/api/v1/receive")
}
