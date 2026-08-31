/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	listCommon "github.com/run-ai/runai/control-plane/runai-cli/cmd/common/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/credential/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/credential/list/manipulate"
	openapi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
)

var _ = Describe("Credential List Display", func() {
	var (
		output      *list.Output
		credentials []openapi.CredentialsListResponseEntry
	)

	BeforeEach(func() {
		cred1 := newTestCredential("aws-access-key", openapi.ASSETKIND_GENERIC_SECRET, "Project")
		cred2 := newTestCredential("docker-hub-creds", openapi.ASSETKIND_DOCKER_REGISTRY, "Cluster")
		cred3 := newTestCredential("ngc-api-prod", openapi.ASSETKIND_NGC_API_KEY, "Department")
		credentials = []openapi.CredentialsListResponseEntry{cred1, cred2, cred3}

		response := &openapi.CredentialsListResponse{
			Entries: credentials,
		}
		var err error
		output, err = list.ParseListResponse(response)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("ParseListResponse", func() {
		It("should parse response successfully", func() {
			Expect(output).ToNot(BeNil())
		})

		It("should return error for nil response", func() {
			output, err := list.ParseListResponse(nil)
			Expect(err).To(HaveOccurred())
			Expect(output).To(BeNil())
		})
	})

	Context("SetRaw", func() {
		It("should set raw flag", func() {
			output.SetRaw(true)
		})
	})

	Context("SetColumns", func() {
		It("should set valid columns", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for invalid columns", func() {
			err := output.SetColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Subject", func() {
		It("should return correct subject", func() {
			subject := output.Subject()
			Expect(subject).To(Equal("credential"))
		})
	})

	Context("Columns", func() {
		It("should return table columns", func() {
			err := output.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			columns := output.Columns()
			Expect(columns).To(HaveLen(2))
			Expect(columns[0].Title).To(Equal("Name"))
			Expect(columns[1].Title).To(Equal("Type"))
		})
	})

	Context("Data", func() {
		It("should return table data with correct type formatting", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(data).To(HaveLen(3))

			Expect(data[0][0]).To(Equal("aws-access-key"))
			Expect(data[0][1]).To(Equal("Generic secret"))
			Expect(data[0][2]).To(Equal("Project"))

			Expect(data[1][0]).To(Equal("docker-hub-creds"))
			Expect(data[1][1]).To(Equal("Docker registry"))
			Expect(data[1][2]).To(Equal("Cluster"))

			Expect(data[2][0]).To(Equal("ngc-api-prod"))
			Expect(data[2][1]).To(Equal("NGC API key"))
			Expect(data[2][2]).To(Equal("Department"))
		})
	})

	Context("GenerateTable", func() {
		It("should generate table output", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())

			table, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(table).To(ContainSubstring("aws-access-key"))
			Expect(table).To(ContainSubstring("docker-hub-creds"))
			Expect(table).To(ContainSubstring("ngc-api-prod"))
			Expect(table).To(ContainSubstring("Generic secret"))
			Expect(table).To(ContainSubstring("Docker registry"))
			Expect(table).To(ContainSubstring("NGC API key"))
		})

		It("should generate table with empty credentials", func() {
			emptyResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{},
			}
			emptyOutput, err := list.ParseListResponse(emptyResponse)
			Expect(err).ToNot(HaveOccurred())

			err = emptyOutput.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			table, err := emptyOutput.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(table).ToNot(BeEmpty())
		})
	})

	Context("GenerateJson", func() {
		It("should generate JSON output", func() {
			err := output.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring("aws-access-key"))
			Expect(json).To(ContainSubstring("docker-hub-creds"))
			Expect(json).To(ContainSubstring("ngc-api-prod"))
		})

		It("should generate raw JSON output", func() {
			output.SetRaw(true)
			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring("aws-access-key"))
		})

		It("should have correct JSON structure with brief format", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())

			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring(`"credentials"`))
			Expect(json).To(ContainSubstring(`"name"`))
			Expect(json).To(ContainSubstring(`"type"`))
			Expect(json).To(ContainSubstring(`"scope"`))
		})
	})

	Context("GenerateYaml", func() {
		It("should generate YAML output", func() {
			err := output.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("aws-access-key"))
			Expect(yaml).To(ContainSubstring("docker-hub-creds"))
		})

		It("should generate raw YAML output", func() {
			output.SetRaw(true)
			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("aws-access-key"))
		})

		It("should have correct YAML structure with brief format", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())

			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("credentials:"))
			Expect(yaml).To(ContainSubstring("name:"))
			Expect(yaml).To(ContainSubstring("type:"))
			Expect(yaml).To(ContainSubstring("scope:"))
		})
	})

	Context("HandleColumnsFlags", func() {
		It("should return nil when set-default-columns is false", func() {
			mockOutput := &listCommon.Output{}
			err := list.HandleColumnsFlags(*mockOutput, nil, "test")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Default Columns", func() {
	It("should have correct default columns", func() {
		defaults := manipulate.GetDefaultColumns()
		Expect(defaults).To(HaveLen(5))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyName))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyType))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyScope))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyK8sName))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyStatus))
	})
})
