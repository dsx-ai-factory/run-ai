/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var _ = Describe("Environment List Model Tests", func() {
	Describe("ParseListResponse", func() {
		It("should handle empty environment list", func() {
			response := &openapi.EnvironmentListResponse{Entries: []openapi.EnvironmentAsset{}}
			output, err := ParseListResponse(response)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
			Expect(output.GetList()).To(HaveLen(0))
		})

		It("should handle environments with data", func() {
			env := openapi.EnvironmentAsset{
				Meta: openapi.AssetMeta{
					Name:  "test-env",
					Scope: "Project",
				},
				Spec: openapi.EnvironmentAssetSpec{},
			}
			env.Spec.SetImage("ubuntu:latest")

			response := &openapi.EnvironmentListResponse{Entries: []openapi.EnvironmentAsset{env}}
			output, err := ParseListResponse(response)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())
			Expect(output.GetList()).To(HaveLen(1))
			Expect(output.GetList()[0].Name).To(Equal("test-env"))
			Expect(output.GetList()[0].Scope).To(Equal("Project"))
			Expect(output.GetList()[0].Image).To(Equal("ubuntu:latest"))
		})
	})
})
