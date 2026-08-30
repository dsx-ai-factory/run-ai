package e2e_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/router"
)

var _ = Describe("Provisioner Service E2E", func() {
	var (
		server *httptest.Server
	)

	BeforeEach(func() {
		handler := router.Create()
		server = httptest.NewServer(handler)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("POST /environments", func() {
		Context("when request is missing required fields", func() {
			It("should return 422 Unprocessable Entity when domain is missing", func() {
				requestBody := map[string]interface{}{
					"provisioningMode": "full",
					"wildcarddomain":   "*.inference.runai.lab123.example.com",
				}

				body, err := json.Marshal(requestBody)
				Expect(err).ToNot(HaveOccurred())

				resp, err := http.Post(server.URL+"/environments", "application/json", bytes.NewBuffer(body))
				Expect(err).ToNot(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnprocessableEntity))
			})
		})
	})
})
