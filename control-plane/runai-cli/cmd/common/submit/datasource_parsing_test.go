/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseDatasourceValue", func() {
	Context("Valid formats", func() {
		It("should parse type=TYPE,name=NAME format", func() {
			name, dsType, err := parseDatasourceValue("type=pvc,name=my-datasource")
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("my-datasource"))
			Expect(dsType).To(Equal("pvc"))
		})

		It("should parse name=NAME,type=TYPE format", func() {
			name, dsType, err := parseDatasourceValue("name=my-datasource,type=s3")
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("my-datasource"))
			Expect(dsType).To(Equal("s3"))
		})

		It("should handle different datasource types", func() {
			types := []string{"pvc", "nfs", "s3", "git", "hostPath", "configMap", "secretVolume"}
			for _, t := range types {
				name, dsType, err := parseDatasourceValue("type=" + t + ",name=test-ds")
				Expect(err).ToNot(HaveOccurred())
				Expect(name).To(Equal("test-ds"))
				Expect(dsType).To(Equal(t))
			}
		})

		It("should handle names with special characters", func() {
			name, dsType, err := parseDatasourceValue("type=pvc,name=my-datasource-123")
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("my-datasource-123"))
			Expect(dsType).To(Equal("pvc"))
		})

		It("should trim whitespace from values", func() {
			name, dsType, err := parseDatasourceValue("type= pvc ,name= my-datasource ")
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("my-datasource"))
			Expect(dsType).To(Equal("pvc"))
		})
	})

	Context("Invalid formats", func() {
		It("should fail when type is missing", func() {
			_, _, err := parseDatasourceValue("name=my-datasource")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datasource format"))
			Expect(err.Error()).To(ContainSubstring("required format is type=TYPE,name=NAME"))
		})

		It("should fail when name is missing", func() {
			_, _, err := parseDatasourceValue("type=pvc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datasource format"))
		})

		It("should fail when just a plain name is provided", func() {
			_, _, err := parseDatasourceValue("my-datasource")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datasource format"))
		})

		It("should fail when empty string is provided", func() {
			_, _, err := parseDatasourceValue("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datasource format"))
		})

		It("should fail when type value is empty", func() {
			_, _, err := parseDatasourceValue("type=,name=my-datasource")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datasource format"))
		})

		It("should fail when name value is empty", func() {
			_, _, err := parseDatasourceValue("type=pvc,name=")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid datasource format"))
		})
	})
})
