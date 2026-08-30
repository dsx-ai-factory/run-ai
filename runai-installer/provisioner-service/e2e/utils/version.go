package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaitls "github.com/run-ai/runai/runai-installer/clients/pkg/tls"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
)

// AssertInstalledVersion queries the /_version endpoint on the control plane and
// asserts that the installed version matches config.DefaultRunaiVersion.
// Pass an empty caCertPath to use the system trust store (no custom CA).
func AssertInstalledVersion(domain, caCertPath string) {
	httpClient, err := runaitls.NewHTTPClientWithCA(caCertPath, 30*time.Second)
	Expect(err).ToNot(HaveOccurred(), "failed to create HTTP client for version check")

	versionURL := fmt.Sprintf("https://%s/_version", domain)
	GinkgoWriter.Printf("Checking installed RunAI version at %s\n", versionURL)

	resp, err := httpClient.Get(versionURL)
	Expect(err).ToNot(HaveOccurred(), "failed to reach /_version endpoint at %s", versionURL)
	defer resp.Body.Close()

	Expect(resp.StatusCode).To(Equal(http.StatusOK), "/_version returned unexpected status code")

	body, err := io.ReadAll(resp.Body)
	Expect(err).ToNot(HaveOccurred(), "failed to read /_version response body")

	var versionResponse struct {
		Version string `json:"version"`
	}
	err = json.Unmarshal(body, &versionResponse)
	Expect(err).ToNot(HaveOccurred(), "failed to parse /_version response: %s", string(body))

	GinkgoWriter.Printf("Installed RunAI version: %s\n", versionResponse.Version)
	Expect(versionResponse.Version).To(Equal(config.DefaultRunaiVersion),
		"installed RunAI version does not match the expected default version")
}
