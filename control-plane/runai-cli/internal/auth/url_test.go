package auth

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("handleAuthURlResponse", func() {
	var (
		response *http.Response
	)

	Context("when the response status code is not StatusTemporaryRedirect", func() {
		BeforeEach(func() {
			response = &http.Response{
				StatusCode: http.StatusOK,
			}
		})

		It("should return an error", func() {
			_, err := handleAuthURlResponse(response)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("unexpected status code"))
		})
	})

	Context("when the response does not contain the LoginLocationKeyHeader", func() {
		BeforeEach(func() {
			response = &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Header:     http.Header{},
			}
		})

		It("should return an error", func() {
			_, err := handleAuthURlResponse(response)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("no location header in response"))
		})
	})

	Context("when the response contains a valid LoginLocationKeyHeader", func() {
		BeforeEach(func() {
			response = &http.Response{
				StatusCode: http.StatusTemporaryRedirect,
				Header: http.Header{
					LoginLocationKeyHeader: []string{"https://example.com/auth"},
				},
			}
		})

		It("should return the parsed URL", func() {
			url, err := handleAuthURlResponse(response)
			Expect(err).NotTo(HaveOccurred())
			Expect(url.String()).To(Equal("https://example.com/auth"))
		})
	})
})
