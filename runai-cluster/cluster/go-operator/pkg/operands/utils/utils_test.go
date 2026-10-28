/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"path"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
	//+kubebuilder:scaffold:imports
)

var _ = Describe("operand utils", func() {
	Context("GetAffinity", func() {
		var (
			runaiConfig *runaiv1.RunaiConfig
		)

		BeforeEach(func() {
			runaiConfig = &runaiv1.RunaiConfig{
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{},
				},
			}
		})

		It("should return affinity by runai config global", func() {
			runaiConfig.Spec.Global.Affinity = &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
						{
							Weight: 1,
							Preference: v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      "test",
										Operator: v1.NodeSelectorOpExists,
									},
								},
							},
						},
					},
				},
			}
			result := GetAffinity(runaiConfig)
			Expect(len(result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution)).To(Equal(1))
			Expect(len(result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions)).To(Equal(1))
			Expect(result.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Preference.MatchExpressions[0].Key).To(Equal("test"))
		})
	})

	Describe("GlobalVolumes", func() {
		var runaiConfig *runaiv1.RunaiConfig

		BeforeEach(func() {
			runaiConfig = &runaiv1.RunaiConfig{
				Spec: runaiv1.RunaiConfigSpec{
					Internal: &internal.Internal{
						Global: &internal.Global{
							CustomCA: &common.CustomCA{
								Enabled: ptr.To[bool](true),
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "my-secret",
									},
									Key: "ca.crt",
								},
							},
						},
					},
				},
			}
		})

		It("should return an empty list when custom CA is not enabled", func() {
			runaiConfig.Spec.Internal.Global.CustomCA.Enabled = ptr.To[bool](false)
			result := GlobalVolumes(runaiConfig)

			Expect(result).To(HaveLen(0))
		})

		It("should return a single volume with Secret source", func() {
			result := GlobalVolumes(runaiConfig)

			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal(caVolumeName))
			Expect(result[0].VolumeSource.Secret.SecretName).To(Equal("my-secret"))
			Expect(result[0].VolumeSource.Secret.Items).To(HaveLen(1))
			Expect(result[0].VolumeSource.Secret.Items[0].Key).To(Equal("ca.crt"))
			Expect(result[0].VolumeSource.Secret.Items[0].Path).To(Equal(caMountFile))
			Expect(result[0].VolumeSource.Secret.Optional).To(BeNil())
		})
	})

	Describe("GlobalVolumeMounts", func() {
		var runaiConfig *runaiv1.RunaiConfig

		BeforeEach(func() {
			runaiConfig = &runaiv1.RunaiConfig{
				Spec: runaiv1.RunaiConfigSpec{
					Global: &global.Global{
						CustomCAGit: &common.CustomCA{
							Enabled: ptr.To[bool](true),
							Secret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "my-secret-for-git",
								},
								Key: "ca-for-git.crt",
							},
						},
						CustomCAS3: &common.CustomCA{
							Enabled: ptr.To[bool](true),
							Secret: &v1.SecretKeySelector{
								LocalObjectReference: v1.LocalObjectReference{
									Name: "my-secret-for-s3",
								},
								Key: "ca-for-s3.crt",
							},
						},
					},
					Internal: &internal.Internal{
						Global: &internal.Global{
							CustomCA: &common.CustomCA{
								Enabled: ptr.To[bool](true),
								Secret: &v1.SecretKeySelector{
									LocalObjectReference: v1.LocalObjectReference{
										Name: "my-secret",
									},
									Key: "ca.crt",
								},
							},
						},
					},
				},
			}
		})

		It("should return an empty list when custom CA is not enabled", func() {
			runaiConfig.Spec.Internal.Global.CustomCA.Enabled = ptr.To[bool](false)
			result := GlobalVolumeMounts(runaiConfig)

			Expect(result).To(HaveLen(0))
		})

		It("should return a single volume mount with correct values", func() {
			result := GlobalVolumeMounts(runaiConfig)

			Expect(result).To(HaveLen(1))
			Expect(result[0].Name).To(Equal(caVolumeName))
			Expect(result[0].MountPath).To(Equal(path.Join(caMountDir, caMountFile)))
			Expect(result[0].SubPath).To(Equal(caMountFile))
		})
	})

	Describe("ImagePullSecrets", func() {
		var runaiConfig *runaiv1.RunaiConfig

		BeforeEach(func() {
			runaiConfig = &runaiv1.RunaiConfig{
				Spec: runaiv1.RunaiConfigSpec{
					Internal: &internal.Internal{
						Global: &internal.Global{},
					},
				},
			}
		})

		It("should return an empty list when none are configured", func() {
			result := GetGlobalImagePullSecrets(runaiConfig)
			Expect(result).To(BeEmpty())
		})

		It("should return exactly the configured image pull secrets", func() {
			// The chart delivers the control-plane registry secret (runai-reg-creds)
			// through this list by default, joined with any additional secrets.
			runaiConfig.Spec.Internal.Global.AdditionalImagePullSecrets = []string{"runai-reg-creds", "secret1", "secret2"}
			result := GetGlobalImagePullSecrets(runaiConfig)
			Expect(result).To(HaveLen(3))
			Expect(result[0].Name).To(Equal("runai-reg-creds"))
			Expect(result[1].Name).To(Equal("secret1"))
			Expect(result[2].Name).To(Equal("secret2"))
		})
	})

	Describe("mergeNodeSelectorTerms", func() {
		cases := map[string]struct {
			inputTerms []v1.NodeSelectorTerm
			inputTerm  v1.NodeSelectorTerm
			expected   []v1.NodeSelectorTerm
		}{
			"empty terms": {
				inputTerms: []v1.NodeSelectorTerm{},
				inputTerm: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "test",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"test"},
						},
					},
				},
				expected: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
						},
					},
				},
			},
			"an empty term": {
				inputTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
						},
					},
				},
				inputTerm: v1.NodeSelectorTerm{},
				expected: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
						},
					},
				},
			},
			"a single term in the terms list": {
				inputTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
						},
					},
				},
				inputTerm: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "test2",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"test2"},
						},
					},
				},
				expected: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
							{
								Key:      "test2",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test2"},
							},
						},
					},
				},
			},
			"multiple terms in the terms list": {
				inputTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
						},
					},
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test2",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test2"},
							},
						},
					},
				},
				inputTerm: v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      "test3",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"test3"},
						},
					},
				},
				expected: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test"},
							},
							{
								Key:      "test3",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test3"},
							},
						},
					},
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "test2",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test2"},
							},
							{
								Key:      "test3",
								Operator: v1.NodeSelectorOpIn,
								Values:   []string{"test3"},
							},
						},
					},
				},
			},
		}

		for caseName, caseDetails := range cases {
			caseDetails := caseDetails
			It(fmt.Sprintf("should merge node selector terms correctly in case of %s", caseName), func() {
				result := mergeNodeSelectorTerms(caseDetails.inputTerms, caseDetails.inputTerm)
				Expect(result).To(Equal(caseDetails.expected))
			})
		}
	})
})

var _ = Describe("MergeAffinities", func() {
	It("should return globalAffinity when localAffinity is nil", func() {
		globalAffinity := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{},
		}
		result := MergeAffinities(nil, globalAffinity, nil, false)
		Expect(result).To(Equal(globalAffinity))
	})

	It("should return localAffinity when globalAffinity is nil", func() {
		localAffinity := &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{},
		}
		result := MergeAffinities(localAffinity, nil, nil, false)
		Expect(result).To(Equal(localAffinity))
	})

	It("should prefer local NodeAffinity over global", func() {
		local := &v1.NodeAffinity{}
		global := &v1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{}}
		localAffinity := &v1.Affinity{NodeAffinity: local}
		globalAffinity := &v1.Affinity{NodeAffinity: global}
		result := MergeAffinities(localAffinity, globalAffinity, nil, false)
		Expect(result.NodeAffinity).To(Equal(local))
	})

	It("should prefer global NodeAffinity when local is nil", func() {
		global := &v1.NodeAffinity{}
		localAffinity := &v1.Affinity{}
		globalAffinity := &v1.Affinity{NodeAffinity: global}
		result := MergeAffinities(localAffinity, globalAffinity, nil, false)
		Expect(result.NodeAffinity).To(Equal(global))
	})

	It("should prefer local PodAffinity over global", func() {
		local := &v1.PodAffinity{}
		global := &v1.PodAffinity{}
		localAffinity := &v1.Affinity{PodAffinity: local}
		globalAffinity := &v1.Affinity{PodAffinity: global}
		result := MergeAffinities(localAffinity, globalAffinity, nil, false)
		Expect(result.PodAffinity).To(Equal(local))
	})

	It("should prefer global PodAffinity when local is nil", func() {
		global := &v1.PodAffinity{}
		localAffinity := &v1.Affinity{}
		globalAffinity := &v1.Affinity{PodAffinity: global}
		result := MergeAffinities(localAffinity, globalAffinity, nil, false)
		Expect(result.PodAffinity).To(Equal(global))
	})

	It("should use local PodAntiAffinity if present", func() {
		local := &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
				{TopologyKey: "foo"},
			},
		}
		localAffinity := &v1.Affinity{PodAntiAffinity: local}
		globalAffinity := &v1.Affinity{}
		result := MergeAffinities(localAffinity, globalAffinity, nil, false)
		Expect(result.PodAntiAffinity).To(Equal(local))
	})

	It("should use global PodAntiAffinity if local is nil", func() {
		global := &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
				{TopologyKey: "bar"},
			},
		}
		localAffinity := &v1.Affinity{}
		globalAffinity := &v1.Affinity{PodAntiAffinity: global}
		result := MergeAffinities(localAffinity, globalAffinity, nil, false)
		Expect(result.PodAntiAffinity).To(Equal(global))
	})

	It("should add default preferred podAntiAffinity when both are nil and label exists", func() {
		labelMap := map[string]string{"app": "my-app"}
		localAffinity := &v1.Affinity{}
		globalAffinity := &v1.Affinity{}
		result := MergeAffinities(localAffinity, globalAffinity, labelMap, true)
		Expect(result.PodAntiAffinity).ToNot(BeNil())
		Expect(result.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution).To(HaveLen(1))
		Expect(result.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].Weight).To(Equal(int32(100)))
		Expect(result.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels).To(Equal(labelMap))
		Expect(result.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.TopologyKey).To(Equal("kubernetes.io/hostname"))
	})

	It("should reject invalid workload-injected container resources", func() {
		globalConfig := &global.Global{
			WorkloadInjectedContainers: &global.WorkloadInjectedContainersConfig{
				OAuth2Proxy: &global.OAuth2ProxyContainerConfig{
					Resources: &global.WorkloadInjectedContainerResourceRequirements{
						Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("-1m")},
					},
				},
			},
		}
		var envVars []v1.EnvVar

		err := AddWorkloadInjectedContainerResourcesEnvVarsIfConfigured(globalConfig, &envVars)

		Expect(err).To(MatchError(ContainSubstring("must not be negative")))
		Expect(envVars).To(BeEmpty())
	})
})
