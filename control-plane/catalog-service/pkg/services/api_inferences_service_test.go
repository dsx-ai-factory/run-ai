package services

import (
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/services/huggingface"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var _ = Describe("InferencesAPIService", func() {
	var service *InferencesAPIService

	BeforeEach(func() {
		service = NewInferencesAPIService()
	})

	It("should create service successfully", func() {
		Expect(service).NotTo(BeNil())
	})

	It("should handle empty query (default models)", func() {
		testModels := []huggingface.Model{
			{
				ID:        "llama-2-7b-chat",
				Downloads: 1000000,
				Gated:     nil,
			},
			{
				ID:        "bert-base-uncased",
				Downloads: 2000000,
				Gated:     nil,
			},
		}

		apiResponse := huggingface.ConvertToAPIResponse(testModels)
		response := api.Response(http.StatusOK, apiResponse)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body).NotTo(BeNil())

		searchResults, ok := response.Body.(api.HuggingFaceModelSearchResults)
		Expect(ok).To(BeTrue())
		Expect(searchResults.Resources).NotTo(BeNil())
		Expect(searchResults.Resources).To(HaveLen(2))

		Expect(searchResults.Resources[0].Id).To(Equal("llama-2-7b-chat"))
		Expect(searchResults.Resources[0].Downloads).To(Equal(int32(1000000)))
		Expect(searchResults.Resources[1].Id).To(Equal("bert-base-uncased"))
		Expect(searchResults.Resources[1].Downloads).To(Equal(int32(2000000)))
	})

	It("should handle search query", func() {
		testModels := []huggingface.Model{
			{
				ID:        "gpt2",
				Downloads: 25000000,
				Gated:     nil,
			},
			{
				ID:        "auto-gated-model",
				Downloads: 15000000,
				Gated:     "auto",
			},
			{
				ID:        "open-model",
				Downloads: 10000000,
				Gated:     "false",
			},
			{
				ID:        "manual-gated-model",
				Downloads: 5000000,
				Gated:     "manual",
			},
		}

		apiResponse := huggingface.ConvertToAPIResponse(testModels)
		response := api.Response(http.StatusOK, apiResponse)

		Expect(response.Code).To(Equal(http.StatusOK))
		Expect(response.Body).NotTo(BeNil())

		searchResults, ok := response.Body.(api.HuggingFaceModelSearchResults)
		Expect(ok).To(BeTrue())
		Expect(searchResults.Resources).NotTo(BeNil())
		Expect(searchResults.Resources).To(HaveLen(4))

		Expect(searchResults.Resources[0].Id).To(Equal("gpt2"))
		Expect(searchResults.Resources[0].Downloads).To(Equal(int32(25000000)))
		Expect(searchResults.Resources[0].Gated).To(BeFalse())

		Expect(searchResults.Resources[1].Id).To(Equal("auto-gated-model"))
		Expect(searchResults.Resources[1].Downloads).To(Equal(int32(15000000)))
		Expect(searchResults.Resources[1].Gated).To(BeTrue())

		Expect(searchResults.Resources[2].Id).To(Equal("open-model"))
		Expect(searchResults.Resources[2].Downloads).To(Equal(int32(10000000)))
		Expect(searchResults.Resources[2].Gated).To(BeFalse())

		Expect(searchResults.Resources[3].Id).To(Equal("manual-gated-model"))
		Expect(searchResults.Resources[3].Downloads).To(Equal(int32(5000000)))
		Expect(searchResults.Resources[3].Gated).To(BeTrue())
	})

})
