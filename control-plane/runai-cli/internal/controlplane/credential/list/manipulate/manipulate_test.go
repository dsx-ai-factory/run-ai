/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseCredentialColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseCredentialColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseCredentialColumns([]string{"name"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseCredentialColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyScope))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseCredentialColumns([]string{"NAME", "Type", "SCOPE"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyScope))
		})

		It("should trim whitespace", func() {
			result, err := ParseCredentialColumns([]string{"  name  ", " type ", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyScope))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{"name", "type", "scope"}
			result, err := ParseCredentialColumns(allColumnNames)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseCredentialColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseCredentialColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseCredentialColumns([]string{"name", "invalid", "scope"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseCredentialColumns([]string{"name", "name", "type"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyType))
		})
	})
})

var _ = Describe("CreateAlignedCredentials", func() {
	var (
		credentials []openapi.CredentialsListResponseEntry
		columns     []CredentialColumn
	)

	BeforeEach(func() {
		cred1 := newTestCredentialEntry("cred1", openapi.ASSETKIND_GENERIC_SECRET, "Cluster")
		cred2 := newTestCredentialEntry("cred2", openapi.ASSETKIND_DOCKER_REGISTRY, "Project")
		credentials = []openapi.CredentialsListResponseEntry{cred1, cred2}
		columns = []CredentialColumn{ColumnKeyName, ColumnKeyType, ColumnKeyScope}
	})

	It("should create aligned credentials with specified columns", func() {
		result := CreateAlignedCredentials(credentials, columns)

		Expect(result.Credentials).To(HaveLen(2))
		Expect(result.Credentials[0].Name).To(Equal("cred1"))
		Expect(result.Credentials[0].Type).To(Equal("Generic secret"))
		Expect(result.Credentials[0].Scope).To(Equal("Cluster"))

		Expect(result.Credentials[1].Name).To(Equal("cred2"))
		Expect(result.Credentials[1].Type).To(Equal("Docker registry"))
		Expect(result.Credentials[1].Scope).To(Equal("Project"))
	})

	It("should only include selected columns", func() {
		columns = []CredentialColumn{ColumnKeyName}
		result := CreateAlignedCredentials(credentials, columns)

		Expect(result.Credentials).To(HaveLen(2))
		Expect(result.Credentials[0].Name).To(Equal("cred1"))
		Expect(result.Credentials[0].Type).To(BeEmpty())
		Expect(result.Credentials[0].Scope).To(BeEmpty())
	})

	It("should handle empty credentials list", func() {
		credentials = []openapi.CredentialsListResponseEntry{}
		result := CreateAlignedCredentials(credentials, columns)

		Expect(result.Credentials).To(HaveLen(0))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedCredentials []CredentialsAlignedRow
		columns            []CredentialColumn
	)

	BeforeEach(func() {
		alignedCredentials = []CredentialsAlignedRow{
			{
				Name:  "cred1",
				Type:  "Generic secret",
				Scope: "Cluster",
			},
			{
				Name:  "cred2",
				Type:  "Docker registry",
				Scope: "Project",
			},
		}
		columns = []CredentialColumn{ColumnKeyName, ColumnKeyType, ColumnKeyScope}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedCredentials, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"cred1", "Generic secret", "Cluster"}))
		Expect(result[1]).To(Equal([]string{"cred2", "Docker registry", "Project"}))
	})

	It("should handle different column order", func() {
		columns = []CredentialColumn{ColumnKeyScope, ColumnKeyName, ColumnKeyType}
		result := CreateTableData(alignedCredentials, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"Cluster", "cred1", "Generic secret"}))
		Expect(result[1]).To(Equal([]string{"Project", "cred2", "Docker registry"}))
	})

	It("should handle single column", func() {
		columns = []CredentialColumn{ColumnKeyName}
		result := CreateTableData(alignedCredentials, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"cred1"}))
		Expect(result[1]).To(Equal([]string{"cred2"}))
	})

	It("should handle empty aligned credentials", func() {
		alignedCredentials = []CredentialsAlignedRow{}
		result := CreateTableData(alignedCredentials, columns)

		Expect(result).To(HaveLen(0))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []CredentialColumn{
			ColumnKeyName, ColumnKeyK8sName, ColumnKeyStatus, ColumnKeyType, ColumnKeyScope,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []CredentialColumn{ColumnKeyName, ColumnKeyType}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]CredentialColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("BriefCredentialFromEntry", func() {
	It("should convert credential entry to brief credential", func() {
		cred := newTestCredentialEntry("test-cred", openapi.ASSETKIND_GENERIC_SECRET, "Project")

		result := BriefCredentialFromEntry(cred)

		Expect(result.Name).To(Equal("test-cred"))
		Expect(result.Type).To(Equal("Generic secret"))
		Expect(result.Scope).To(Equal("Project"))
	})

	It("should handle different credential types", func() {
		testCases := []struct {
			kind         openapi.AssetKind
			expectedType string
		}{
			{openapi.ASSETKIND_GENERIC_SECRET, "Generic secret"},
			{openapi.ASSETKIND_DOCKER_REGISTRY, "Docker registry"},
			{openapi.ASSETKIND_NGC_API_KEY, "NGC API key"},
			{openapi.ASSETKIND_PASSWORD, "Password"},
			{openapi.ASSETKIND_ACCESS_KEY, "Access key"},
		}

		for _, tc := range testCases {
			cred := newTestCredentialEntry("test-cred", tc.kind, "Cluster")
			result := BriefCredentialFromEntry(cred)
			Expect(result.Type).To(Equal(tc.expectedType), "for kind %s", tc.kind)
		}
	})

	It("should include status when available", func() {
		cred := newTestCredentialEntry("test-cred", openapi.ASSETKIND_GENERIC_SECRET, "Project")
		statusInfo := openapi.NewAssetClusterStatusInfo()
		statusInfo.SetStatus(openapi.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND)
		cred.Status.Set(statusInfo)

		result := BriefCredentialFromEntry(cred)
		Expect(result.Status).To(Equal(string(openapi.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND)))
	})
})

func newTestCredentialEntry(name string, kind openapi.AssetKind, scope string) openapi.CredentialsListResponseEntry {
	return openapi.CredentialsListResponseEntry{
		Meta: openapi.AssetMeta{
			Id:    "cred-id-" + name,
			Name:  name,
			Kind:  kind,
			Scope: openapi.Scope(scope),
		},
		Spec: openapi.CredentialsListResponseAssetSpec{},
	}
}
