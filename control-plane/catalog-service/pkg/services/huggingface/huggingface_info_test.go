/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package huggingface

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHuggingface(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Huggingface Suite")
}

var _ = Describe("HuggingFace Core Logic", func() {
	Describe("DeduplicateAndCombine", func() {
		It("should prioritize trending models over downloaded duplicates", func() {
			trending := []Model{
				{ID: "model1", Downloads: 150},
				{ID: "model2", Downloads: 200},
			}
			downloaded := []Model{
				{ID: "model1", Downloads: 150}, // duplicate, should be ignored
				{ID: "model3", Downloads: 300},
			}
			expected := []Model{
				{ID: "model1", Downloads: 150}, // trending version kept
				{ID: "model2", Downloads: 200},
				{ID: "model3", Downloads: 300},
			}

			result := DeduplicateAndCombine(trending, downloaded)

			Expect(result).To(Equal(expected))
		})

		It("should limit to 100 models maximum", func() {
			trending := make([]Model, 10)
			downloaded := make([]Model, 95)

			// Populate with unique IDs
			for i := 0; i < 10; i++ {
				trending[i] = Model{ID: "trending" + string(rune('0'+i)), Downloads: int32(i)}
			}
			for i := 0; i < 95; i++ {
				downloaded[i] = Model{ID: "downloaded" + string(rune('0'+i)), Downloads: int32(i)}
			}

			result := DeduplicateAndCombine(trending, downloaded)

			Expect(len(result)).To(Equal(100))
			// First 10 should be trending models
			for i := 0; i < 10; i++ {
				Expect(result[i]).To(Equal(trending[i]))
			}
		})
	})

	Describe("ConvertToAPIResponse", func() {
		It("should handle gated field correctly with production API format", func() {
			models := []Model{
				{ID: "model1", Downloads: 100, Gated: "false"},  // false string
				{ID: "model2", Downloads: 200, Gated: "manual"}, // manual gating
				{ID: "model3", Downloads: 300, Gated: "auto"},   // auto gating
				{ID: "model4", Downloads: 400, Gated: ""},       // empty string should be false
				{ID: "model5", Downloads: 500, Gated: nil},      // nil should be false
			}

			result := ConvertToAPIResponse(models)

			Expect(result.Resources).To(HaveLen(5))
			Expect(result.Resources[0].Gated).To(BeFalse(), "false string should be treated as not gated")
			Expect(result.Resources[1].Gated).To(BeTrue(), "manual string should be treated as gated")
			Expect(result.Resources[2].Gated).To(BeTrue(), "auto string should be treated as gated")
			Expect(result.Resources[3].Gated).To(BeFalse(), "empty string should be treated as not gated")
			Expect(result.Resources[4].Gated).To(BeFalse(), "nil should be treated as not gated")
		})

		It("should preserve trending status in API response", func() {
			models := []Model{
				{ID: "trending-model", Downloads: 100, Gated: nil, Trending: true},
				{ID: "regular-model", Downloads: 200, Gated: nil, Trending: false},
			}

			result := ConvertToAPIResponse(models)

			Expect(result.Resources).To(HaveLen(2))
			Expect(result.Resources[0].Trending).To(BeTrue(), "trending model should have Trending=true")
			Expect(result.Resources[1].Trending).To(BeFalse(), "non-trending model should have Trending=false")
		})
	})

	Describe("SearchModels Integration", func() {
		var (
			service ModelSearchInterface
			server  *httptest.Server
			ctx     context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("should NOT use filter (model tag) parameter for search queries", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				// Verify SearchModels does NOT use filter parameter
				Expect(query.Get(PipelineTagParam)).To(BeEmpty(), "SearchModels should not use pipeline_tag parameter") // when searching for specific model we do not filter by model tag (like text-generation)

				// Verify it uses the correct parameters
				Expect(query.Get(ParamSearch)).To(Equal("test-query"))
				Expect(query.Get(ParamLimit)).To(Equal(LimitDownload))
				Expect(query.Get(ParamSort)).To(Equal(SortDownloads))
				Expect(query.Get(ParamFull)).To(Equal(FullTrue))

				models := []Model{
					{ID: "search-result-1", Downloads: 100},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(models)
			}))

			service = NewModelSearchServiceWithConfig(server.Client(), server.URL)
			result, apiErr := service.SearchModels(ctx, "test-query")

			Expect(apiErr).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ID).To(Equal("search-result-1"))
		})
	})

	Describe("FetchDefaultModels Integration", func() {
		var (
			service ModelSearchInterface
			server  *httptest.Server
			ctx     context.Context
		)

		BeforeEach(func() {
			ctx = context.Background()
		})

		AfterEach(func() {
			if server != nil {
				server.Close()
			}
		})

		It("should use pipeline_tag parameter to filter for text-generation models", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				if query.Get(ParamSort) == SortTrendingScore {
					// Verify trending request uses pipeline_tag parameter to filter for text-generation models
					Expect(query.Get(PipelineTagParam)).To(Equal(TextGeneration), "FetchDefaultModels should use pipeline_tag parameter for trending models")
					Expect(query.Get(ParamLimit)).To(Equal(LimitTrending))

					models := []Model{
						{ID: "trending-1", Downloads: 50},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models)
				} else if query.Get(ParamSort) == SortDownloads {
					// Verify downloaded models request uses pipeline_tag parameter to filter for text-generation models
					Expect(query.Get(PipelineTagParam)).To(Equal(TextGeneration), "FetchDefaultModels should use pipeline_tag parameter for downloaded models")
					Expect(query.Get(ParamLimit)).To(Equal(LimitDownload))

					models := []Model{
						{ID: "downloaded-1", Downloads: 1000},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models)
				}
			}))

			service = NewModelSearchServiceWithConfig(server.Client(), server.URL)
			result, apiErr := service.FetchDefaultModels(ctx)

			Expect(apiErr).To(BeNil())
			Expect(result).To(HaveLen(2))
		})

		It("should mark trending models with Trending=true", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				if query.Get(ParamSort) == SortTrendingScore {
					models := []Model{
						{ID: "trending-model", Downloads: 50},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models)
				} else if query.Get(ParamSort) == SortDownloads {
					models := []Model{
						{ID: "downloaded-model", Downloads: 1000},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models)
				}
			}))

			service = NewModelSearchServiceWithConfig(server.Client(), server.URL)
			result, apiErr := service.FetchDefaultModels(ctx)

			Expect(apiErr).To(BeNil())
			Expect(result).To(HaveLen(2))

			// Trending model should be marked as trending
			Expect(result[0].ID).To(Equal("trending-model"))
			Expect(result[0].Trending).To(BeTrue(), "model from trending API should have Trending=true")

			// Downloaded model should NOT be marked as trending
			Expect(result[1].ID).To(Equal("downloaded-model"))
			Expect(result[1].Trending).To(BeFalse(), "model from downloads API should have Trending=false")
		})

		It("should fallback to downloaded models when trending fails", func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				query := r.URL.Query()

				if query.Get(ParamSort) == SortTrendingScore {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else if query.Get(ParamSort) == SortDownloads {
					models := []Model{
						{ID: "downloaded-1", Downloads: 1000},
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(models)
				}
			}))

			service = NewModelSearchServiceWithConfig(server.Client(), server.URL)
			result, apiErr := service.FetchDefaultModels(ctx)

			Expect(apiErr).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0].ID).To(Equal("downloaded-1"))
		})
	})
})
