/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/api/specset/storage/pvc"
	"k8s.io/utils/ptr"
)

var _ = Describe("PVC Parser", func() {
	Context("NewFromDeprecatedOldStyle", func() {
		Context("Existing PVC", func() {
			It("should parse basic existing PVC format", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("runai-claire-brioschi-scratch:/claire-rcp-scratch")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetClaimName()).To(Equal("runai-claire-brioschi-scratch"))
				Expect(result.GetPath()).To(Equal("/claire-rcp-scratch"))
				Expect(result.GetExistingPvc()).To(BeTrue())
				Expect(result.GetReadOnly()).To(BeFalse())
				Expect(result.GetEphemeral()).To(BeFalse())
			})

			It("should parse existing PVC with read-only flag", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("existing-pvc:/mount/path:ro")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetClaimName()).To(Equal("existing-pvc"))
				Expect(result.GetPath()).To(Equal("/mount/path"))
				Expect(result.GetExistingPvc()).To(BeTrue())

				claimInfo := result.GetClaimInfo()
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadOnlyMany.Get(), false)).To(BeTrue())
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadWriteMany.Get(), false)).To(BeFalse())
			})

			It("should parse existing PVC with read-write flag", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("existing-pvc:/mount/path:rw")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetClaimName()).To(Equal("existing-pvc"))
				Expect(result.GetPath()).To(Equal("/mount/path"))
				Expect(result.GetExistingPvc()).To(BeTrue())

				claimInfo := result.GetClaimInfo()
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadOnlyMany.Get(), false)).To(BeFalse())
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadWriteMany.Get(), false)).To(BeTrue())
			})

			It("should fail when PVC name is empty", func() {
				_, err := pvc.NewFromDeprecatedOldStyle(":/mount/path")
				Expect(err).To(MatchError("persistent volume claim name must be specified"))
			})

			It("should fail when mount path is empty", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("existing-pvc:")
				Expect(err).To(MatchError("container mount path must be specified"))
			})
		})

		Context("Dynamic PVC", func() {
			It("should parse basic dynamic PVC format", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("storageclass:5Gi:/path:")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetPath()).To(Equal("/path"))
				Expect(result.GetExistingPvc()).To(BeFalse())

				claimInfo := result.GetClaimInfo()
				Expect(ptr.Deref(claimInfo.StorageClass.Get(), "")).To(Equal("storageclass"))
				Expect(ptr.Deref(claimInfo.Size.Get(), "")).To(Equal("5Gi"))
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadOnlyMany.Get(), false)).To(BeTrue())
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadWriteMany.Get(), false)).To(BeFalse())
			})

			It("should parse dynamic PVC without storage class", func() {
				result, err := pvc.NewFromDeprecatedOldStyle(":5Gi:/path:")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetPath()).To(Equal("/path"))
				Expect(result.GetExistingPvc()).To(BeFalse())

				claimInfo := result.GetClaimInfo()
				Expect(ptr.Deref(claimInfo.StorageClass.Get(), "")).To(BeEmpty())
				Expect(ptr.Deref(claimInfo.Size.Get(), "")).To(Equal("5Gi"))
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadOnlyMany.Get(), false)).To(BeTrue())
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadWriteMany.Get(), false)).To(BeFalse())
			})

			It("should parse dynamic PVC with read-only flag", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("storageclass:5Gi:/path:ro")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetPath()).To(Equal("/path"))
				Expect(result.GetExistingPvc()).To(BeFalse())

				claimInfo := result.GetClaimInfo()
				Expect(ptr.Deref(claimInfo.StorageClass.Get(), "")).To(Equal("storageclass"))
				Expect(ptr.Deref(claimInfo.Size.Get(), "")).To(Equal("5Gi"))
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadOnlyMany.Get(), false)).To(BeTrue())
				Expect(ptr.Deref(claimInfo.AccessModes.Get().ReadWriteMany.Get(), false)).To(BeFalse())
			})

			It("should parse dynamic PVC with mixed units", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("storageclass:1.5Gi:/path:")
				Expect(err).ToNot(HaveOccurred())
				claimInfo := result.GetClaimInfo()
				Expect(ptr.Deref(claimInfo.Size.Get(), "")).To(Equal("1.5Gi"))
			})

			It("should fail when size is empty", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("storageclass::/path:")
				Expect(err).To(MatchError("persistent volume size must be specified"))
			})

			It("should fail when mount path is empty", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("storageclass:5Gi::")
				Expect(err).To(MatchError("container mount path must be specified"))
			})

			It("should fail with invalid storage class name", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("Invalid_Storage_Class:5Gi:/path:")
				Expect(err).To(MatchError(ContainSubstring("A Storage Class name must consist of lower case alphanumeric characters")))
			})

			It("should fail with invalid size format", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("storageclass:5DM:/path:")
				Expect(err).To(MatchError(ContainSubstring("Badly formatted resource request for volume size requierment")))
			})
		})

		Context("Invalid Input", func() {
			It("should fail with too few parameters", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("single-param")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the --pvc parameter can be given in either of the forms"))
			})

			It("should fail with too many parameters", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("param1:param2:param3:param4:param5")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the --pvc parameter can be given in either of the forms"))
			})

			It("should fail with invalid access mode", func() {
				_, err := pvc.NewFromDeprecatedOldStyle("existing-pvc:/path:invalid")
				Expect(err).To(MatchError(ContainSubstring("invalid access permission parameter given")))
			})
		})

		Context("Complex Path Cases", func() {

			It("should handle paths with special characters", func() {
				result, err := pvc.NewFromDeprecatedOldStyle("my-pvc:/path/with space/and-dash")
				Expect(err).ToNot(HaveOccurred())
				Expect(result.GetPath()).To(Equal("/path/with space/and-dash"))
			})
		})
	})
})
