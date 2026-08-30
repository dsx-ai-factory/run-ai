package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/mycredential/client"
)

func newTestEntry(name, description string) client.MyCredentialEntry {
	return client.MyCredentialEntry{Type: "genericSecret", Name: name, Description: description}
}

func newTestEntryWithType(credType, name, description string) client.MyCredentialEntry {
	return client.MyCredentialEntry{Type: credType, Name: name, Description: description}
}

var _ = Describe("CreateAlignedMyCredentials", func() {
	var (
		entries []client.MyCredentialEntry
		columns []MyCredentialColumn
	)

	BeforeEach(func() {
		entries = []client.MyCredentialEntry{
			newTestEntryWithType("genericSecret", "cred-1", "first credential"),
			newTestEntryWithType("dockerRegistry", "cred-2", "second credential"),
			newTestEntryWithType("genericSecret", "cred-3", ""),
		}
		columns = []MyCredentialColumn{ColumnKeyType, ColumnKeyName, ColumnKeyDescription}
	})

	It("should extract all fields for all columns", func() {
		entriesWithType := []client.MyCredentialEntry{
			{Type: "dockerRegistry", Name: "cred-1", Description: "first credential"},
			{Type: "genericSecret", Name: "cred-2", Description: "second credential"},
			{Type: "dockerRegistry", Name: "cred-3", Description: ""},
		}
		allColumns := []MyCredentialColumn{ColumnKeyType, ColumnKeyName, ColumnKeyDescription}
		result := CreateAlignedMyCredentials(entriesWithType, allColumns)

		Expect(result.Credentials).To(HaveLen(3))
		Expect(result.Credentials[0]).To(Equal(MyCredentialAlignedRow{
			Type: "dockerRegistry", Name: "cred-1", Description: "first credential",
		}))
		Expect(result.Credentials[1]).To(Equal(MyCredentialAlignedRow{
			Type: "genericSecret", Name: "cred-2", Description: "second credential",
		}))
		Expect(result.Credentials[2]).To(Equal(MyCredentialAlignedRow{
			Type: "dockerRegistry", Name: "cred-3", Description: "",
		}))
	})

	It("should only extract name when only name column is requested", func() {
		result := CreateAlignedMyCredentials(entries, []MyCredentialColumn{ColumnKeyName})

		Expect(result.Credentials).To(HaveLen(3))
		for _, row := range result.Credentials {
			Expect(row.Name).ToNot(BeEmpty())
			Expect(row.Description).To(BeEmpty())
		}
	})

	It("should only extract description when only description column is requested", func() {
		entries := []client.MyCredentialEntry{
			newTestEntryWithType("genericSecret", "cred-1", "has description"),
		}
		result := CreateAlignedMyCredentials(entries, []MyCredentialColumn{ColumnKeyDescription})

		Expect(result.Credentials).To(HaveLen(1))
		Expect(result.Credentials[0].Name).To(BeEmpty())
		Expect(result.Credentials[0].Description).To(Equal("has description"))
	})

	It("should handle empty entries", func() {
		result := CreateAlignedMyCredentials([]client.MyCredentialEntry{}, columns)
		Expect(result.Credentials).To(HaveLen(0))
	})

	It("should handle nil entries", func() {
		result := CreateAlignedMyCredentials(nil, columns)
		Expect(result.Credentials).To(HaveLen(0))
	})

	It("should handle entries with special characters", func() {
		entries := []client.MyCredentialEntry{
			newTestEntryWithType("genericSecret", "my-secret/v1", `description with "quotes" & <special> chars`),
		}
		result := CreateAlignedMyCredentials(entries, columns)

		Expect(result.Credentials).To(HaveLen(1))
		Expect(result.Credentials[0].Type).To(Equal("genericSecret"))
		Expect(result.Credentials[0].Name).To(Equal("my-secret/v1"))
		Expect(result.Credentials[0].Description).To(Equal(`description with "quotes" & <special> chars`))
	})

	Context("end-to-end pipeline with CreateTableData", func() {
		It("should produce correct table data through the full pipeline", func() {
			aligned := CreateAlignedMyCredentials(entries, columns)
			tableData := CreateTableData(aligned.Credentials, columns)

			Expect(tableData).To(HaveLen(3))
			Expect(tableData[0]).To(Equal([]string{"genericSecret", "cred-1", "first credential"}))
			Expect(tableData[1]).To(Equal([]string{"dockerRegistry", "cred-2", "second credential"}))
			Expect(tableData[2]).To(Equal([]string{"genericSecret", "cred-3", ""}))
		})

		It("should produce single-column table through name-only pipeline", func() {
			nameOnly := []MyCredentialColumn{ColumnKeyName}
			aligned := CreateAlignedMyCredentials(entries, nameOnly)
			tableData := CreateTableData(aligned.Credentials, nameOnly)

			Expect(tableData).To(HaveLen(3))
			for i, row := range tableData {
				Expect(row).To(HaveLen(1))
				Expect(row[0]).To(Equal(entries[i].Name))
			}
		})
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		rows    []MyCredentialAlignedRow
		columns []MyCredentialColumn
	)

	BeforeEach(func() {
		rows = []MyCredentialAlignedRow{
			{Type: "dockerRegistry", Name: "cred-1", Description: "first credential"},
			{Type: "genericSecret", Name: "cred-2", Description: "second credential"},
		}
		columns = []MyCredentialColumn{ColumnKeyType, ColumnKeyName, ColumnKeyDescription}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"dockerRegistry", "cred-1", "first credential"}))
		Expect(result[1]).To(Equal([]string{"genericSecret", "cred-2", "second credential"}))
	})

	It("should respect column order - description before name", func() {
		columns = []MyCredentialColumn{ColumnKeyDescription, ColumnKeyName, ColumnKeyType}
		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"first credential", "cred-1", "dockerRegistry"}))
		Expect(result[1]).To(Equal([]string{"second credential", "cred-2", "genericSecret"}))
	})

	It("should handle single column", func() {
		result := CreateTableData(rows, []MyCredentialColumn{ColumnKeyName})

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"cred-1"}))
		Expect(result[1]).To(Equal([]string{"cred-2"}))
	})

	It("should handle empty rows", func() {
		result := CreateTableData([]MyCredentialAlignedRow{}, columns)
		Expect(result).To(HaveLen(0))
	})

	It("should handle nil rows", func() {
		result := CreateTableData(nil, columns)
		Expect(result).To(HaveLen(0))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return type, name and description columns in expected order", func() {
		result := GetDefaultColumns()

		expectedColumns := []MyCredentialColumn{
			ColumnKeyType,
			ColumnKeyName,
			ColumnKeyDescription,
		}
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return exactly three columns: type, name and description", func() {
		result := GetDefaultColumns()

		Expect(result).To(HaveLen(3))
		Expect(result).To(ContainElement(ColumnKeyType))
		Expect(result).To(ContainElement(ColumnKeyName))
		Expect(result).To(ContainElement(ColumnKeyDescription))
	})
})
