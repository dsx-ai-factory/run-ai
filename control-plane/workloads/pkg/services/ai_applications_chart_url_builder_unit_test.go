package services

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func testChartURLBuilder() *chartDownloadURLBuilder {
	return &chartDownloadURLBuilder{baseURL: "https://cp.example.com/"}
}

var _ = Describe("chartDownloadURLBuilder", func() {
	It("builds the content download URL from the base URL and chart id", func() {
		chartID := "3fa85f64-5717-4562-b3fc-2c963f66afa6"
		got, err := testChartURLBuilder().downloadURL(chartID)
		Expect(err).NotTo(HaveOccurred())
		want := fmt.Sprintf("https://cp.example.com/api/v1/ai-applications/charts/%s/content", chartID)
		Expect(got).To(Equal(want))
	})

	It("fails when the base URL is not configured", func() {
		b := testChartURLBuilder()
		b.baseURL = ""
		_, err := b.downloadURL("id")
		Expect(errors.Is(err, errChartBaseURLNotConfigured)).To(BeTrue())
	})
})
