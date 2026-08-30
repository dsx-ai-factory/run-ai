package services

import (
	"bytes"
	"compress/gzip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("chart upload helpers", func() {
	Describe("isGzip", func() {
		var validGzip []byte

		BeforeEach(func() {
			var buf bytes.Buffer
			gw := gzip.NewWriter(&buf)
			_, err := gw.Write([]byte("helm chart bytes"))
			Expect(err).NotTo(HaveOccurred())
			Expect(gw.Close()).To(Succeed())
			validGzip = buf.Bytes()
		})

		It("reports a complete gzip stream as gzip", func() {
			Expect(isGzip(validGzip)).To(BeTrue())
		})

		It("rejects a truncated gzip stream whose trailer is missing", func() {
			Expect(isGzip(validGzip[:len(validGzip)-4])).To(BeFalse())
		})

		DescribeTable("rejects payloads that are not a full gzip stream",
			func(data []byte) {
				Expect(isGzip(data)).To(BeFalse())
			},
			Entry("magic bytes only", []byte{0x1f, 0x8b, 0x08, 0x00}),
			Entry("too short", []byte{0x1f}),
			Entry("zip archive header", []byte("PK\x03\x04")),
			Entry("empty input", []byte{}),
			Entry("plain yaml text", []byte("apiVersion: v2")),
		)
	})

	DescribeTable("originalChartFilename recovers the uploaded name from the temp path",
		func(tempPath, want string) {
			Expect(originalChartFilename(tempPath)).To(Equal(want))
		},
		Entry("strips the random temp suffix", "/tmp/a1.tgz.1849302", "a1.tgz"),
		Entry("keeps dots inside the original name", "/tmp/my.chart.tgz.99", "my.chart.tgz"),
		Entry("falls back when there is no suffix dot", "/tmp/plainname", "plainname"),
	)
})
