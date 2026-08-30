package services

import (
	"encoding/json"

	"github.com/run-ai/runai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/database/models"
	workloadsManagerClient "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("networking submission profiles", func() {
	It("derives generated-native only from the trusted internal origin", func() {
		record := &models.WorkloadsQueueRecord{
			InputRequestPath: InternalWorkloadsAPIPath,
			TranslatedFromV1: true,
		}

		Expect(networkingSubmissionProfileForRecord(record)).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_GENERATED_NATIVE))

		record.TranslatedFromV1 = false
		Expect(networkingSubmissionProfileForRecord(record)).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2))
		record.TranslatedFromV1 = true
		record.InputRequestPath = "/api/v2/workloads"
		Expect(networkingSubmissionProfileForRecord(record)).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2))
	})

	It("derives generated-NIM only from the recorded NIM route", func() {
		record := &models.WorkloadsQueueRecord{
			InputRequestPath: NimServicesAPIPath,
		}

		Expect(networkingSubmissionProfileForRecord(record)).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_GENERATED_NIM))

		record.InputRequestPath = "/api/v2/workloads"
		Expect(networkingSubmissionProfileForRecord(record)).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2))
	})

	It("defaults missing records to direct-v2", func() {
		Expect(networkingSubmissionProfileForRecord(nil)).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2))
	})

	It("adds the derived profile only to active endpoint-networking updates", func() {
		networking := models.NetworkingConfiguration{{}}
		record := &models.WorkloadsQueueRecord{
			ID:               "workload-id",
			Name:             "demo",
			InputRequestPath: "/api/v2/workloads",
			DesiredPhase:     string(api.DESIREDPHASE_RUNNING),
			Configuration:    &models.WorkloadConfiguration{Networking: &networking},
		}

		update := workloadToClusterUpdateWithNetworkingSubmissionProfile(record)
		Expect(update.Configuration.NetworkingSubmissionProfile).To(Equal(api.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2))
		payload, err := json.Marshal(update)
		Expect(err).NotTo(HaveOccurred())
		var agentUpdate workloadsManagerClient.WorkloadV2Update
		Expect(json.Unmarshal(payload, &agentUpdate)).To(Succeed())
		Expect(agentUpdate.Configuration.GetNetworkingSubmissionProfile()).To(Equal(
			workloadsManagerClient.NETWORKINGSUBMISSIONPROFILE_DIRECT_V2,
		))

		emptyNetworking := models.NetworkingConfiguration{}
		record.Configuration.Networking = &emptyNetworking
		update = workloadToClusterUpdateWithNetworkingSubmissionProfile(record)
		Expect(update.Configuration.NetworkingSubmissionProfile).To(BeEmpty())

		record.Configuration.Networking = &networking
		record.DesiredPhase = string(api.DESIREDPHASE_DELETED)
		update = workloadToClusterUpdateWithNetworkingSubmissionProfile(record)
		Expect(update.Configuration.NetworkingSubmissionProfile).To(BeEmpty())
	})
})
