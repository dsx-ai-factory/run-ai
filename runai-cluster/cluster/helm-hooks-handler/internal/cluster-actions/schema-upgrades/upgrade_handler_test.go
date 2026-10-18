/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package schema_upgrades

import (
	"context"
	"testing"

	cpclusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	controlplanedetailsdownloader "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/control-plane-details-downloader"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	rwlc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs"
	runaiv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/google/uuid"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
)

var icd = &controlplanedetailsdownloader.InstallationConfigurationData{
	Internal: runaiv1internal.Internal{
		Global: &runaiv1internal.Global{
			ControlPlane: &runaiv1internal.ControlPlane{
				Config: &runaiv1internal.ClusterControlPlaneConfig{
					cpclusterclient.ClusterControlPlaneConfig{
						MetricsRemoteWrites: []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite{
							{
								Url:   ptr.To("https://aaa.example.com/api/v1/receive"),
								Label: ptr.To(string(rwlc.ClusterMetricsRemoteWriteLabel)),
							},
							{
								Url:   ptr.To("https://app.example.com/api/v1/receive"),
								Label: ptr.To(rwlc.ApplicationMetricsRemoteWriteLabel),
							},
						},
					},
				},
			},
		},
	},
}

func TestPrometheusUpgradeClientRemoteWrite(t *testing.T) {
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
					URL: "https://my.remoteWrite.com/api/v1/receive",
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

	newUUID, _ := uuid.NewUUID()
	typeUId := types.UID(newUUID.String())
	clusterCr := &runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			UID:  typeUId,
			Name: runaiv1.ClusterName,
		},
	}
	runaiUpgradeHandler := NewRunaiConfigUpgrade(icd, clusterCr, "2.24.0")
	runaiConfig, err := runaiUpgradeHandler.UpgradeRunaiConfigSpec(context.TODO(), runaiConfig)
	assert.Nil(t, err)
	assert.Equal(t, len(runaiConfig.ObjectMeta.OwnerReferences), 1)
	assert.Equal(t, runaiConfig.ObjectMeta.OwnerReferences[0].UID, typeUId)

	assert.Equal(t, len(runaiConfig.Spec.Prometheus.Spec.RemoteWrite), 1)
	assert.Equal(t, runaiConfig.Spec.Prometheus.Spec.RemoteWrite[0].URL, "https://my.remoteWrite.com/api/v1/receive")
}

func TestFullUpgradeCycleWithOwnerReferencesAndDaemonSetTolerations(t *testing.T) {
	runaiConfig := &runaiv1.RunaiConfig{}
	tolerations := []coreV1.Toleration{
		{
			Key:      "key1",
			Operator: coreV1.TolerationOpEqual,
			Value:    "value1",
			Effect:   coreV1.TaintEffectNoSchedule,
		},
		{
			Key:               "key2",
			Operator:          coreV1.TolerationOpExists,
			Effect:            coreV1.TaintEffectNoExecute,
			TolerationSeconds: ptr.To(int64(3600)),
		},
	}
	runaiConfig.Spec.Global = &global.Global{
		AdditionalTolerations: tolerations,
	}
	newUUID, _ := uuid.NewUUID()
	typeUId := types.UID(newUUID.String())
	clusterCr := &runaiv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			UID:  typeUId,
			Name: runaiv1.ClusterName,
		},
	}

	runaiUpgradeHandler := NewRunaiConfigUpgrade(icd, clusterCr, "2.24.0")
	runaiConfig, err := runaiUpgradeHandler.UpgradeRunaiConfigSpec(context.TODO(), runaiConfig)
	assert.Nil(t, err)
	assert.Equal(t, len(runaiConfig.ObjectMeta.OwnerReferences), 1)
	assert.Equal(t, runaiConfig.ObjectMeta.OwnerReferences[0].UID, typeUId)

	assert.Equal(t, len(runaiConfig.Spec.Global.AdditionalTolerations), 0)
	assert.Equal(t, len(runaiConfig.Spec.Global.DaemonsetsTolerations), 2)
	assert.Equal(t, runaiConfig.Spec.Global.DaemonsetsTolerations, tolerations)
	assert.True(t, runaiConfig.Spec.Global.Affinity == nil)
}
