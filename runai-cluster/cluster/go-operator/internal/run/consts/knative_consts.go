/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package consts

// Knative feature names
const (
	// Features configmap features
	KnativeFeaturePodSpecSchedulerName         = "kubernetes.podspec-schedulername"
	KnativeFeaturePodSpecAffinity              = "kubernetes.podspec-affinity"
	KnativeFeaturePodSpecTolerations           = "kubernetes.podspec-tolerations"
	KnativeFeaturePodSpecVolumesEmptyDir       = "kubernetes.podspec-volumes-emptydir"
	KnativeFeaturePodSpecSecurityContext       = "kubernetes.podspec-securitycontext"
	KnativeFeaturePodSpecPersistentVolumeClaim = "kubernetes.podspec-persistent-volume-claim"
	KnativeFeaturePodSpecPersistentVolumeWrite = "kubernetes.podspec-persistent-volume-write"
	KnativeFeaturePodSpecNodeSelector          = "kubernetes.podspec-nodeselector"
	KnativeFeatureMultiContainer               = "multi-container"
	KnativeFeatureMultiContainerProbing        = "multi-container-probing"
	KnativeFeaturePodSpecInitContainers        = "kubernetes.podspec-init-containers"

	// Autoscaler configmap features
	KnativeFeatureEnableScaleToZero = "enable-scale-to-zero"

	// External configuration features
	KnativeExternalDomainSuffix = "example.com"

	// Feature values
	KnativeFeatureValueEnabled  = "enabled"
	KnativeFeatureValueDisabled = "disabled"
	KnativeFeatureValueTrue     = "true"
)

var (
	// Default values for features
	KnativeFeatureDefaultValueTrue     = KnativeFeatureValueTrue
	KnativeFeatureDefaultValueEnabled  = KnativeFeatureValueEnabled
	KnativeFeatureDefaultValueDisabled = KnativeFeatureValueDisabled
)

// KnativeFeatureConfig represents the configuration for a Knative feature
type KnativeFeatureConfig struct {
	Required      bool
	ExpectedValue string
	DefaultValue  *string
}

// KnativeFeaturesConfigMap contains the configuration for features in the features configmap
var KnativeFeaturesConfigMap = map[string]KnativeFeatureConfig{
	KnativeFeaturePodSpecSchedulerName: {
		Required:      true,
		ExpectedValue: KnativeFeatureValueEnabled,
		DefaultValue:  &KnativeFeatureDefaultValueDisabled,
	},
	KnativeFeaturePodSpecAffinity: {
		Required:      true,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecTolerations: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecVolumesEmptyDir: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecSecurityContext: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecPersistentVolumeClaim: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecPersistentVolumeWrite: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecNodeSelector: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeatureMultiContainer: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeatureMultiContainerProbing: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
	KnativeFeaturePodSpecInitContainers: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueEnabled,
	},
}

// KnativeAutoscalerConfigMap contains the configuration for features in the autoscaler configmap
var KnativeAutoscalerConfigMap = map[string]KnativeFeatureConfig{
	KnativeFeatureEnableScaleToZero: {
		Required:      false,
		ExpectedValue: KnativeFeatureValueTrue,
		DefaultValue:  &KnativeFeatureDefaultValueTrue,
	},
}

// KnativeExternalConfigMaps contains the list of external configuration configmaps to check
var KnativeExternalConfigMaps = []string{
	"config-domain",
}
