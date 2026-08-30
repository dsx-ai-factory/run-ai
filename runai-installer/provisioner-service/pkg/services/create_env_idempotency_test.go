package services

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/state"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
)

var _ = Describe("CreateEnvironment Idempotency Behavior", func() {
	var (
		adapter *ProvisionServiceAdapter
		service *ProvisionerService
		ctx     context.Context
	)

	BeforeEach(func() {
		state.Clear()
		service = NewProvisionerService()
		service.provisionFunc = func(envID types.EnvironmentID, mode types.ProvisioningMode, domain string, wildcardDomain string, kubernetesDistribution *string, userEmail *string, projectName *string) {
			state.SetInProgress(envID)
			time.Sleep(2 * time.Second)
			state.SetSucceeded(envID, "Provisioning completed successfully")
		}
		adapter = NewProvisionServiceAdapter(service)
		ctx = context.Background()
	})

	It("should return the same environment ID when creating while one is in progress", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response1, err1 := adapter.CreateEnvironment(ctx, req)
		Expect(err1).ToNot(HaveOccurred())
		Expect(response1.Code).To(Equal(http.StatusAccepted))
		env1 := response1.Body.(api.EnvironmentAccepted)

		response2, err2 := adapter.CreateEnvironment(ctx, req)
		Expect(err2).ToNot(HaveOccurred())
		Expect(response2.Code).To(Equal(http.StatusAccepted))
		env2 := response2.Body.(api.EnvironmentAccepted)

		Expect(env1.Id).To(Equal(env2.Id))
	})

	It("should return different environment IDs when creating after first one completes", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response1, err1 := adapter.CreateEnvironment(ctx, req)
		Expect(err1).ToNot(HaveOccurred())
		Expect(response1.Code).To(Equal(http.StatusAccepted))
		env1 := response1.Body.(api.EnvironmentAccepted)

		Eventually(func() types.EnvironmentPhase {
			statusResponse, err := adapter.GetEnvironment(ctx, env1.Id)
			if err != nil || statusResponse.Code == http.StatusNotFound {
				return types.EnvironmentPhasePending
			}
			status := statusResponse.Body.(api.EnvironmentStatus)
			return types.EnvironmentPhase(status.Phase)
		}, 10*time.Second, 500*time.Millisecond).Should(Equal(types.EnvironmentPhaseSucceeded))

		response2, err2 := adapter.CreateEnvironment(ctx, req)
		Expect(err2).ToNot(HaveOccurred())
		Expect(response2.Code).To(Equal(http.StatusAccepted))
		env2 := response2.Body.(api.EnvironmentAccepted)

		Expect(env1.Id).ToNot(Equal(env2.Id))
	})
})
