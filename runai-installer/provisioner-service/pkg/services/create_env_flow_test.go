package services

import (
	"context"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	installer_api "github.com/run-ai/runai/runai-installer/installer/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/config"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/state"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/workflow"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/workflow/mocks"
)

func setupMockConfig() {
	runai_config.FlushAll()

	err := runai_config.Register[*config.Server]()
	Expect(err).NotTo(HaveOccurred())

	err = runai_config.Register[*config.TLS]()
	Expect(err).NotTo(HaveOccurred())

	err = runai_config.Register[*config.JFrog]()
	Expect(err).NotTo(HaveOccurred())
}

func cleanupMockConfig() {
}

func waitForPhase(adapter *ProvisionServiceAdapter, ctx context.Context, envID string, expectedPhase types.EnvironmentPhase) {
	Eventually(func() types.EnvironmentPhase {
		statusResponse, err := adapter.GetEnvironment(ctx, envID)
		if err != nil || statusResponse.Code == http.StatusNotFound {
			return types.EnvironmentPhasePending
		}
		status := statusResponse.Body.(api.EnvironmentStatus)
		return types.EnvironmentPhase(status.Phase)
	}, 5*time.Second, 100*time.Millisecond).Should(Equal(expectedPhase))
}

func waitForMessageSubstring(adapter *ProvisionServiceAdapter, ctx context.Context, envID string, substring string) {
	Eventually(func() string {
		statusResponse, err := adapter.GetEnvironment(ctx, envID)
		if err != nil || statusResponse.Code == http.StatusNotFound {
			return ""
		}
		status := statusResponse.Body.(api.EnvironmentStatus)
		return status.Message
	}, 5*time.Second, 50*time.Millisecond).Should(ContainSubstring(substring))
}

var _ = Describe("CreateEnvironment Installation", func() {
	var (
		adapter *ProvisionServiceAdapter
		service *ProvisionerService
		ctx     context.Context
		mockAPI *mocks.MockInstallerAPI
	)

	BeforeEach(func() {
		state.Clear()
		setupMockConfig()
		mockAPI = mocks.NewMockInstallerAPI()
		workflow.SetInstallerAPI(mockAPI)
		service = NewProvisionerService()
		adapter = NewProvisionServiceAdapter(service)
		ctx = context.Background()
	})

	AfterEach(func() {
		workflow.SetInstallerAPI(&workflow.InstallerAPI{})
		cleanupMockConfig()
	})

	It("should successfully install environment with real workflow execution", func() {
		mockAPI.WithSuccessfulInstallation()

		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))

		env := response.Body.(api.EnvironmentAccepted)
		Expect(env.Id).ToNot(BeEmpty())
		Expect(env.Phase).To(Equal(string(types.EnvironmentPhasePending)))
		Expect(env.Mode).To(Equal("full"))

		waitForPhase(adapter, ctx, env.Id, types.EnvironmentPhaseInProgress)
		waitForPhase(adapter, ctx, env.Id, types.EnvironmentPhaseSucceeded)

		statusResponse, err := adapter.GetEnvironment(ctx, env.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(statusResponse.Code).To(Equal(http.StatusOK))

		finalStatus := statusResponse.Body.(api.EnvironmentStatus)
		Expect(finalStatus.Phase).To(Equal(string(types.EnvironmentPhaseSucceeded)))
		Expect(finalStatus.Message).To(Equal("Run:AI installation completed successfully"))
		Expect(finalStatus.UpdatedAt).ToNot(BeZero())
	})

	It("should handle installation failure correctly", func() {
		mockAPI.WithFailedInstallation("control-plane")

		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))

		env := response.Body.(api.EnvironmentAccepted)

		waitForPhase(adapter, ctx, env.Id, types.EnvironmentPhaseFailed)

		statusResponse, err := adapter.GetEnvironment(ctx, env.Id)
		Expect(err).ToNot(HaveOccurred())
		Expect(statusResponse.Code).To(Equal(http.StatusOK))

		finalStatus := statusResponse.Body.(api.EnvironmentStatus)
		Expect(finalStatus.Phase).To(Equal(string(types.EnvironmentPhaseFailed)))
		Expect(finalStatus.Message).To(ContainSubstring("Run:AI installation failed"))
	})

	It("should track installation progress through state messages", func() {
		mockAPI.WithEvents([]mocks.MockEvent{
			{FlowName: "prerequisites", FlowStatus: installer_api.Installing, Delay: 100 * time.Millisecond},
			{FlowName: "prerequisites", FlowStatus: installer_api.Installed, Delay: 100 * time.Millisecond},
			{FlowName: "control-plane", FlowStatus: installer_api.Installing, Delay: 100 * time.Millisecond},
			{FlowName: "control-plane", FlowStatus: installer_api.Installed, Delay: 100 * time.Millisecond},
		})

		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)
		Expect(err).ToNot(HaveOccurred())

		env := response.Body.(api.EnvironmentAccepted)

		waitForMessageSubstring(adapter, ctx, env.Id, "Installing")
		waitForPhase(adapter, ctx, env.Id, types.EnvironmentPhaseSucceeded)
	})
})
