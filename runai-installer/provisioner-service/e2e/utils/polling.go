package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
)

func CreateEnvironment(serverURL string, requestBody map[string]interface{}) api.EnvironmentAccepted {
	body, err := json.Marshal(requestBody)
	Expect(err).ToNot(HaveOccurred())

	resp, err := http.Post(serverURL+"/environments", "application/json", bytes.NewBuffer(body))
	Expect(err).ToNot(HaveOccurred())
	defer resp.Body.Close()

	Expect(resp.StatusCode).To(Equal(http.StatusAccepted))

	var accepted api.EnvironmentAccepted
	err = json.NewDecoder(resp.Body).Decode(&accepted)
	Expect(err).ToNot(HaveOccurred())
	Expect(accepted.Id).ToNot(BeEmpty())
	Expect(accepted.Phase).To(Equal("pending"))

	GinkgoWriter.Printf("Created environment with ID: %s\n", accepted.Id)
	return accepted
}

func pollEnvironmentStatus(serverURL, envID string) (api.EnvironmentStatus, error) {
	var status api.EnvironmentStatus
	pollResp, err := http.Get(serverURL + "/environments/" + envID)
	if err != nil {
		GinkgoWriter.Printf("Poll error: %v\n", err)
		return status, err
	}
	defer pollResp.Body.Close()

	if pollResp.StatusCode != http.StatusOK {
		GinkgoWriter.Printf("Poll returned HTTP %d\n", pollResp.StatusCode)
		return status, fmt.Errorf("HTTP %d", pollResp.StatusCode)
	}

	err = json.NewDecoder(pollResp.Body).Decode(&status)
	if err != nil {
		GinkgoWriter.Printf("Decode error: %v\n", err)
		return status, err
	}
	GinkgoWriter.Printf("Status | Phase: %s | Mode: %s | Message: %s | Updated: %s\n",
		status.Phase,
		status.Mode,
		status.Message,
		status.UpdatedAt,
	)
	return status, nil
}

func WaitForEnvironmentCompletion(serverURL, envID string, timeout, interval time.Duration, customMessage ...string) api.EnvironmentStatus {
	var finalStatus api.EnvironmentStatus
	Eventually(func() string {
		status, err := pollEnvironmentStatus(serverURL, envID)
		if err != nil {
			return "error"
		}
		finalStatus = status
		return status.Phase
	}, timeout, interval).Should(Equal("succeeded"), func() string {
		msg := fmt.Sprintf("Installation failed. Final status: phase=%s, mode=%s, message=%s",
			finalStatus.Phase, finalStatus.Mode, finalStatus.Message)
		if len(customMessage) > 0 {
			msg += "\n" + customMessage[0]
		}
		return msg
	})
	return finalStatus
}

func WaitForEnvironmentFailure(serverURL, envID string, timeout, interval time.Duration, customMessage ...string) api.EnvironmentStatus {
	var finalStatus api.EnvironmentStatus
	Eventually(func() string {
		status, err := pollEnvironmentStatus(serverURL, envID)
		if err != nil {
			return "error"
		}
		finalStatus = status
		return status.Phase
	}, timeout, interval).Should(Equal("failed"), func() string {
		msg := fmt.Sprintf("Installation did not fail as expected. Final status: phase=%s, mode=%s, message=%s",
			finalStatus.Phase, finalStatus.Mode, finalStatus.Message)
		if len(customMessage) > 0 {
			msg += "\n" + customMessage[0]
		}
		return msg
	})
	return finalStatus
}
