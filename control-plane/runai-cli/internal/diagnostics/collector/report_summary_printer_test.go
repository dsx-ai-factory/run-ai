package collector

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CollectionReporter", func() {
	Context("Subject", func() {
		It("no reports, should return only current", func() {
			reportByNamespaces := map[string]*NamespaceReport{}
			rsp := newReportSummaryPrinter(reportByNamespaces, true)
			Expect("Containers [current]:").To(Equal(rsp.Subject()))
		})
		It("should return only current", func() {
			reportByNamespaces := map[string]*NamespaceReport{"namespace": {ContainerReports: []*ContainerReport{{}}}}
			rsp := newReportSummaryPrinter(reportByNamespaces, true)
			Expect("Containers [current]:").To(Equal(rsp.Subject()))
		})
		It("should return only current", func() {
			reportByNamespaces := map[string]*NamespaceReport{"namespace-a": {ContainerReports: []*ContainerReport{{}}}, "namespace-b": {ContainerReports: []*ContainerReport{{Previous: true}}}}
			rsp := newReportSummaryPrinter(reportByNamespaces, false)
			Expect("Containers [current, previous]:").To(Equal(rsp.Subject()))
		})
	})

	Context("Data", func() {
		It("no reports, should return empty data", func() {
			reportByNamespaces := map[string]*NamespaceReport{}
			rsp := newReportSummaryPrinter(reportByNamespaces, false)
			Expect(rsp.Data()).To(BeEmpty())
		})

		It("with reports, data should not be empty", func() {
			reportByNamespaces := map[string]*NamespaceReport{
				"namespace-a": {
					SuccessCount: 2,
					FailCount:    1,
				},
				"namespace-b": {
					SuccessCount: 1,
					FailCount:    5,
				},
				"namespace-c": {
					SuccessCount: 3,
					FailCount:    0,
				},
			}
			rsp := newReportSummaryPrinter(reportByNamespaces, false)
			data := rsp.Data()
			Expect(data).To(Equal([][]string{{"namespace-a", "3", "2", "1"}, {"namespace-b", "6", "1", "5"}, {"namespace-c", "3", "3", "0"}}))
		})
	})

	Context("summary table", func() {
		It("should be not empty formated table", func() {
			reportByNamespaces := map[string]*NamespaceReport{
				"namespace-a": {
					SuccessCount: 2,
					FailCount:    1,
				},
				"namespace-b": {
					SuccessCount: 1,
					FailCount:    5,
				},
			}
			rsp := newReportSummaryPrinter(reportByNamespaces, false)
			table := rsp.SummaryTable()
			expected :=
				` Namespace       Total       Succeeded   Failed     
────────────────────────────────────────────────────
 namespace-a     3           2           1          
 namespace-b     6           1           5          `
			Expect(table).To(Equal(expected))
		})
	})
})
