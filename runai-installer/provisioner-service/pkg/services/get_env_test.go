package services

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/state"
	"github.com/run-ai/runai/runai-installer/provisioner-service/pkg/types"
)

var _ = Describe("GetEnvironment", func() {
	var (
		adapter *ProvisionServiceAdapter
		ctx     context.Context
	)

	BeforeEach(func() {
		state.Clear()
		service := NewProvisionerService()
		adapter = NewProvisionServiceAdapter(service)
		ctx = context.Background()
	})

	It("should return 404 when environment does not exist", func() {
		response, err := adapter.GetEnvironment(ctx, "non-existent-id")

		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusNotFound))
	})

	It("should return environment and update when state changes", func() {
		envID := types.EnvironmentID("test-env-123")
		mode := types.ProvisioningModeFull

		state.CreateEnvironmentStatus(envID, mode)

		response1, err1 := adapter.GetEnvironment(ctx, string(envID))
		Expect(err1).ToNot(HaveOccurred())
		Expect(response1.Code).To(Equal(http.StatusOK))

		status1 := response1.Body.(api.EnvironmentStatus)
		Expect(status1.Id).To(Equal(string(envID)))
		Expect(status1.Phase).To(Equal(string(types.EnvironmentPhasePending)))
		Expect(status1.Message).To(Equal("Queued"))
		Expect(status1.Mode).To(Equal(string(mode)))

		state.SetSucceeded(envID, "Provisioning completed successfully")

		response2, err2 := adapter.GetEnvironment(ctx, string(envID))
		Expect(err2).ToNot(HaveOccurred())
		Expect(response2.Code).To(Equal(http.StatusOK))

		status2 := response2.Body.(api.EnvironmentStatus)
		Expect(status2.Id).To(Equal(string(envID)))
		Expect(status2.Phase).To(Equal(string(types.EnvironmentPhaseSucceeded)))
		Expect(status2.Message).To(Equal("Provisioning completed successfully"))
		Expect(status2.Mode).To(Equal(string(mode)))
		Expect(status2.UpdatedAt).To(BeTemporally(">", status1.UpdatedAt))
	})
})
