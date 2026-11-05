/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// +kubebuilder:object:generate:=true
package prometheus

import (
	"fmt"
	"strings"

	cpclusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions"
	remotewritedefinitions "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus/definitions/remote-write-relabel-configs"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	MimirPushUsernameKey = "mimirPushUsername"
	MimirPushPasswordKey = "mimirPushPassword"

	GrafanaLabsUsernameKey = "grafanaLabUsername"
	GrafanaLabsPasswordKey = "grafanaLabPassword"

	RemoteWriteUsernameKey = "remoteWriteUsername"
	RemoteWritePasswordKey = "remoteWritePassword"

	SecretName = "runai-prometheus"
)

// Prometheus defines configuration for the prometheus stack
type Prometheus struct {
	// Enabled specifies if prometheus should be installed
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// AgentPullPushRateMinForAlert specifies the value, under which the alert
	// for agent pull push rate will start firing (RunaiAgentPullRateLow & RunaiAgentClusterInfoPushRateLow)
	AgentPullPushRateMinForAlert *float32 `json:"agentPullPushRateMinForAlert,omitempty"`

	// ClusterSyncQueueDepthIncreaseMaxForAlert specifies the value, above which the alert
	// for cluster sync queue depth increase rate will start firing (RunaiClusterSyncHandlingRateLow)
	ClusterSyncQueueDepthIncreaseMaxForAlert *float32 `json:"clusterSyncQueueDepthIncreaseMaxForAlert,omitempty"`

	// RuleGroups are the PrometheusRuleGroups to create within the PrometheusRule for runai
	// +kubebuilder:validation:Optional
	RuleGroups RuleGroups `json:"ruleGroups,omitempty"`

	// Spec specifies the spec fields of Prometheus the operator should configure.
	// Some fields are not configurable, and will be set to default values.
	// NOTE: Make sure your installed Prometheus Operator supports the fields you are configuring.
	// +kubebuilder:validation:Optional
	Spec Spec `json:"spec,omitempty"`

	// Config specifies additional non-spec configuration for the prometheus stack
	// +kubebuilder:validation:Optional
	Config Config `json:"config,omitempty"`

	// AdditionalAlertLabels additional labels to add to runai alerts
	// +kubebuilder:validation:Optional
	AdditionalAlertLabels *map[string]string `json:"additionalAlertLabels,omitempty"`

	// Openshift specifies the openshift specific configuration
	// +kubebuilder:validation:Optional
	Openshift Openshift `json:"openshift,omitempty"`
}

type Openshift struct {
	// UseBuiltInPrometheus specifies whether to use the built-in OpenShift Prometheus or create a dedicated Prometheus instance
	// Defaults to true.
	// +kubebuilder:validation:Optional
	UseBuiltInPrometheus *bool `json:"useBuiltInPrometheus,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (p *Prometheus) SetDefaultsWhereNeeded(internalSpec *runv1_internal.Internal, agentSyncIntervalMillis *int, globalReplicaCount *int32) {
	if p.Enabled == nil {
		p.Enabled = ptr.To(!internalSpec.RemoteConfig.GetSystemConfig().Deactivated)
	}

	if p.AgentPullPushRateMinForAlert == nil {
		var agentSyncIntervalSeconds float32
		if agentSyncIntervalMillis == nil {
			agentSyncIntervalSeconds = float32(common.AgentSyncIntervalMillisDefault) / 1000
		} else {
			agentSyncIntervalSeconds = float32(*agentSyncIntervalMillis) / 1000
		}

		// the "rate" prometheus query calculates the rate per-second; hence the "1 / agentSyncIntervalSeconds"
		// and we are okay with half of the requests failing, or requests taking longer than expected, hence the " / 2".
		agentPullPushRateMinForAlert := 1 / agentSyncIntervalSeconds / 2
		p.AgentPullPushRateMinForAlert = ptr.To[float32](agentPullPushRateMinForAlert)
	}

	if p.ClusterSyncQueueDepthIncreaseMaxForAlert == nil {
		p.ClusterSyncQueueDepthIncreaseMaxForAlert = ptr.To[float32](50)
	}

	if p.AdditionalAlertLabels == nil {
		p.AdditionalAlertLabels = ptr.To(map[string]string{})
	}

	ruleGroupsConfig := &definitions.RuleGroupsConfig{
		AgentPullPushRateMinForAlert:             *p.AgentPullPushRateMinForAlert,
		ClusterSyncQueueDepthIncreaseMaxForAlert: *p.ClusterSyncQueueDepthIncreaseMaxForAlert,
		AdditionalAlertLabels:                    *p.AdditionalAlertLabels,
	}

	p.RuleGroups.SetDefaultsWhereNeeded(ruleGroupsConfig)
	p.Spec.SetDefaultsWhereNeeded(internalSpec, remotewritedefinitions.GetHandler(internalSpec), globalReplicaCount)
	p.Config.SetDefaultsWhereNeeded()
	p.Openshift.SetDefaultsWhereNeeded()
}

type RuleGroups []monitoringv1.RuleGroup

// SetDefaultsWhereNeeded sets default for unset fields
func (r *RuleGroups) SetDefaultsWhereNeeded(ruleGroupsConfig *definitions.RuleGroupsConfig) {
	if len(*r) != 0 {
		return
	}

	*r = definitions.DefaultRuleGroups(ruleGroupsConfig)
}

type Spec monitoringv1.PrometheusSpec

// SetDefaultsWhereNeeded sets default for unset fields
func (s *Spec) SetDefaultsWhereNeeded(internalSpec *runv1_internal.Internal, handler remotewritedefinitions.RemoteWriteRelabelConfigsHandler, globalReplicaCount *int32) {
	if s.ServiceMonitorSelector == nil {
		s.ServiceMonitorSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "app",
					Operator: metav1.LabelSelectorOpIn,
					Values: []string{
						"runai-scheduler",
						"nvidia-dcgm-exporter",
						"dcgm-exporter",
						"runai-project-controller",
						"kube-prometheus-stack-node-exporter",
						"kube-prometheus-stack-kube-state-metrics",
						"kube-prometheus-stack-kubelet",
						"runai-knative-serving",
						"runai-agent",
					},
				},
			},
		}
	}

	if s.PodMonitorNamespaceSelector == nil {
		s.PodMonitorNamespaceSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"name": "runai",
			},
		}
	}

	if s.PodMonitorSelector == nil {
		s.PodMonitorSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "release",
					Operator: metav1.LabelSelectorOpIn,
					Values: []string{
						"runai",
						"runai-cluster",
					},
				},
			},
		}
	}

	if s.ScrapeInterval == "" {
		s.ScrapeInterval = "30s"
	}

	if s.EvaluationInterval == "" {
		s.EvaluationInterval = "30s"
	}

	if s.Replicas == nil {
		s.Replicas = ptr.To(ptr.Deref(globalReplicaCount, 1))
	}

	if s.Retention == "" {
		s.Retention = "2h"
	}

	if (s.Image == nil || *s.Image == "") && internalSpec.Global.ImageRegistry != "" {
		reg := strings.TrimRight(internalSpec.Global.ImageRegistry, "/")
		s.Image = ptr.To(fmt.Sprintf("%s/prometheus:v2.54.1", reg))
	}

	s.configureRemoteMetricStoreWrites(internalSpec.Global.ControlPlane.Config.MetricsRemoteWrites, internalSpec.Global.CustomCA, handler)
}

func (s *Spec) configureRemoteMetricStoreWrites(remoteWrites []cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite, customCA *common.CustomCA, handler remotewritedefinitions.RemoteWriteRelabelConfigsHandler) {
	specsRemoteWrites := map[string]int{}
	for i, remoteWriteSpec := range s.RemoteWrite {
		if remoteWriteSpec.Name == nil || *remoteWriteSpec.Name == "" {
			continue
		}
		specsRemoteWrites[*remoteWriteSpec.Name] = i
	}

	for _, remoteWrite := range remoteWrites {
		var existingRemoteWriteSpec *monitoringv1.RemoteWriteSpec
		if remoteWriteSpec, ok := specsRemoteWrites[remoteWrite.GetLabel()]; ok {
			existingRemoteWriteSpec = &s.RemoteWrite[remoteWriteSpec]
		}
		clusterMetricsRemoteWrite := s.configureRemoteMetricStoreWrite(&remoteWrite, existingRemoteWriteSpec, customCA, handler)
		if clusterMetricsRemoteWrite != nil {
			if existingRemoteWriteSpec != nil {
				s.RemoteWrite[specsRemoteWrites[remoteWrite.GetLabel()]] = *clusterMetricsRemoteWrite
			} else {
				s.RemoteWrite = append(s.RemoteWrite, *clusterMetricsRemoteWrite)
			}
		}
	}
}

func (s *Spec) configureRemoteMetricStoreWrite(remoteWrite *cpclusterclient.ClusterControlPlaneConfigMetricsRemoteWrite, existingRemoteWriteSpec *monitoringv1.RemoteWriteSpec, customCA *common.CustomCA, handler remotewritedefinitions.RemoteWriteRelabelConfigsHandler) *monitoringv1.RemoteWriteSpec {
	remoteWriteRegexp := handler.GetRemoteWriteRelabelConfigs(remoteWrite.GetLabel())
	if remoteWriteRegexp == nil {
		return existingRemoteWriteSpec
	}

	if existingRemoteWriteSpec == nil {
		existingRemoteWriteSpec = &monitoringv1.RemoteWriteSpec{}
	}

	if existingRemoteWriteSpec.RemoteTimeout == nil || *existingRemoteWriteSpec.RemoteTimeout == "" {
		existingRemoteWriteSpec.RemoteTimeout = ptr.To(monitoringv1.Duration("30s"))
	}

	existingRemoteWriteSpec.WriteRelabelConfigs = remoteWriteRegexp
	existingRemoteWriteSpec.Name = ptr.To(remoteWrite.GetLabel())
	existingRemoteWriteSpec.URL = remoteWrite.GetUrl()
	userKey, passwordKey := GetSecretIdentifiers(remoteWrite.GetLabel())
	existingRemoteWriteSpec.BasicAuth = &monitoringv1.BasicAuth{
		Username: v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: SecretName,
			},
			Key: userKey,
		},
		Password: v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: SecretName,
			},
			Key: passwordKey,
		},
	}

	if headers, ok := remoteWrite.GetHeadersOk(); ok {
		if existingRemoteWriteSpec.Headers == nil {
			existingRemoteWriteSpec.Headers = map[string]string{}
		}
		for k, v := range *headers {
			existingRemoteWriteSpec.Headers[k] = v
		}
	}

	if customCA != nil && *customCA.Enabled {
		if existingRemoteWriteSpec.TLSConfig == nil {
			existingRemoteWriteSpec.TLSConfig = &monitoringv1.TLSConfig{}
		}

		existingRemoteWriteSpec.TLSConfig.CA.Secret = &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: customCA.Secret.Name,
			},
			Key:      customCA.Secret.Key,
			Optional: customCA.Secret.Optional,
		}
	}
	return existingRemoteWriteSpec
}

type Config struct {
	// AdvancedMetricsEnabled specifies if advanced metrics should be enabled
	// +kubebuilder:validation:Optional
	AdvancedMetricsEnabled *bool `json:"advancedMetricsEnabled,omitempty"`

	// EnableOTLPReceiver specifies if the Prometheus OTLP receiver endpoint should be opened.
	// Disabling this when knative is used in the cluster for inference services will disable inference services metrics.
	// Defaults to true.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	EnableOTLPReceiver *bool `json:"enableOTLPReceiver,omitempty"`
}

// SetDefaultsWhereNeeded sets default for unset fields
func (c *Config) SetDefaultsWhereNeeded() {
	if c.AdvancedMetricsEnabled == nil {
		c.AdvancedMetricsEnabled = ptr.To(false)
	}
	if c.EnableOTLPReceiver == nil {
		c.EnableOTLPReceiver = ptr.To(true)
	}
}

func GetSecretIdentifiers(remoteWriteLabel string) (string, string) {
	return fmt.Sprintf("%s-%s", RemoteWriteUsernameKey, remoteWriteLabel),
		fmt.Sprintf("%s-%s", RemoteWritePasswordKey, remoteWriteLabel)
}

func (o *Openshift) SetDefaultsWhereNeeded() {
	if o.UseBuiltInPrometheus == nil {
		o.UseBuiltInPrometheus = ptr.To(true)
	}
}
