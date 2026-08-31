/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	featuregate "github.com/run-ai/runai/runai-cluster/cluster/workload-common/feature-gate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var _ = Describe("S3 tests", func() {
	var (
		ctx          context.Context
		sharedFields *field_groups.WorkloadFields
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()
		sharedFields = &field_groups.WorkloadFields{
			StorageFields: field_groups.StorageFields{
				S3buckets: &field_types.S3bucketField{
					Value: &map[string]field_types.S3bucketItemAndRules{},
				},
			},
		}

		// Disable feature gate for consistent testing
		featureGateUtil := featuregate.GetFeatureGateUtilInstance()
		featureGateUtil.FeatureGateSidecarContainersEnable = false
	})

	addS3Bucket := func(id string, bucket, path, key, secret, url, secretName, image string, sessionToken *string) {
		(*sharedFields.S3buckets.Value)[id] = field_types.S3bucketItemAndRules{
			Value: &field_types.S3bucketItemValue{
				Bucket:       &bucket,
				Path:         &path,
				Key:          &key,
				Secret:       &secret,
				Url:          &url,
				SecretName:   &secretName,
				SessionToken: sessionToken,
				Image:        &image,
			},
		}
	}

	Context("Basic functionality", func() {
		It("should return nil when no S3 buckets are specified", func() {
			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(BeNil())
		})

		It("should create a container with plain text credentials", func() {
			// Add an S3 bucket with plain text credentials
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]
			Expect(container.Name).To(Equal("s3-fs-0"))
			Expect(container.Image).NotTo(BeEmpty())
			Expect(container.Resources).To(Equal(corev1.ResourceRequirements{}))
			Expect(container.SecurityContext.Privileged).To(Equal(ptr.To(true)))

			// Verify env vars for plain text credentials
			hasAccessKey := false
			hasSecretKey := false
			for _, env := range container.Env {
				if env.Name == "AWS_ACCESS_KEY_ID" {
					hasAccessKey = true
					Expect(env.Value).To(Equal("access-key"))
				}
				if env.Name == "AWS_SECRET_ACCESS_KEY" {
					hasSecretKey = true
					Expect(env.Value).To(Equal("secret-key"))
				}
			}
			Expect(hasAccessKey).To(BeTrue())
			Expect(hasSecretKey).To(BeTrue())

			// Verify command includes bucket and path
			Expect(container.Command).To(ContainElement("my-bucket"))
			Expect(container.Command).To(ContainElement("/mount/path"))

			// Verify volume mounts
			Expect(container.VolumeMounts).To(HaveLen(1))
			Expect(container.VolumeMounts[0].MountPath).To(Equal("/mount/path"))
		})

		It("should create a container with secret-based credentials", func() {
			// Add an S3 bucket with secret-based credentials
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key-field", "secret-key-field", "", "my-secret", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]

			// Verify env vars for secret-based credentials
			hasAccessKey := false
			hasSecretKey := false
			for _, env := range container.Env {
				if env.Name == "AWS_ACCESS_KEY_ID" {
					hasAccessKey = true
					Expect(env.ValueFrom).NotTo(BeNil())
					Expect(env.ValueFrom.SecretKeyRef.Name).To(Equal("my-secret"))
					Expect(env.ValueFrom.SecretKeyRef.Key).To(Equal("access-key-field"))
				}
				if env.Name == "AWS_SECRET_ACCESS_KEY" {
					hasSecretKey = true
					Expect(env.ValueFrom).NotTo(BeNil())
					Expect(env.ValueFrom.SecretKeyRef.Name).To(Equal("my-secret"))
					Expect(env.ValueFrom.SecretKeyRef.Key).To(Equal("secret-key-field"))
				}
			}
			Expect(hasAccessKey).To(BeTrue())
			Expect(hasSecretKey).To(BeTrue())
		})

		It("should create a container with secret-based credentials + session token", func() {
			// Add an S3 bucket with secret-based credentials
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key-field", "secret-key-field", "", "my-secret", "", ptr.To("my-session-token"))

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]

			// Verify env vars for secret-based credentials
			hasAccessKey := false
			hasSecretKey := false
			hasSessionToken := false
			for _, env := range container.Env {
				if env.Name == "AWS_ACCESS_KEY_ID" {
					hasAccessKey = true
					Expect(env.ValueFrom).NotTo(BeNil())
					Expect(env.ValueFrom.SecretKeyRef.Name).To(Equal("my-secret"))
					Expect(env.ValueFrom.SecretKeyRef.Key).To(Equal("access-key-field"))
				}
				if env.Name == "AWS_SECRET_ACCESS_KEY" {
					hasSecretKey = true
					Expect(env.ValueFrom).NotTo(BeNil())
					Expect(env.ValueFrom.SecretKeyRef.Name).To(Equal("my-secret"))
					Expect(env.ValueFrom.SecretKeyRef.Key).To(Equal("secret-key-field"))
				}
				if env.Name == "AWS_SESSION_TOKEN" {
					hasSessionToken = true
					Expect(env.Value).To(Equal("my-session-token"))
				}
			}
			Expect(hasAccessKey).To(BeTrue())
			Expect(hasSecretKey).To(BeTrue())
			Expect(hasSessionToken).To(BeTrue())
		})
	})

	Context("Command configuration", func() {
		It("should include custom endpoint when URL is provided", func() {
			// Add an S3 bucket with custom endpoint
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "https://custom-s3.example.com", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]

			// Verify command includes endpoint
			endpointIndex := -1
			for i, arg := range container.Command {
				if arg == "--endpoint" {
					endpointIndex = i
					break
				}
			}
			Expect(endpointIndex).To(BeNumerically(">", 0))
			Expect(container.Command[endpointIndex+1]).To(Equal("https://custom-s3.example.com"))
		})

		It("should include UID/GID when provided in shared fields", func() {
			// Add an S3 bucket
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)

			// Set UID/GID in shared fields
			uid := 1000
			gid := 2000
			sharedFields.RunAsUid = &field_types.IntField{Value: ptr.To(uid)}
			sharedFields.RunAsGid = &field_types.IntField{Value: ptr.To(gid)}

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]

			// Verify command includes UID/GID
			uidIndex := -1
			gidIndex := -1
			for i, arg := range container.Command {
				if arg == "--uid" {
					uidIndex = i
				}
				if arg == "--gid" {
					gidIndex = i
				}
			}
			Expect(uidIndex).To(BeNumerically(">", 0))
			Expect(gidIndex).To(BeNumerically(">", 0))
			Expect(container.Command[uidIndex+1]).To(Equal("1000"))
			Expect(container.Command[gidIndex+1]).To(Equal("2000"))
		})

		It("should use custom image when provided", func() {
			// Add an S3 bucket with custom image
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "custom-s3-image:latest", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]
			Expect(container.Image).To(Equal("custom-s3-image:latest"))
		})
	})

	Context("Multiple S3 buckets", func() {
		It("should create multiple containers for multiple buckets", func() {
			// Add multiple S3 buckets
			addS3Bucket("bucket1", "bucket-1", "/mount/path1", "key1", "secret1", "", "", "", nil)
			addS3Bucket("bucket2", "bucket-2", "/mount/path2", "key2", "secret2", "", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(2))

			// Verify first container
			Expect(containers[0].Name).To(Equal("s3-fs-0"))
			Expect(containers[0].VolumeMounts[0].MountPath).To(Equal("/mount/path1"))

			// Verify second container
			Expect(containers[1].Name).To(Equal("s3-fs-1"))
			Expect(containers[1].VolumeMounts[0].MountPath).To(Equal("/mount/path2"))
		})

		It("should skip nil bucket values", func() {
			// Add valid bucket
			addS3Bucket("bucket1", "bucket-1", "/mount/path1", "key1", "secret1", "", "", "", nil)

			// Add nil bucket
			(*sharedFields.S3buckets.Value)["bucket2"] = field_types.S3bucketItemAndRules{
				Value: nil,
			}

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))
			Expect(containers[0].VolumeMounts[0].MountPath).To(Equal("/mount/path1"))
		})
	})

	Context("FIPS GODEBUG propagation", func() {
		findGodebug := func(env []corev1.EnvVar) (string, bool) {
			for _, e := range env {
				if e.Name == config.EnvVarGodebug {
					return e.Value, true
				}
			}
			return "", false
		}

		It("should propagate the controller's GODEBUG to the goofys container", func() {
			GinkgoT().Setenv(config.EnvVarGodebug, "fips140=on")
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			value, found := findGodebug(containers[0].Env)
			Expect(found).To(BeTrue())
			Expect(value).To(Equal("fips140=on"))
		})

		It("should not set GODEBUG when the controller has none", func() {
			GinkgoT().Setenv(config.EnvVarGodebug, "")
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			_, found := findGodebug(containers[0].Env)
			Expect(found).To(BeFalse())
		})
	})

	Context("Feature Gate", func() {
		It("should set restart policy when feature gate is enabled", func() {
			// Enable feature gate
			featureGateUtil := featuregate.GetFeatureGateUtilInstance()
			featureGateUtil.FeatureGateSidecarContainersEnable = true

			// Add an S3 bucket
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]
			Expect(container.RestartPolicy).NotTo(BeNil())
			Expect(*container.RestartPolicy).To(Equal(corev1.ContainerRestartPolicyAlways))
		})

		It("should not set restart policy when feature gate is disabled", func() {
			// Feature gate is disabled by default in BeforeEach

			// Add an S3 bucket
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)

			containers := GetS3container(ctx, sharedFields)
			Expect(containers).To(HaveLen(1))

			container := containers[0]
			Expect(container.RestartPolicy).To(BeNil())
		})
	})

	Context("Resource requirements", func() {
		BeforeEach(func() {
			addS3Bucket("test-bucket", "my-bucket", "/mount/path", "access-key", "secret-key", "", "", "", nil)
		})

		It("applies configured resources to every placement", func() {
			GinkgoT().Setenv(config.EnvVarS3FileSystemResources, `{"requests":{"cpu":"25m"},"limits":{"memory":"64Mi"}}`)

			containers, err := GetS3containerWithError(ctx, sharedFields)

			Expect(err).NotTo(HaveOccurred())
			Expect(containers).To(HaveLen(1))
			Expect(containers[0].Resources).To(Equal(corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("25m")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("64Mi")},
			}))
		})

		It("treats explicit empty resources as empty", func() {
			GinkgoT().Setenv(config.EnvVarS3FileSystemResources, `{}`)

			containers, err := GetS3containerWithError(ctx, sharedFields)

			Expect(err).NotTo(HaveOccurred())
			Expect(containers[0].Resources).To(Equal(corev1.ResourceRequirements{}))
		})

		It("rejects malformed resources", func() {
			GinkgoT().Setenv(config.EnvVarS3FileSystemResources, `{"requests":{"cpu":"invalid"}}`)

			containers, err := GetS3containerWithError(ctx, sharedFields)

			Expect(containers).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring(config.EnvVarS3FileSystemResources)))
		})
	})
})
