/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sanitization utility", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.WithValue(context.Background(), runaictx.CfSubjectID, testing.TestUser)
	})

	Context("SanitizeWorkload", func() {
		It("should sanitize workload when user is not creator", func() {
			// Create a workload with user credentials
			workload := &models_crud.Workload{
				CreatedBy: testing.AnotherUser,
				Spec: &workloadsApi.SupersetSpec{
					EnvironmentVariables: []workloadsApi.EnvironmentVariable{
						{
							Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
							Value: *workloadsApi.NewNullableString(convutils.Ref("value1")),
							UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
								&workloadsApi.EnvironmentVariableUserCredential{
									Name: *workloadsApi.NewNullableString(convutils.Ref("credential1")),
									Key:  *workloadsApi.NewNullableString(convutils.Ref("key1")),
								},
							),
						},
						{
							Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
							Value:          *workloadsApi.NewNullableString(convutils.Ref("value2")),
							UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
						},
					},
					ImagePullSecrets: []workloadsApi.ImagePullSecret{
						{
							Name:           convutils.Ref("secret1"),
							UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(true)),
						},
						{
							Name:           convutils.Ref("secret2"),
							UserCredential: workloadsApi.NullableBool{},
						},
					},
				},
			}

			// Sanitize the workload
			sanitizeWorkload(ctx, workload)

			// Verify that environment variables with user credentials were removed
			Expect(len(workload.Spec.EnvironmentVariables)).To(Equal(1))
			Expect(*workload.Spec.EnvironmentVariables[0].Name.Get()).To(Equal("ENV2"))
			Expect(*workload.Spec.EnvironmentVariables[0].Value.Get()).To(Equal("value2"))
			Expect(workload.Spec.EnvironmentVariables[0].UserCredential.IsSet()).To(BeFalse())

			// Verify that image pull secrets with user credentials were removed
			Expect(len(workload.Spec.ImagePullSecrets)).To(Equal(1))
			Expect(*workload.Spec.ImagePullSecrets[0].Name).To(Equal("secret2"))
			Expect(workload.Spec.ImagePullSecrets[0].UserCredential.IsSet()).To(BeFalse())
		})

		It("should not sanitize workload when user is creator", func() {
			// Create a workload with user credentials
			workload := &models_crud.Workload{
				CreatedBy: testing.TestUser,
				Spec: &workloadsApi.SupersetSpec{
					EnvironmentVariables: []workloadsApi.EnvironmentVariable{
						{
							Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
							Value: *workloadsApi.NewNullableString(convutils.Ref("value1")),
							UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
								&workloadsApi.EnvironmentVariableUserCredential{
									Name: *workloadsApi.NewNullableString(convutils.Ref("credential1")),
									Key:  *workloadsApi.NewNullableString(convutils.Ref("key1")),
								},
							),
						},
						{
							Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
							Value:          *workloadsApi.NewNullableString(convutils.Ref("value2")),
							UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
						},
					},
				},
			}

			// Sanitize the workload
			sanitizeWorkload(ctx, workload)

			// Verify that all environment variables remain unchanged since user is creator
			Expect(len(workload.Spec.EnvironmentVariables)).To(Equal(2))
			Expect(workload.Spec.EnvironmentVariables[0].UserCredential.IsSet()).To(BeTrue())
			Expect(workload.Spec.EnvironmentVariables[1].UserCredential.IsSet()).To(BeFalse())
		})

		It("should sanitize master spec when user is not creator", func() {
			masterSpec := workloadsApi.MasterSpec{
				EnvironmentVariables: []workloadsApi.EnvironmentVariable{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("MASTER_ENV")),
						Value: *workloadsApi.NewNullableString(convutils.Ref("master_value")),
						UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
							&workloadsApi.EnvironmentVariableUserCredential{
								Name: *workloadsApi.NewNullableString(convutils.Ref("master_credential")),
								Key:  *workloadsApi.NewNullableString(convutils.Ref("master_key")),
							},
						),
					},
					{
						Name:           *workloadsApi.NewNullableString(convutils.Ref("MASTER_ENV2")),
						Value:          *workloadsApi.NewNullableString(convutils.Ref("master_value2")),
						UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
					},
				},
				ImagePullSecrets: []workloadsApi.ImagePullSecret{
					{
						Name:           convutils.Ref("master_secret"),
						UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(true)),
					},
					{
						Name:           convutils.Ref("master_secret2"),
						UserCredential: workloadsApi.NullableBool{},
					},
				},
			}

			workload := &models_crud.Workload{
				CreatedBy:  testing.AnotherUser,
				MasterSpec: &masterSpec,
			}

			// Sanitize the workload
			sanitizeWorkload(ctx, workload)

			// Verify that environment variables with user credentials were removed from master spec
			Expect(len(workload.MasterSpec.EnvironmentVariables)).To(Equal(1))
			Expect(*workload.MasterSpec.EnvironmentVariables[0].Name.Get()).To(Equal("MASTER_ENV2"))
			Expect(*workload.MasterSpec.EnvironmentVariables[0].Value.Get()).To(Equal("master_value2"))
			Expect(workload.MasterSpec.EnvironmentVariables[0].UserCredential.IsSet()).To(BeFalse())

			// Verify that image pull secrets with user credentials were removed from master spec
			Expect(len(workload.MasterSpec.ImagePullSecrets)).To(Equal(1))
			Expect(*workload.MasterSpec.ImagePullSecrets[0].Name).To(Equal("master_secret2"))
			Expect(workload.MasterSpec.ImagePullSecrets[0].UserCredential.IsSet()).To(BeFalse())
		})

		It("should handle nil spec gracefully", func() {
			workload := &models_crud.Workload{
				CreatedBy: testing.AnotherUser,
				Spec:      nil,
			}

			// Should not panic
			Expect(func() {
				sanitizeWorkload(ctx, workload)
			}).NotTo(Panic())
		})

		It("should handle nil master spec gracefully", func() {
			workload := &models_crud.Workload{
				CreatedBy:  testing.AnotherUser,
				MasterSpec: nil,
			}

			// Should not panic
			Expect(func() {
				sanitizeWorkload(ctx, workload)
			}).NotTo(Panic())
		})
	})

	Context("SanitizeWorkloadResponse", func() {
		It("should sanitize workload response when user is not creator", func() {
			// Create a workload response with user credentials
			workload := &workloadsApi.SupersetWorkload{
				CreatedBy: testing.AnotherUser,
				Spec: &workloadsApi.SupersetSpec{
					EnvironmentVariables: []workloadsApi.EnvironmentVariable{
						{
							Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
							Value: *workloadsApi.NewNullableString(convutils.Ref("value1")),
							UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
								&workloadsApi.EnvironmentVariableUserCredential{
									Name: *workloadsApi.NewNullableString(convutils.Ref("credential1")),
									Key:  *workloadsApi.NewNullableString(convutils.Ref("key1")),
								},
							),
						},
						{
							Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
							Value:          *workloadsApi.NewNullableString(convutils.Ref("value2")),
							UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
						},
					},
					ImagePullSecrets: []workloadsApi.ImagePullSecret{
						{
							Name:           convutils.Ref("secret1"),
							UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(true)),
						},
						{
							Name:           convutils.Ref("secret2"),
							UserCredential: workloadsApi.NullableBool{},
						},
					},
				},
			}

			// Sanitize the workload response
			sanitizeWorkloadResponse(ctx, workload)

			// Verify that environment variables with user credentials were removed
			Expect(len(workload.Spec.EnvironmentVariables)).To(Equal(1))
			Expect(*workload.Spec.EnvironmentVariables[0].Name.Get()).To(Equal("ENV2"))
			Expect(*workload.Spec.EnvironmentVariables[0].Value.Get()).To(Equal("value2"))
			Expect(workload.Spec.EnvironmentVariables[0].UserCredential.IsSet()).To(BeFalse())

			// Verify that image pull secrets with user credentials were removed
			Expect(len(workload.Spec.ImagePullSecrets)).To(Equal(1))
			Expect(*workload.Spec.ImagePullSecrets[0].Name).To(Equal("secret2"))
			Expect(workload.Spec.ImagePullSecrets[0].UserCredential.IsSet()).To(BeFalse())
		})

		It("should not sanitize workload response when user is creator", func() {
			// Create a workload response with user credentials
			workload := &workloadsApi.SupersetWorkload{
				CreatedBy: testing.TestUser,
				Spec: &workloadsApi.SupersetSpec{
					EnvironmentVariables: []workloadsApi.EnvironmentVariable{
						{
							Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
							Value: *workloadsApi.NewNullableString(convutils.Ref("value1")),
							UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
								&workloadsApi.EnvironmentVariableUserCredential{
									Name: *workloadsApi.NewNullableString(convutils.Ref("credential1")),
									Key:  *workloadsApi.NewNullableString(convutils.Ref("key1")),
								},
							),
						},
						{
							Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
							Value:          *workloadsApi.NewNullableString(convutils.Ref("value2")),
							UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
						},
					},
				},
			}

			// Sanitize the workload response
			sanitizeWorkloadResponse(ctx, workload)

			// Verify that all environment variables remain unchanged since user is creator
			Expect(len(workload.Spec.EnvironmentVariables)).To(Equal(2))
			Expect(workload.Spec.EnvironmentVariables[0].UserCredential.IsSet()).To(BeTrue())
			Expect(workload.Spec.EnvironmentVariables[1].UserCredential.IsSet()).To(BeFalse())
		})

		It("should sanitize master spec in workload response when user is not creator", func() {
			masterSpec := workloadsApi.MasterSpec{
				EnvironmentVariables: []workloadsApi.EnvironmentVariable{
					{
						Name:  *workloadsApi.NewNullableString(convutils.Ref("MASTER_ENV")),
						Value: *workloadsApi.NewNullableString(convutils.Ref("master_value")),
						UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
							&workloadsApi.EnvironmentVariableUserCredential{
								Name: *workloadsApi.NewNullableString(convutils.Ref("master_credential")),
								Key:  *workloadsApi.NewNullableString(convutils.Ref("master_key")),
							},
						),
					},
					{
						Name:           *workloadsApi.NewNullableString(convutils.Ref("MASTER_ENV2")),
						Value:          *workloadsApi.NewNullableString(convutils.Ref("master_value2")),
						UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
					},
				},
				ImagePullSecrets: []workloadsApi.ImagePullSecret{
					{
						Name:           convutils.Ref("master_secret"),
						UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(true)),
					},
					{
						Name:           convutils.Ref("master_secret2"),
						UserCredential: workloadsApi.NullableBool{},
					},
				},
			}

			workload := &workloadsApi.SupersetWorkload{
				CreatedBy:  testing.AnotherUser,
				MasterSpec: *workloadsApi.NewNullableMasterSpec(&masterSpec),
			}

			// Sanitize the workload response
			sanitizeWorkloadResponse(ctx, workload)

			// Verify that environment variables with user credentials were removed from master spec
			Expect(len(workload.MasterSpec.Get().EnvironmentVariables)).To(Equal(1))
			Expect(*workload.MasterSpec.Get().EnvironmentVariables[0].Name.Get()).To(Equal("MASTER_ENV2"))
			Expect(*workload.MasterSpec.Get().EnvironmentVariables[0].Value.Get()).To(Equal("master_value2"))
			Expect(workload.MasterSpec.Get().EnvironmentVariables[0].UserCredential.IsSet()).To(BeFalse())

			// Verify that image pull secrets with user credentials were removed from master spec
			Expect(len(workload.MasterSpec.Get().ImagePullSecrets)).To(Equal(1))
			Expect(*workload.MasterSpec.Get().ImagePullSecrets[0].Name).To(Equal("master_secret2"))
			Expect(workload.MasterSpec.Get().ImagePullSecrets[0].UserCredential.IsSet()).To(BeFalse())
		})

		It("should handle nil spec gracefully", func() {
			workload := &workloadsApi.SupersetWorkload{
				CreatedBy: testing.AnotherUser,
				Spec:      nil,
			}

			// Should not panic
			Expect(func() {
				sanitizeWorkloadResponse(ctx, workload)
			}).NotTo(Panic())
		})

		It("should handle nil master spec gracefully", func() {
			workload := &workloadsApi.SupersetWorkload{
				CreatedBy:  testing.AnotherUser,
				MasterSpec: workloadsApi.NullableMasterSpec{},
			}

			// Should not panic
			Expect(func() {
				sanitizeWorkloadResponse(ctx, workload)
			}).NotTo(Panic())
		})
	})

	Context("sanitizeEnvironmentVariables", func() {
		It("should remove environment variables with user credentials", func() {
			envVars := []workloadsApi.EnvironmentVariable{
				{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
					Value: *workloadsApi.NewNullableString(convutils.Ref("value1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
						&workloadsApi.EnvironmentVariableUserCredential{
							Name: *workloadsApi.NewNullableString(convutils.Ref("credential1")),
							Key:  *workloadsApi.NewNullableString(convutils.Ref("key1")),
						},
					),
				},
				{
					Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
					Value:          *workloadsApi.NewNullableString(convutils.Ref("value2")),
					UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
				},
				{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV3")),
					Value: *workloadsApi.NewNullableString(convutils.Ref("value3")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
						&workloadsApi.EnvironmentVariableUserCredential{
							Name: *workloadsApi.NewNullableString(convutils.Ref("credential3")),
							Key:  *workloadsApi.NewNullableString(convutils.Ref("key3")),
						},
					),
				},
			}

			result := sanitizeEnvironmentVariables(envVars)

			// Verify that only environment variables without user credentials remain
			Expect(len(result)).To(Equal(1))
			Expect(*result[0].Name.Get()).To(Equal("ENV2"))
			Expect(*result[0].Value.Get()).To(Equal("value2"))
			Expect(result[0].UserCredential.IsSet()).To(BeFalse())
		})

		It("should handle nil environment variables", func() {
			// Should not panic
			Expect(func() {
				sanitizeEnvironmentVariables(nil)
			}).NotTo(Panic())
		})

		It("should handle empty environment variables slice", func() {
			envVars := []workloadsApi.EnvironmentVariable{}

			// Should not panic
			Expect(func() {
				sanitizeEnvironmentVariables(envVars)
			}).NotTo(Panic())
		})

		It("should remove all environment variables when all have user credentials", func() {
			envVars := []workloadsApi.EnvironmentVariable{
				{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
					Value: *workloadsApi.NewNullableString(convutils.Ref("value1")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
						&workloadsApi.EnvironmentVariableUserCredential{
							Name: *workloadsApi.NewNullableString(convutils.Ref("credential1")),
							Key:  *workloadsApi.NewNullableString(convutils.Ref("key1")),
						},
					),
				},
				{
					Name:  *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
					Value: *workloadsApi.NewNullableString(convutils.Ref("value2")),
					UserCredential: *workloadsApi.NewNullableEnvironmentVariableUserCredential(
						&workloadsApi.EnvironmentVariableUserCredential{
							Name: *workloadsApi.NewNullableString(convutils.Ref("credential2")),
							Key:  *workloadsApi.NewNullableString(convutils.Ref("key2")),
						},
					),
				},
			}

			result := sanitizeEnvironmentVariables(envVars)

			// Verify that all environment variables were removed
			Expect(len(result)).To(Equal(0))
		})

		It("should keep all environment variables when none have user credentials", func() {
			envVars := []workloadsApi.EnvironmentVariable{
				{
					Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV1")),
					Value:          *workloadsApi.NewNullableString(convutils.Ref("value1")),
					UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
				},
				{
					Name:           *workloadsApi.NewNullableString(convutils.Ref("ENV2")),
					Value:          *workloadsApi.NewNullableString(convutils.Ref("value2")),
					UserCredential: workloadsApi.NullableEnvironmentVariableUserCredential{},
				},
			}

			result := sanitizeEnvironmentVariables(envVars)

			// Verify that all environment variables remain unchanged
			Expect(len(result)).To(Equal(2))
			Expect(*result[0].Name.Get()).To(Equal("ENV1"))
			Expect(*result[0].Value.Get()).To(Equal("value1"))
			Expect(*result[1].Name.Get()).To(Equal("ENV2"))
			Expect(*result[1].Value.Get()).To(Equal("value2"))
		})
	})

	Context("sanitizeImagePullSecrets", func() {
		It("should remove image pull secrets with user credentials", func() {
			imagePullSecrets := []workloadsApi.ImagePullSecret{
				{
					Name:           convutils.Ref("secret1"),
					UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(true)),
				},
				{
					Name:           convutils.Ref("secret2"),
					UserCredential: workloadsApi.NullableBool{},
				},
				{
					Name:           convutils.Ref("secret3"),
					UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(false)),
				},
			}

			result := sanitizeImagePullSecrets(imagePullSecrets)

			// Verify that only image pull secrets without user credentials remain
			Expect(len(result)).To(Equal(1))
			Expect(*result[0].Name).To(Equal("secret2"))
			Expect(result[0].UserCredential.IsSet()).To(BeFalse())
		})

		It("should handle nil image pull secrets", func() {
			// Should not panic
			Expect(func() {
				sanitizeImagePullSecrets(nil)
			}).NotTo(Panic())
		})

		It("should handle empty image pull secrets slice", func() {
			imagePullSecrets := []workloadsApi.ImagePullSecret{}

			// Should not panic
			Expect(func() {
				sanitizeImagePullSecrets(imagePullSecrets)
			}).NotTo(Panic())
		})

		It("should remove all image pull secrets when all have user credentials", func() {
			imagePullSecrets := []workloadsApi.ImagePullSecret{
				{
					Name:           convutils.Ref("secret1"),
					UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(true)),
				},
				{
					Name:           convutils.Ref("secret2"),
					UserCredential: *workloadsApi.NewNullableBool(convutils.Ref(false)),
				},
			}

			result := sanitizeImagePullSecrets(imagePullSecrets)

			// Verify that all image pull secrets were removed
			Expect(len(result)).To(Equal(0))
		})

		It("should keep all image pull secrets when none have user credentials", func() {
			imagePullSecrets := []workloadsApi.ImagePullSecret{
				{
					Name:           convutils.Ref("secret1"),
					UserCredential: workloadsApi.NullableBool{},
				},
				{
					Name:           convutils.Ref("secret2"),
					UserCredential: workloadsApi.NullableBool{},
				},
			}

			result := sanitizeImagePullSecrets(imagePullSecrets)

			// Verify that all image pull secrets remain unchanged
			Expect(len(result)).To(Equal(2))
			Expect(*result[0].Name).To(Equal("secret1"))
			Expect(*result[1].Name).To(Equal("secret2"))
		})
	})
})
