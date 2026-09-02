/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compliance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/policy-service/pkg/database/dbmodel"
	policyCompliance "github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"gopkg.in/yaml.v3"
	"os"
)

const (
	exposerPath = "../../../workloads-and-policies/policy_compliance"
)

var _ = Describe("test compliance", func() {
	var samplePolicy api.WorkspacePolicyDefaultsAndRulesV2

	BeforeEach(func() {
		samplePolicy = api.WorkspacePolicyDefaultsAndRulesV2{
			Defaults: &api.WorkspacePolicyDefaultsV2{
				CreateHomeDir: convutils.Ref(true),
				Args:          convutils.Ref(""),
				Image:         convutils.Ref("myimage"),
				Security: &api.SupersetDefaultsAllOfSecurity{
					RunAsUid: convutils.Ref(int64(1)),
					RunAsGid: convutils.Ref(int64(6)),
				},
			},
			Rules: &api.WorkspacePolicyRulesV2{
				CreateHomeDir: &api.BooleanRules{
					CanEdit: convutils.Ref(true),
				},
				Args: &api.StringRules{
					CanEdit: convutils.Ref(false),
				},
				Image: &api.StringRules{
					Required: convutils.Ref(true),
					Options: &[]api.StringOption{
						{Value: "ubuntu", Displayed: convutils.Ref("ubuntu latest")},
						{Value: "python"},
					},
				},
				ImagePullPolicy: &api.ImagePullPolicyRules{
					CanEdit: convutils.Ref(false),
				},
				EnvironmentVariables: &api.InstancesRules{
					Instances: &api.ItemRules{
						CanAdd: convutils.Ref(false),
						Locked: []string{
							"HOME",
							"USER",
						},
					},
				},
				Security: &api.SupersetRulesAllOfSecurity{
					AllowPrivilegeEscalation: &api.BooleanRules{
						Required: convutils.Ref(true),
					},
					HostIpc: &api.BooleanRules{
						Required: convutils.Ref(true),
					},
					RunAsUid: &api.IntegerRulesOptional{
						Min: convutils.Ref(int32(2)),
					},
					RunAsGid: &api.IntegerRulesOptional{
						Min:  convutils.Ref(int32(3)),
						Step: convutils.Ref(int32(2)),
					},
				},
			},
		}
	})

	Context("no rules", func() {
		It("no errors", func() {
			effectivePolicy := dbmodel.EffectivePolicy{}

			values := openapi.WorkspaceSpecSpec{
				EnvironmentVariables: []openapi.EnvironmentVariable{},
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&effectivePolicy, valuesAsMap)

			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{}))
		})
	})

	Context("composer tests from shared-packages", func() {
		DescribeTable("should create proper workloads or return sensible errors",
			func(testDetails policyCompliance.ComplianceTest) {
				policyCompliance.RunComplianceTest(testDetails, exposerPath)
			}, policyCompliance.ComplianceTests)
	})

	Context("blocked consistency", func() {
		It("blocked <-> canEdit=false", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					SupplementalGroups: &api.StringRules{
						Blocked:  convutils.Ref(true),
						CanEdit:  convutils.Ref(false),
						Required: convutils.Ref(true),
						Options: &[]api.StringOption{
							{Value: "ubuntu", Displayed: convutils.Ref("ubuntu latest")},
						},
					},
				},
			}

			samplePolicy.Defaults = nil

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			err := ValidateSubmittedPolicy(effectivePolicy.EffectivePolicy.RulesV2, effectivePolicy.EffectivePolicy.DefaultsV2)
			Expect(err).NotTo(BeNil())

			errors := policyCompliance.ComplianceErrors{}
			validateRulesConsistency(effectivePolicy.EffectivePolicy.RulesV2, &errors)

			Expect(len(errors)).To(Equal(3))
			Expect(errors[0].Details).To(Equal("specify either blocked=true or canEdit=false, not both"))
			Expect(errors[1].Details).To(Equal("specify either blocked=true or required=true, not both"))
			Expect(errors[2].Details).To(Equal("specify either blocked=true or options, not both"))
		})
	})

	Context("steps too big", func() {
		It("error", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						Min:  convutils.Ref(int32(5)),
						Max:  convutils.Ref(int32(10)),
						Step: convutils.Ref(int32(6)),
					},
				},
			}
			samplePolicy.Defaults = nil

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			err := ValidateSubmittedPolicy(effectivePolicy.EffectivePolicy.RulesV2, effectivePolicy.EffectivePolicy.DefaultsV2)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"security.runAsUid -> [step] step value must not exceed the calculated range between min and max values",
			}))
		})
	})

	Context("validate the defaults themselves", func() {
		It("send the defaults as the spec (defaults should be nil in this case)", func() {
			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"environmentVariables.instances[HOME] -> environmentVariables.instances: [UnknownRule] default value is missing for locked item 'HOME'",
				"environmentVariables.instances[USER] -> environmentVariables.instances: [UnknownRule] default value is missing for locked item 'USER'",
				"image -> [options] the value of this field must be one of: ubuntu, python",
				"imagePullPolicy -> [canEdit] default value must be provided for policy fields having canEdit=false",
				"security.runAsGid -> [step] value must follow 2 steps from 3 (3.00, 5.00, etc.)",
				"security.runAsUid -> [min] must be no less than 2",
			}))
		})
	})

	Context("itemized with missing keys", func() {
		It("return with error", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				EnvironmentVariables: &api.EnvironmentVariablesDefaults{
					Instances: &[]api.EnvironmentVariable{
						{
							Value: convutils.Ref("aaa"),
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"environmentVariables.instances -> [UnknownRule] attribute 'name' must be provided to all instances",
			}))
		})
	})

	Context("itemized with duplicate keys", func() {
		It("return with error", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				EnvironmentVariables: &api.EnvironmentVariablesDefaults{
					Instances: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("HOME"),
							Value: convutils.Ref("aaa"),
						},
						{
							Name:  convutils.Ref("HOME"),
							Value: convutils.Ref("bbb"),
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"environmentVariables.instances[HOME] -> environmentVariables.instances: [UnknownRule] duplicate name 'HOME'",
			}))
		})
	})

	Context("exclude flag", func() {
		It("is not allowed in policy defaults", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				EnvironmentVariables: &api.EnvironmentVariablesDefaults{
					Instances: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("HOME"),
							Value: convutils.Ref("aaa"),
						},
						{
							Name:    convutils.Ref("USER"),
							Exclude: convutils.Ref(true),
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"environmentVariables.instances -> [UnknownRule] 'exclude' is not allowed in policy defaults",
			}))
		})
	})

	Context("min max", func() {
		It("min should be less or equal to max", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						Min: convutils.Ref(int32(5)),
						Max: convutils.Ref(int32(3)),
					},
				},
				Compute: &api.SupersetRulesAllOfCompute{
					GpuPortionRequest: &api.NumberRules{
						Max: convutils.Ref(float32(3.5)),
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Compute: &api.SupersetDefaultsAllOfCompute{
					GpuPortionRequest: convutils.Ref(3.6),
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"compute.gpuPortionRequest -> [max] must be no greater than 3.5",
				"security.runAsUid -> [max] max value must be no less than min value",
			}))
		})
	})

	Context("distributed", func() {
		It("rules for both master and worker", func() {
			sampleDistributed := api.DistributedPolicyDefaultsAndRulesV2{
				Defaults: &api.DistributedPolicyDefaultsV2{
					Worker: &api.DistributedPolicyDefaultsV2Worker{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(50)),
							RunAsGid: convutils.Ref(int64(150)),
						},
					},
					Master: &api.ReplicaDefaultsV2{
						Security: &api.SupersetDefaultsAllOfSecurity{
							RunAsUid: convutils.Ref(int64(250)),
							RunAsGid: convutils.Ref(int64(350)),
						},
					},
				},
			}

			sampleDistributed.Rules = &api.DistributedPolicyRulesV2{
				Worker: &api.DistributedPolicyRulesV2Worker{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(200)),
							Max: convutils.Ref(int32(100)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(300)),
							Max: convutils.Ref(int32(400)),
						},
					},
				},
				Master: &api.ReplicaRulesV2{
					Security: &api.SupersetRulesAllOfSecurity{
						RunAsUid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(500)),
							Max: convutils.Ref(int32(600)),
						},
						RunAsGid: &api.IntegerRulesOptional{
							Min: convutils.Ref(int32(800)),
							Max: convutils.Ref(int32(700)),
						},
					},
				},
			}
			effectivePolicy := loadDistributedEffectivePolicy(sampleDistributed)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"master.security.runAsGid -> [max] max value must be no less than min value",
				"master.security.runAsGid -> [min] must be no less than 800",
				"master.security.runAsUid -> [min] must be no less than 500",
				"worker.security.runAsGid -> [min] must be no less than 300",
				"worker.security.runAsUid -> [max] max value must be no less than min value",
				"worker.security.runAsUid -> [min] must be no less than 200",
			}))
		})
	})

	Context("member rules", func() {
		It("needs to be enforces", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Compute: &api.SupersetRulesAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesRules{
						Attributes: &api.ExtendedResourceRules{
							Quantity: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "10M"},
									{Value: "20M"},
								},
							},
						},
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Compute: &api.SupersetDefaultsAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesDefaults{
						Attributes: &api.ExtendedResource{
							Quantity: convutils.Ref("40M"), // should fail, incorrect value
						},
						Instances: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("nvidia/a100"),
								Quantity: convutils.Ref("10M"),
							},
							{
								Resource: convutils.Ref("nvidia/v100"),
								Quantity: convutils.Ref("30M"),
							},
							{
								Resource: convutils.Ref("nvidia/x100"), // this one takes from the defaults
							},
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"compute.extendedResources.attributes.quantity -> [options] the value of this field must be one of: 10M, 20M",
				"compute.extendedResources.instances[nvidia/v100].quantity -> compute.extendedResources.attributes.quantity: [options] the value of this field must be one of: 10M, 20M",
			}))
		})
	})

	Context("member rules - single item", func() {
		It("needs to be enforced", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Compute: &api.SupersetRulesAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesRules{
						Attributes: &api.ExtendedResourceRules{
							Quantity: &api.StringRules{
								Options: &[]api.StringOption{
									{Value: "10M"},
									{Value: "20M"},
								},
							},
						},
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Compute: &api.SupersetDefaultsAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesDefaults{
						Attributes: &api.ExtendedResource{
							Quantity: convutils.Ref("40M"), // should fail, incorrect value
						},
						Instances: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("nvidia/v100"),
								Quantity: convutils.Ref("30M"),
							},
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"compute.extendedResources.attributes.quantity -> [options] the value of this field must be one of: 10M, 20M",
				"compute.extendedResources.instances.quantity -> compute.extendedResources.attributes.quantity: [options] the value of this field must be one of: 10M, 20M",
			}))
		})
	})

	Context("member rules multiple key", func() {
		It("needs to be rejected", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Compute: &api.SupersetDefaultsAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesDefaults{
						Instances: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("nvidia/a100"),
								Quantity: convutils.Ref("10M"),
							},
							{
								Resource: convutils.Ref("nvidia/a100"),
								Quantity: convutils.Ref("30M"),
							},
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"compute.extendedResources.instances[nvidia/a100] -> compute.extendedResources.instances: [UnknownRule] duplicate resource 'nvidia/a100'",
			}))
		})
	})

	Context("empty key", func() {
		It("needs to be rejected", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Compute: &api.SupersetDefaultsAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesDefaults{
						Instances: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref(""),
								Quantity: convutils.Ref("10M"),
							},
							{
								Resource: convutils.Ref("nvidia/v100"),
								Quantity: convutils.Ref("30M"),
							},
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"compute.extendedResources.instances -> [UnknownRule] attribute 'resource' must be provided to all instances",
			}))
		})
	})

	Context("mandatory for storage which is not given", func() {
		It("should be ok", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				EnvironmentVariables: &api.EnvironmentVariablesDefaults{
					Instances: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("HOME"),
							Value: convutils.Ref("/my/home"),
						},
					},
				},
			}

			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						Min: convutils.Ref(int32(1)),
					},
				},
				EnvironmentVariables: &api.InstancesRules{
					Instances: &api.ItemRules{
						Locked: []string{"HOME"},
					},
				},
				Storage: &api.SupersetRulesAllOfStorage{
					S3: &api.S3sRules{
						Attributes: &api.S3Rules{
							Bucket: &api.StringRules{
								Required: convutils.Ref(true),
							},
						},
					},
					Git: &api.GitsRules{
						Attributes: &api.GitRules{
							Repository: &api.StringRules{
								Required: convutils.Ref(true),
							},
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)

			workload := openapi.SupersetSpec{
				EnvironmentVariables: []openapi.EnvironmentVariable{
					{
						Name:  *openapi.NewNullableString(convutils.Ref("HOME")),
						Value: *openapi.NewNullableString(convutils.Ref("/his/home")),
					},
				},
				Security: *openapi.NewNullableWorkspaceSpecSpecAllOfSecurity(&openapi.WorkspaceSpecSpecAllOfSecurity{
					RunAsUid: *openapi.NewNullableInt64(convutils.Ref(int64(0))),
				}),
				Storage: *openapi.NewNullableWorkspaceSpecSpecAllOfStorage(&openapi.WorkspaceSpecSpecAllOfStorage{
					S3: []openapi.S3Instance{
						{
							Bucket: *openapi.NewNullableString(convutils.Ref("my-bucket")),
						},
						{
							Path: *openapi.NewNullableString(convutils.Ref("/a/b")),
						},
					},
				}),
			}

			valuesAsMap, _ := gojsonq.LoadJson(workload)

			err := ValidateWorkload(&effectivePolicy.EffectivePolicy, valuesAsMap)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"environmentVariables[HOME] -> environmentVariables.instances: [locked] the administrator prohibited modifying item 'HOME'",
				"security.runAsUid -> [min] must be no less than 1",
				"storage.s3[2].bucket -> storage.s3.attributes.bucket: [required] value must be provided for required fields",
			}))
		})
	})

	Context("exclude in policy", func() {
		It("not allowed", func() {
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Compute: &api.SupersetDefaultsAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesDefaults{
						Instances: &[]api.ExtendedResource{
							{
								Resource: convutils.Ref("nvidia/a100"),
								Exclude:  convutils.Ref(true),
							},
						},
					},
				},
			}

			effectivePolicy := loadEffectivePolicy(samplePolicy)
			_, err := ValidateDefaults(effectivePolicy)
			Expect(err).NotTo(BeNil())

			Expect(err.AsSortedArray()).To(Equal([]string{
				"compute.extendedResources.instances -> [UnknownRule] 'exclude' is not allowed in policy defaults",
			}))
		})
	})

	Context("exclude a locked variable", func() {
		It("should be rejected", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				EnvironmentVariables: &api.InstancesRules{
					Instances: &api.ItemRules{
						CanAdd: convutils.Ref(false),
						Locked: []string{
							"HOME",
							"USER",
							"ANOTHER",
						},
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				EnvironmentVariables: &api.EnvironmentVariablesDefaults{
					Instances: &[]api.EnvironmentVariable{
						{
							Name:  convutils.Ref("USER"),
							Value: convutils.Ref("myuser"),
						},
						{
							Name:  convutils.Ref("HOME"),
							Value: convutils.Ref("/myhome"),
						},
						{
							Name:  convutils.Ref("NOT_LOCKED"),
							Value: convutils.Ref("i am free"),
						},
					},
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.WorkspaceSpecSpec{
				EnvironmentVariables: []openapi.EnvironmentVariable{
					{
						Name:  *openapi.NewNullableString(convutils.Ref("USER")),
						Value: *openapi.NewNullableString(convutils.Ref("hisuser")),
					},
					{
						Name:    *openapi.NewNullableString(convutils.Ref("HOME")),
						Exclude: *openapi.NewNullableBool(convutils.Ref(true)),
					},
					{
						Name:    *openapi.NewNullableString(convutils.Ref("NOT_LOCKED")),
						Exclude: *openapi.NewNullableBool(convutils.Ref(true)),
					},
					{
						Name:  *openapi.NewNullableString(convutils.Ref("ADDED")),
						Value: *openapi.NewNullableString(convutils.Ref("added")),
					},
				},
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"environmentVariables -> environmentVariables.instances: [canAdd] the administrator prohibited adding new items",
				"environmentVariables[HOME] -> environmentVariables.instances: [locked] the administrator prohibited modifying item 'HOME'",
				"environmentVariables[USER] -> environmentVariables.instances: [locked] the administrator prohibited modifying item 'USER'",
			}))
		})
	})

	Context("no rules for you", func() {
		It("considered as compliance", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Image: &api.StringRules{
					Required: convutils.Ref(true),
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.WorkspaceSpecSpecAllOfCompute{
				CpuMemoryLimit: *openapi.NewNullableString(convutils.Ref("10MB")),
			}
			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(len(complianceErrors)).To(Equal(0))
		})
	})

	Context("required variable with no value", func() {
		It("should be rejected", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Image: &api.StringRules{
					Required: convutils.Ref(true),
				},
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						Min: convutils.Ref(int32(1)),
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Args: convutils.Ref("default"),
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.WorkspaceSpecSpec{
				EnvironmentVariables: []openapi.EnvironmentVariable{},
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"image -> [required] value must be provided for required fields",
			}))
		})
	})

	Context("canEdit = false", func() {
		It("cannot edit the value", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					RunAsUid: &api.IntegerRulesOptional{
						CanEdit: convutils.Ref(false),
					},
					RunAsGid: &api.IntegerRulesOptional{
						CanEdit: convutils.Ref(true),
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Security: &api.SupersetDefaultsAllOfSecurity{
					RunAsUid: convutils.Ref(int64(10)),
					RunAsGid: convutils.Ref(int64(20)),
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.WorkspaceSpecSpec{
				Security: *openapi.NewNullableWorkspaceSpecSpecAllOfSecurity(&openapi.WorkspaceSpecSpecAllOfSecurity{
					RunAsUid: *openapi.NewNullableInt64(convutils.Ref(int64(20))),
				}),
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"security.runAsUid -> [canEdit] the administrator prohibited changing the value of this field",
			}))
		})
	})

	Context("members rules cannot run when key is not found", func() {
		It("should be rejected", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Compute: &api.SupersetRulesAllOfCompute{
					ExtendedResources: &api.ExtendedResourcesRules{
						Attributes: &api.ExtendedResourceRules{
							Quantity: &api.StringRules{
								Options: &[]api.StringOption{
									{
										Value: "10M",
									},
									{
										Value: "20M",
									},
								},
							},
						},
					},
				},
			}
			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.SupersetSpec{
				Compute: *openapi.NewNullableWorkspaceSpecSpecAllOfCompute(&openapi.WorkspaceSpecSpecAllOfCompute{
					ExtendedResources: []openapi.ExtendedResource{
						{
							Resource: *openapi.NewNullableString(convutils.Ref("Abc/Def")),
							Quantity: *openapi.NewNullableString(convutils.Ref("20MB")),
						},
						{
							Quantity: *openapi.NewNullableString(convutils.Ref("10MB")),
						},
					},
				}),
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"compute.extendedResources -> compute.extendedResources.instances: [UnknownRule] attribute 'resource' must be provided to all instances",
			}))
		})
	})

	Context("capabilities canEdit = false", func() {
		It("cannot edit the array", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Capabilities: &api.ArrayRules{
						CanEdit: convutils.Ref(false),
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Security: &api.SupersetDefaultsAllOfSecurity{
					Capabilities: &[]api.Capability{
						api.CAPABILITY_CHOWN,
						api.CAPABILITY_FOWNER,
					},
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.SupersetSpec{
				Security: *openapi.NewNullableWorkspaceSpecSpecAllOfSecurity(&openapi.WorkspaceSpecSpecAllOfSecurity{
					Capabilities: []openapi.Capability{
						openapi.CAPABILITY_CHOWN,
						openapi.CAPABILITY_KILL,
					},
				}),
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"security.capabilities -> [canEdit] the administrator prohibited changing the value of this field",
			}))
		})
	})

	Context("capabilities canEdit = false", func() {
		It("same array is ok", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Capabilities: &api.ArrayRules{
						CanEdit: convutils.Ref(false),
					},
				},
			}
			samplePolicy.Defaults = &api.WorkspacePolicyDefaultsV2{
				Security: &api.SupersetDefaultsAllOfSecurity{
					Capabilities: &[]api.Capability{
						api.CAPABILITY_CHOWN,
						api.CAPABILITY_FOWNER,
					},
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.SupersetSpec{
				Security: *openapi.NewNullableWorkspaceSpecSpecAllOfSecurity(&openapi.WorkspaceSpecSpecAllOfSecurity{
					Capabilities: []openapi.Capability{
						openapi.CAPABILITY_CHOWN,
						openapi.CAPABILITY_FOWNER,
					},
				}),
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(len(complianceErrors)).To(Equal(0))
		})
	})

	Context("capabilities - required", func() {
		It("must validate array is provided", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Security: &api.SupersetRulesAllOfSecurity{
					Capabilities: &api.ArrayRules{
						Required: convutils.Ref(true),
					},
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.SupersetSpec{}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"security.capabilities -> [required] value must be provided for required fields",
			}))
		})
	})

	Context("compliance errors to string", func() {
		It("should return proper string", func() {
			complianceErrors := &policyCompliance.ComplianceErrors{
				{
					WorkloadField: "fld-a",
					PolicyField:   "fld-a",
					Details:       "details-a",
				},
				{
					WorkloadField: "fld-b",
					PolicyField:   "fld-b2",
					Details:       "details-b",
				},
				{
					Details: "details-x",
				},
			}

			str := complianceErrors.String()
			Expect(str).To(Equal("fld-a -> details-a\nfld-b -> fld-b2: details-b\ndetails-x"))
		})
	})

	Context("compliance errors to array", func() {
		It("should return proper sorted array", func() {
			complianceErrors := &policyCompliance.ComplianceErrors{
				{
					WorkloadField: "fld-b",
					PolicyField:   "fld-b",
					Details:       "details-b",
				},
				{
					WorkloadField: "fld-a",
					PolicyField:   "fld-a2",
					Details:       "details-a",
				},
				{
					Details: "details-x",
				},
			}

			str := complianceErrors.AsSortedArray()
			Expect(str).To(Equal([]string{
				"details-x",
				"fld-a -> fld-a2: [UnknownRule] details-a",
				"fld-b -> [UnknownRule] details-b",
			}))
		})
	})

	Context("blocked storage", func() {
		It("should reject any attempt to use that storage in workload", func() {
			samplePolicy.Rules = &api.WorkspacePolicyRulesV2{
				Storage: &api.SupersetRulesAllOfStorage{
					Nfs: &api.NfssRules{
						Instances: &api.ItemRules{
							CanAdd: convutils.Ref(false),
						},
					},
				},
			}

			finalPolicy := loadEffectivePolicy(samplePolicy)

			values := openapi.SupersetSpec{
				Storage: *openapi.NewNullableWorkspaceSpecSpecAllOfStorage(&openapi.WorkspaceSpecSpecAllOfStorage{
					S3: []openapi.S3Instance{
						{
							Bucket: *openapi.NewNullableString(convutils.Ref("my_bucket")),
							Path:   *openapi.NewNullableString(convutils.Ref("/my/path")),
						},
					},
					Nfs: []openapi.NfsInstance{
						{
							Path:      *openapi.NewNullableString(convutils.Ref("/my/nfs")),
							Server:    *openapi.NewNullableString(convutils.Ref("my-server")),
							MountPath: *openapi.NewNullableString(convutils.Ref("/mount/path")),
						},
					},
				}),
			}

			valuesAsMap, _ := gojsonq.LoadJson(values)

			complianceErrors := ValidateWorkload(&finalPolicy.EffectivePolicy, valuesAsMap)
			Expect(complianceErrors.AsSortedArray()).To(Equal([]string{
				"storage.nfs -> storage.nfs.instances: [canAdd] the administrator prohibited adding new items",
			}))
		})
	})

	Context("re-arrange defaults", func() {
		It("flat", func() {
			asMapDefaults := loadYamlFile("testing/inputs/workspace_defaults.yaml")
			flattened := FlattenInstances(asMapDefaults)
			expectedSpecAsMap := loadYamlFile("testing/inputs/workspace_as_spec.yaml")
			testutils.Expect(flattened).ToEqual(expectedSpecAsMap)
		})
	})

	It("distributed", func() {
		asMapDefaults := loadYamlFile("testing/inputs/distributed_defaults.yaml")
		flattened := FlattenInstances(asMapDefaults)
		expectedSpecAsMap := loadYamlFile("testing/inputs/distributed_as_spec.yaml")
		testutils.Expect(flattened).ToEqual(expectedSpecAsMap)
	})
})

func loadYamlFile(path string) map[string]interface{} {
	text, err := os.ReadFile(path)
	Expect(err).To(BeNil())

	asMap := make(map[string]interface{})
	yamlErr := yaml.Unmarshal(text, &asMap)
	Expect(yamlErr).To(BeNil())

	return asMap
}

func loadEffectivePolicy(source api.WorkspacePolicyDefaultsAndRulesV2) *dbmodel.PolicyWithOptionalEffective {
	result := dbmodel.PolicyWithOptionalEffective{
		EffectivePolicy: dbmodel.EffectivePolicy{},
	}

	defaults, err := gojsonq.LoadJson(source.Defaults)
	Expect(err).To(BeNil())
	result.EffectivePolicy.DefaultsV2 = defaults

	rules, err := gojsonq.LoadJson(source.Rules)
	Expect(err).To(BeNil())
	result.EffectivePolicy.RulesV2 = rules

	defaults = database.CleanMap(defaults)
	rules = database.CleanMap(rules)

	return &result
}

func loadDistributedEffectivePolicy(source api.DistributedPolicyDefaultsAndRulesV2) *dbmodel.PolicyWithOptionalEffective {
	result := dbmodel.PolicyWithOptionalEffective{
		EffectivePolicy: dbmodel.EffectivePolicy{},
	}

	defaults, err := gojsonq.LoadJson(source.Defaults)
	Expect(err).To(BeNil())
	result.EffectivePolicy.DefaultsV2 = defaults

	rules, err := gojsonq.LoadJson(source.Rules)
	Expect(err).To(BeNil())
	result.EffectivePolicy.RulesV2 = rules

	defaults = database.CleanMap(defaults)
	rules = database.CleanMap(rules)

	return &result
}
