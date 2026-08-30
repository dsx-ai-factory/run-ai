package services

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-installer/provisioner-service/internal/generated/pkg/api"
)

var _ = Describe("CreateEnvironment Validation", func() {
	var (
		adapter *ProvisionServiceAdapter
		ctx     context.Context
	)

	BeforeEach(func() {
		service := NewProvisionerService()
		adapter = NewProvisionServiceAdapter(service)
		ctx = context.Background()
	})

	It("should return 400 when provisioning mode is a number", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "123",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid provisioning mode: 123 must be either 'control-plane-only' or 'full'"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should return 400 when provisioning mode is an invalid string", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "invalid-mode",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid provisioning mode: invalid-mode must be either 'control-plane-only' or 'full'"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should return 400 when domain is invalid", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "invalid_domain",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid domain: invalid_domain is not a valid FQDN"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should return 400 when wildcard domain is invalid", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "invalid.wildcard.com",
			UserEmail:        "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid wildcard domain: invalid.wildcard.com is not a valid wildcard domain (must start with *.)"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should accept valid kubernetesDistribution", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode:       "full",
			Domain:                 "runai.lab123.example.com",
			Wildcarddomain:         "*.inference.runai.lab123.example.com",
			KubernetesDistribution: "openshift",
			UserEmail:              "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
	})

	It("should return 400 when kubernetesDistribution is invalid", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode:       "full",
			Domain:                 "runai.lab123.example.com",
			Wildcarddomain:         "*.inference.runai.lab123.example.com",
			KubernetesDistribution: "invalid-distro",
			UserEmail:              "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid kubernetes distribution: invalid-distro must be one of: [openshift], or empty"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should accept empty kubernetesDistribution", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode:       "full",
			Domain:                 "runai.lab123.example.com",
			Wildcarddomain:         "*.inference.runai.lab123.example.com",
			KubernetesDistribution: "",
			UserEmail:              "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
	})

	It("should accept request without kubernetesDistribution field for full mode", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "researcher@example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
	})

	It("should return 400 when userEmail is missing for full mode", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("userEmail is required when provisioningMode is 'full'"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should return 400 when userEmail is invalid for full mode", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "full",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
			UserEmail:        "invalid-email",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("invalid userEmail: invalid-email is not a valid email address"))
		Expect(response.Code).To(Equal(http.StatusBadRequest))
	})

	It("should accept control-plane-only mode without userEmail", func() {
		req := api.CreateEnvironmentRequest{
			ProvisioningMode: "control-plane-only",
			Domain:           "runai.lab123.example.com",
			Wildcarddomain:   "*.inference.runai.lab123.example.com",
		}

		response, err := adapter.CreateEnvironment(ctx, req)

		Expect(err).ToNot(HaveOccurred())
		Expect(response.Code).To(Equal(http.StatusAccepted))
	})
})
