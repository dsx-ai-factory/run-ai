/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"bytes"
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"k8s.io/utils/ptr"
)

const (
	workloadName  = "test-workload"
	testNamespace = "runai-default"
	image         = "test-image"
)

var _ = Describe("Workload Parameters Validator", func() {
	var (
		workload  *jobs.JobWorkload
		validator *GenericParamsValidator
		ctx       context.Context
	)
	Context("workload validation", func() {

		BeforeEach(func() {
			workload = workloadUtils.NewJobWorkloadBuilder().WithName(workloadName).WithImage(image).WithTraining().WithNamespace(testNamespace).WithLabels(nil).Build()
			ctx = context.Background()
			validator = NewGenericParamsValidator()
		})
		It("workload should ne invalid labels bigger than 63", func() {

			workload.Spec.Usage = common.StatusSubmitted
			workload.Labels = map[string]string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			}

			err := validator.Validate(ctx, workload)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("label key and value must not be longer than 63 character"))
		})
		It("workload should ne invalid labels with empty key", func() {

			workload.Spec.Usage = common.StatusSubmitted
			workload.Labels = map[string]string{
				"": "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
			}

			err := validator.Validate(ctx, workload)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to validate labels"))
		})
		It("workload should ne invalid labels with empty value", func() {

			workload.Spec.Usage = common.StatusSubmitted
			workload.Labels = map[string]string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.": "",
			}

			err := validator.Validate(ctx, workload)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to validate labels"))
		})
		It("workload invalid, usage is template and base is not empty", func() {

			workload.Spec.Usage = common.UsedAsTemplate
			workload.Spec.BaseWorkload = ptr.To("base")

			err := validator.Validate(ctx, workload)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("baseWorkload is only allowed for workloads with usage=Submit"))
		})
		It("workload valid, usage is template and base is empty", func() {

			workload.Spec.Usage = common.UsedAsTemplate
			workload.Spec.BaseWorkload = nil

			err := validator.Validate(ctx, workload)

			Expect(err).To(BeNil())
		})
		It("workload invalid, pod affinity", func() {
			workload.Spec.RunaiSchedulerFields = field_groups.RunaiSchedulerFields{}
			workload.Spec.RunaiSchedulerFields.PodAffinity = &field_types.BoolField{Value: ptr.To(true)}
			err := validator.Validate(ctx, workload)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("podAffinityTopology and podAffinitySchedulingRule can be used only when podAffinity is enabled"))
		})
		It("should fail to validate when image is empty", func() {
			workload.Spec.ContainerFields.Image.Value = ptr.To("")
			err := validator.Validate(ctx, workload)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("image must be provided"))
		})
		It("should not fail to validate when image is not empty", func() {
			err := validator.Validate(ctx, workload)
			Expect(err).To(BeNil())
		})

		It("should fail to validate when memory quantity is invalid", func() {
			workload.Spec.GpuMemory = setQuantityField("4i")
			err := validator.Validate(ctx, workload)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid container resource, invalid GPU memory (4i): unable to parse quantity's suffix"))
		})

		It("should fail to validate when master memory quantity is invalid", func() {
			workload.MasterSpec = &jobs.JobSpec{}
			workload.MasterSpec.GpuMemory = setQuantityField("4i")
			err := validator.Validate(ctx, workload)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("invalid container resource, invalid master GPU memory (4i): unable to parse quantity's suffix"))
		})
	})

	Context("ValidateResourceRequirements", func() {
		var resourceFields field_groups.ResourceAllocationFields

		BeforeEach(func() {
			resourceFields = field_groups.ResourceAllocationFields{
				Memory:      &field_types.QuantityField{},
				MemoryLimit: &field_types.QuantityField{},
				Cpu:         &field_types.NumberField{},
				CpuLimit:    &field_types.NumberField{},
			}
		})

		Context("CPU compute", func() {
			It("should be a valid workload with empty CPU limit compute", func() {
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with only with CPU compute", func() {
				resourceFields.Cpu.Value = ptr.To("1")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with only with CPU compute limit", func() {
				resourceFields.CpuLimit.Value = ptr.To("1")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with same CPU compute and CPU limit compute", func() {
				resourceFields.Cpu.Value = ptr.To("1")
				resourceFields.CpuLimit.Value = ptr.To("1")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with CPU compute and higher CPU limit compute", func() {
				resourceFields.Cpu.Value = ptr.To("1")
				resourceFields.CpuLimit.Value = ptr.To("2")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be a invalid workload with CPU compute and lower CPU limit compute", func() {
				resourceFields.Cpu.Value = ptr.To("2")
				resourceFields.CpuLimit.Value = ptr.To("1")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("CPU compute limit (1) must be equal to or higher than CPU compute request (2)"))
			})

			It("should be a invalid workload with CPU compute and lower float CPU limit compute", func() {
				resourceFields.Cpu.Value = ptr.To("2.5")
				resourceFields.CpuLimit.Value = ptr.To("1.5")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("CPU compute limit (1500m) must be equal to or higher than CPU compute request (2500m)"))
			})

			It("should be a valid workload with only CPU compute set (limit unset)", func() {
				resourceFields.Cpu.Value = ptr.To("2")
				resourceFields.CpuLimit.Value = nil
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with only CPU limit set (request unset)", func() {
				resourceFields.Cpu.Value = nil
				resourceFields.CpuLimit.Value = ptr.To("2")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).To(BeNil())
			})

			It("should be invalid if CPU request is negative and limit is unset", func() {
				resourceFields.Cpu.Value = ptr.To("-1")
				resourceFields.CpuLimit.Value = nil
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("CPU compute (-1) must be positive"))
			})

			It("should be invalid if CPU limit is negative and request is unset", func() {
				resourceFields.Cpu.Value = nil
				resourceFields.CpuLimit.Value = ptr.To("-1")
				err := ValidateCPUResourceRequirements(resourceFields, "")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("CPU compute limit (-1) must be positive"))
			})
		})
		Context("Image Validation", func() {
			Context("when the image is empty", func() {
				It("returns an error", func() {
					err := ValidateImage("")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("image must be provided"))
				})
			})

			Context("when the image is not empty", func() {
				It("does not return an error", func() {
					err := ValidateImage("valid-image")
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("GitSync Validation", func() {
			workloadGitSyncItems := map[string]field_types.GitSyncItemAndRules{
				"idx1": {
					Value: &field_types.GitSyncItemValue{
						Repository: ptr.To("https://github.com/example/repository"),
						Revision:   ptr.To(""),
						Branch:     ptr.To(""),
					},
				},
			}

			Context("git ref validation", func() {
				It("when only revision is provided, does not return an error", func() {
					*workloadGitSyncItems["idx1"].Value.Revision = "v1.2"
					err := ValidateGitRevision(workloadGitSyncItems)
					Expect(err).NotTo(HaveOccurred())
				})
				It("when only branch is provided, does not return an error", func() {
					*workloadGitSyncItems["idx1"].Value.Branch = "main"
					*workloadGitSyncItems["idx1"].Value.Revision = ""
					err := ValidateGitRevision(workloadGitSyncItems)
					Expect(err).NotTo(HaveOccurred())
				})
				It("both revision and branch are provided, returns an error", func() {
					*workloadGitSyncItems["idx1"].Value.Branch = "main"
					*workloadGitSyncItems["idx1"].Value.Revision = "v1.2"
					err := ValidateGitRevision(workloadGitSyncItems)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("please provide only one of the following: 'branch' or 'revision'"))
				})
			})

			Context("git authentication validation", func() {
				It("without authentication , does not return an error", func() {
					err := ValidateGitAuthenticatingMethod(workloadGitSyncItems)
					Expect(err).NotTo(HaveOccurred())
				})

				It("with password authentication", func() {
					workloadGitSyncItems["idx1"].Value.AuthenticatingMethod = ptr.To(constants.GitAuthenticatingMethodPassword)
					err := ValidateGitAuthenticatingMethod(workloadGitSyncItems)
					Expect(err).NotTo(HaveOccurred())
				})

				It("with ssh-key authentication", func() {
					workloadGitSyncItems["idx1"].Value.AuthenticatingMethod = ptr.To(constants.GitAuthenticatingMethodSSHKey)
					workloadGitSyncItems["idx1"].Value.SecretName = ptr.To("secret-name")
					err := ValidateGitAuthenticatingMethod(workloadGitSyncItems)
					Expect(err).NotTo(HaveOccurred())
				})

				It("with invalid authentication", func() {
					workloadGitSyncItems["idx1"].Value.AuthenticatingMethod = ptr.To("pass")
					err := ValidateGitAuthenticatingMethod(workloadGitSyncItems)
					Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("authenticating method must be one of the following: '%s' or '%s'", constants.GitAuthenticatingMethodPassword, constants.GitAuthenticatingMethodSSHKey)))
				})

				It("with ssh-key authentication but no secret", func() {
					workloadGitSyncItems["idx1"].Value.AuthenticatingMethod = ptr.To(constants.GitAuthenticatingMethodSSHKey)
					workloadGitSyncItems["idx1"].Value.SecretName = nil
					err := ValidateGitAuthenticatingMethod(workloadGitSyncItems)
					Expect(err.Error()).To(ContainSubstring("when authenticating with ssh, a secret to the SSH key must be provided"))
				})

			})
		})

		Context("CPU memory", func() {
			It("should be a valid workload with empty CPU memory limit", func() {
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with only with CPU memory", func() {
				resourceFields.Memory.Value = ptr.To("100m")
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with only with CPU memory limit", func() {
				resourceFields.MemoryLimit.Value = ptr.To("100m")
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with same CPU memory and CPU limit memory", func() {
				resourceFields.Memory.Value = ptr.To("100m")
				resourceFields.MemoryLimit.Value = ptr.To("100m")
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with CPU memory and higher CPU limit memory", func() {
				resourceFields.Memory.Value = ptr.To("100m")
				resourceFields.MemoryLimit.Value = ptr.To("200m")
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).To(BeNil())
			})

			It("should be a invalid workload with CPU memory and lower CPU limit memory", func() {
				resourceFields.Memory.Value = ptr.To("200m")
				resourceFields.MemoryLimit.Value = ptr.To("100m")
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("CPU memory limit (100m) must be equal to or higher than CPU memory request (200m)"))
			})

			It("should be a invalid workload with CPU memory in Megabyte and lower CPU limit memory in Byte", func() {
				resourceFields.Memory.Value = ptr.To("100M")
				resourceFields.MemoryLimit.Value = ptr.To("100m")
				err := validateMemoryValue(resourceFields.Memory, resourceFields.MemoryLimit, specMemoryName, "")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("CPU memory limit (100m) must be equal to or higher than CPU memory request (100M)"))
			})

		})
	})

	Context("ValidatePVCs", func() {
		var storageFields field_groups.StorageFields

		BeforeEach(func() {
			storageFields = field_groups.StorageFields{
				Pvcs: &field_types.PvcsField{},
			}
		})

		Context("PVC size", func() {
			It("should be a valid workload with empty PVCs struct", func() {
				err := ValidatePVCs(storageFields)
				Expect(err).To(BeNil())
			})

			It("should be an invalid workload with empty PVC value", func() {
				storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{}
				err := ValidatePVCs(storageFields)
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with existing PVC", func() {
				storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
					"pvc-1": {
						Value: &field_types.PvcsItemValue{
							ExistingPvc: ptr.To(true),
						},
					},
				}
				err := ValidatePVCs(storageFields)
				Expect(err).To(BeNil())
			})

			It("should be an invalid workload with nil PVC size field", func() {
				storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
					"pvc-1": {
						Value: &field_types.PvcsItemValue{},
					},
				}

				err := ValidatePVCs(storageFields)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("size is empty"))
			})

			It("should be an invalid workload with 0 values size PVC", func() {
				emptySizes := []string{"0", "-1", "0M", "-1M", "-1G", "0G", "0K", "0Ki", "0Mi", "0Gi", "0m", "0k", "0ki", "0mi", "0gi"}
				for _, size := range emptySizes {
					storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
						"pvc-1": {
							Value: &field_types.PvcsItemValue{
								Size: ptr.To(size),
							},
						},
					}

					err := ValidatePVCs(storageFields)
					Expect(err).ToNot(BeNil())
					Expect(err.Error()).To(ContainSubstring("size is invalid"))
				}
			})

			It("should be a valid workload with none zeros value size PVC", func() {
				sizes := []string{"10", "10M", "10G", "10Mi", "10Gi", "10m", "10k"}
				for _, size := range sizes {
					storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
						"pvc-1": {
							Value: &field_types.PvcsItemValue{
								Size: ptr.To(size),
							},
						},
					}

					err := ValidatePVCs(storageFields)
					Expect(err).To(BeNil())
				}
			})

			It("should be a valid workload with multiple PVCs", func() {
				storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
					"pvc-1": {
						Value: &field_types.PvcsItemValue{
							Size: ptr.To("10M"),
						},
					},
					"pvc-2": {
						Value: &field_types.PvcsItemValue{
							Size: ptr.To("10G"),
						},
					},
					"pvc-3": {
						Value: &field_types.PvcsItemValue{
							Size: ptr.To("10k"),
						},
					},
				}

				err := ValidatePVCs(storageFields)
				Expect(err).To(BeNil())

			})
			It("should be an invalid workload with invalid PVC name", func() {
				storageFields.Pvcs.Value = &map[string]field_types.PvcsItemAndRules{
					"invalid_pvc_name": {
						Value: &field_types.PvcsItemValue{
							Size:        ptr.To("10M"),
							ExistingPvc: ptr.To(false),
							ClaimName:   ptr.To("invalid_pvc_name"),
						},
					},
				}

				err := ValidatePVCs(storageFields)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("claim name is invalid"))
			})
		})
		Context("validate pvc claim name", func() {
			var _ = Describe("validatePVCName", func() {
				It("should return nil when pvcName is empty", func() {
					err := validatePVCName("")
					Expect(err).To(BeNil())
				})

				It("should return nil when pvcName is a valid DNS1123 subdomain", func() {
					err := validatePVCName("valid-pvc-name")
					Expect(err).To(BeNil())
				})

				It("should return error when pvcName is not a valid DNS1123 subdomain", func() {
					err := validatePVCName("invalid_pvc_name")
					Expect(err).To(HaveOccurred())
				})

				It("should return error when pvcName is too long", func() {
					longPVCName := "this-is-a-very-long-pvc-name-that-is-more-than-sixty-three-characters"
					var buffer bytes.Buffer
					for i := 0; i < 20; i++ {
						buffer.WriteString(longPVCName)
					}
					err := validatePVCName(buffer.String())
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	Context("ValidateProbes", func() {
		var containerFields field_groups.ContainerFields

		BeforeEach(func() {
			containerFields = field_groups.ContainerFields{
				ReadinessProbe: &field_types.ProbeField{},
			}
		})

		Context("valid probes", func() {
			It("should be a valid workload with no probe", func() {
				containerFields.ReadinessProbe = nil
				err := ValidateProbes(containerFields)
				Expect(err).To(BeNil())
			})

			It("should be a valid workload with valid probe", func() {
				containerFields.ReadinessProbe.Values = &field_types.ProbeMembers{
					HttpGetPort: field_types.IntField{
						Value: ptr.To(8080),
					},
					HttpGetPath: &field_types.StringField{
						Value: ptr.To("/ready"),
					},
				}
				err := ValidateProbes(containerFields)
				Expect(err).To(BeNil())
			})

			It("should be an invalid workload with invalid path given to probe", func() {
				containerFields.ReadinessProbe.Values = &field_types.ProbeMembers{
					HttpGetPort: field_types.IntField{
						Value: ptr.To(8080),
					},
					HttpGetPath: &field_types.StringField{
						Value: ptr.To("invalid/path"),
					},
				}
				err := ValidateProbes(containerFields)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("httpGetPath must start with /"))
			})
		})
	})

	Context("ValidateAuthorizedUsersAndGroups", func() {
		It("when both AuthorizedUsers and AuthorizedGroups ", func() {
			exposedUrls := flds.ExposedUrlsField{
				Value: &map[string]flds.ExposedUrlItemAndRules{
					"idx1": {
						Value: &flds.ExposedUrlItemValue{
							AuthorizedUsers:  ptr.To("user1"),
							AuthorizedGroups: ptr.To("group1"),
						},
					},
				},
			}
			err := ValidateAuthorizedUsersAndGroups(&exposedUrls, nil)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("authorizedUsers and authorizedGroups are mutually exclusive"))
		})

		It("when AuthorizedUsers is set with valid value", func() {
			exposedUrls := flds.ExposedUrlsField{
				Value: &map[string]flds.ExposedUrlItemAndRules{
					"idx1": {
						Value: &flds.ExposedUrlItemValue{
							AuthorizedUsers: ptr.To("user1"),
						},
					},
				},
			}
			err := ValidateAuthorizedUsersAndGroups(&exposedUrls, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("when AuthorizedGroups is set with valid value", func() {
			exposedUrls := flds.ExposedUrlsField{
				Value: &map[string]flds.ExposedUrlItemAndRules{
					"idx1": {
						Value: &flds.ExposedUrlItemValue{
							AuthorizedGroups: ptr.To("group1"),
						},
					},
				},
			}
			err := ValidateAuthorizedUsersAndGroups(&exposedUrls, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("when AuthorizedGroups is set with invalid value", func() {
			exposedUrls := flds.ExposedUrlsField{
				Value: &map[string]flds.ExposedUrlItemAndRules{
					"idx1": {
						Value: &flds.ExposedUrlItemValue{
							AuthorizedGroups: ptr.To("/group1"),
						},
					},
				},
			}
			err := ValidateAuthorizedUsersAndGroups(&exposedUrls, nil)
			Expect(err).To(HaveOccurred())
		})

		It("when ServingPort auth is valid - public", func() {
			auth := field_groups.RoutingAndAuthorizationFields{
				ServingPortAuthorizationType: &flds.StringField{
					Value: ptr.To("public"),
				},
			}
			err := ValidateAuthorizedUsersAndGroups(nil, &auth)
			Expect(err).ToNot(HaveOccurred())
		})

		It("when ServingPort auth is valid - authenticated users", func() {
			auth := field_groups.RoutingAndAuthorizationFields{
				ServingPortAuthorizationType: &flds.StringField{
					Value: ptr.To("authenticatedUsers"),
				},
			}
			err := ValidateAuthorizedUsersAndGroups(nil, &auth)
			Expect(err).ToNot(HaveOccurred())
		})

		It("when ServingPort auth is valid - auth users or groups with users", func() {
			auth := field_groups.RoutingAndAuthorizationFields{
				ServingPortAuthorizationType: &flds.StringField{
					Value: ptr.To("authorizedUsersOrGroups"),
				},
				ServingPortAuthorizedUsers: &flds.StringField{
					Value: ptr.To("user1"),
				},
			}
			err := ValidateAuthorizedUsersAndGroups(nil, &auth)
			Expect(err).ToNot(HaveOccurred())
		})

		It("when ServingPort auth is valid - auth users or groups with groups", func() {
			auth := field_groups.RoutingAndAuthorizationFields{
				ServingPortAuthorizationType: &flds.StringField{
					Value: ptr.To("authorizedUsersOrGroups"),
				},
				ServingPortAuthorizedGroups: &flds.StringField{
					Value: ptr.To("group1"),
				},
			}
			err := ValidateAuthorizedUsersAndGroups(nil, &auth)
			Expect(err).ToNot(HaveOccurred())
		})

		It("when ServingPort auth is invalid - auth users used with the wrong type", func() {
			auth := field_groups.RoutingAndAuthorizationFields{
				ServingPortAuthorizationType: &flds.StringField{
					Value: ptr.To("public"),
				},
				ServingPortAuthorizedUsers: &flds.StringField{
					Value: ptr.To("user1"),
				},
			}
			err := ValidateAuthorizedUsersAndGroups(nil, &auth)
			Expect(err).To(HaveOccurred())
		})

		It("when ServingPort auth is invalid - both auth users and groups", func() {
			auth := field_groups.RoutingAndAuthorizationFields{
				ServingPortAuthorizationType: &flds.StringField{
					Value: ptr.To("authorizedUsersOrGroups"),
				},
				ServingPortAuthorizedUsers: &flds.StringField{
					Value: ptr.To("user1"),
				},
				ServingPortAuthorizedGroups: &flds.StringField{
					Value: ptr.To("group1"),
				},
			}
			err := ValidateAuthorizedUsersAndGroups(nil, &auth)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ValidateWorkloadContainerResourceValues", func() {
		BeforeEach(func() {
			workload = workloadUtils.NewJobWorkloadBuilder().WithName(workloadName).WithImage(image).WithTraining().WithNamespace(testNamespace).WithLabels(nil).Build()
		})
		// there isn't much validation needed to be done, just to see the naive way of use
		Context("Valid", func() {
			It("no memory set", func() {
				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).ToNot(HaveOccurred())
			})

			It("Mega", func() {
				workload.Spec.Memory = setQuantityField("4M")
				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).ToNot(HaveOccurred())
			})

			It("Gi", func() {
				workload.Spec.Memory = setQuantityField("4Gi")

				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("Invalid", func() {
			It("Memory quantity with space", func() {
				workload.Spec.Memory = setQuantityField("4 M")
				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid CPU memory (4 M): quantities must match the regular expression"))
			})

			It("Memory quantity bad suffix", func() {
				workload.Spec.MemoryLimit = setQuantityField("4i")
				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid CPU memory limit (4i): unable to parse quantity's suffix"))
			})

			It("GPU memory limit 3 letters quantity", func() {
				workload.Spec.GpuMemoryLimit = setQuantityField("4imk")
				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid GPU memory limit (4imk): unable to parse quantity's suffix"))
			})

			It("GPU memory with plus suffix", func() {
				workload.Spec.GpuMemory = setQuantityField("4g+")
				err := ValidateWorkloadContainerResourceValues(&workload.Spec.ResourceAllocationFields, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid GPU memory (4g+): quantities must match the regular expression"))
			})

			It("Master GPU memory with plus suffix", func() {
				workload.MasterSpec = &jobs.JobSpec{}
				workload.MasterSpec.GpuMemory = setQuantityField("4g+")
				err := ValidateWorkloadContainerResourceValues(&workload.MasterSpec.ResourceAllocationFields, "master ")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid master GPU memory (4g+): quantities must match the regular expression"))
			})

			It("Master GPU memory with zero value", func() {
				workload.MasterSpec = &jobs.JobSpec{}
				workload.MasterSpec.GpuMemory = setQuantityField("0")
				err := ValidateWorkloadContainerResourceValues(&workload.MasterSpec.ResourceAllocationFields, "")
				Expect(err).ToNot(HaveOccurred())
			})

			It("Master GPU memory with negative value", func() {
				workload.MasterSpec = &jobs.JobSpec{}
				workload.MasterSpec.GpuMemory = setQuantityField("-1")
				err := ValidateWorkloadContainerResourceValues(&workload.MasterSpec.ResourceAllocationFields, "master ")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid master GPU memory (-1): must be positive"))
			})
		})
	})

	Describe("validateMemoryValue", func() {
		It("should be valid if both request and limit are unset (nil)", func() {
			err := validateMemoryValue(nil, nil, "memory", "")
			Expect(err).To(BeNil())
		})

		It("should be valid if request=0 and limit>0", func() {
			err := validateMemoryValue(setQuantityField("0"), setQuantityField("100"), "memory", "")
			Expect(err).To(BeNil())
		})

		It("should be valid if request=0 and limit=0", func() {
			err := validateMemoryValue(setQuantityField("0"), setQuantityField("0"), "memory", "")
			Expect(err).To(BeNil())
		})

		It("should be valid if request>0 and limit is unset", func() {
			err := validateMemoryValue(setQuantityField("100"), nil, "memory", "")
			Expect(err).To(BeNil())
		})

		It("should be valid if request>0 and limit>0 and limit >= request", func() {
			err := validateMemoryValue(setQuantityField("100"), setQuantityField("200"), "memory", "")
			Expect(err).To(BeNil())
		})

		It("should be invalid if request>0 and limit>0 and limit < request", func() {
			err := validateMemoryValue(setQuantityField("200"), setQuantityField("100"), "memory", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("limit (100) must be equal to or higher than memory request (200)"))
		})

		It("should be invalid if request<0 and limit>0", func() {
			err := validateMemoryValue(setQuantityField("-5"), setQuantityField("100"), "memory", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid memory (-5): must be positive"))
		})

		It("should be invalid if request>0 and limit<0", func() {
			err := validateMemoryValue(setQuantityField("100"), setQuantityField("-5"), "memory", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid memory limit (-5): must be positive"))
		})

		It("should be invalid if request=0 and limit<0", func() {
			err := validateMemoryValue(setQuantityField("0"), setQuantityField("-5"), "memory", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid memory limit (-5): must be positive"))
		})

		It("should be invalid if request<0 and limit is unset", func() {
			err := validateMemoryValue(setQuantityField("-5"), nil, "memory", "")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("invalid memory (-5): must be positive"))
		})
	})
})

func setQuantityField(quantity string) *flds.QuantityField {
	return &flds.QuantityField{
		Value: ptr.To(quantity),
	}
}
