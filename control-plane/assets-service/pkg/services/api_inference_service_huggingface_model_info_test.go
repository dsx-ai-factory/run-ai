/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/go-chi/chi/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// the detailed tests for the 'get huggingface model info' api can be found in the huggingface package.
// here we're only testing the interface to inference service:
// * basic valid call
// * basic invalid call
// * model name format validations (as it's done in the swagger but we still want to test it)
// * token format validations (as it's done in the swagger but we still want to test it)
// * too many concurrent requests (as it's implemented in the inference service level)

func createHuggingFaceModelInfoRequest(modelName, token string) api.HuggingFaceModelInfoRequest {
	req := &api.HuggingFaceModelInfoRequest{
		ModelName: modelName,
	}

	if token != "" {
		req.HfToken = &token
	}

	return *req
}

func callHuggingFaceModelInfoApi(randoms *testutils.RandomValues, router *chi.Router, modelName string, hfToken string) *httptest.ResponseRecorder {
	reqPayload := createHuggingFaceModelInfoRequest(modelName, hfToken)
	createHttpReq, err := http.NewRequest("POST", urlModelHuggingFaceInfo, makePayload(reqPayload))
	Expect(err).To(BeNil())
	createHttpReq.Header = *headerForAdmin(randoms)

	createHttpResp := httptest.NewRecorder()
	(*router).ServeHTTP(createHttpResp, createHttpReq)

	return createHttpResp
}

func testValidHfToken(randoms *testutils.RandomValues, router *chi.Router, hfToken string) {
	resp := callHuggingFaceModelInfoApi(randoms, router, "valid-model", hfToken)
	Expect(resp.Code).To(Equal(http.StatusOK))
}

func testInvalidModelName(randoms *testutils.RandomValues, router *chi.Router, modelName string) {
	resp := callHuggingFaceModelInfoApi(randoms, router, modelName, "")
	Expect(resp.Code).To(Equal(http.StatusBadRequest))
	Expect(resp.Body.String()).To(ContainSubstring("value for ModelName must comply to the pattern"))
}

func testValidModelName(randoms *testutils.RandomValues, router *chi.Router, modelName string) {
	resp := callHuggingFaceModelInfoApi(randoms, router, modelName, "")
	Expect(resp.Code).To(Equal(http.StatusOK))
}

func testInvalidHfToken(randoms *testutils.RandomValues, router *chi.Router, hfToken string) {
	resp := callHuggingFaceModelInfoApi(randoms, router, "valid-model", hfToken)
	Expect(resp.Code).To(Equal(http.StatusBadRequest))
	Expect(resp.Body.String()).To(ContainSubstring("value for HfToken must comply to the pattern"))
}

var _ = Describe("Test Hugging Face Model Info - Inference Service", Ordered, func() {
	var randoms *testutils.RandomValues
	router := CreateRouter()

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()
	})

	Context("valid request", func() {
		It("should succeed", func() {
			resp := callHuggingFaceModelInfoApi(randoms, &router, "valid-model", "")
			Expect(resp.Code).To(Equal(http.StatusOK))

			var modelInfo api.HuggingFaceModelInfo
			err := json.Unmarshal(resp.Body.Bytes(), &modelInfo)
			Expect(err).To(BeNil())

			Expect(modelInfo.ModelName).To(Equal("valid-model"))
		})
	})

	Context("invalid request", func() {
		It("should fail", func() {
			resp := callHuggingFaceModelInfoApi(randoms, &router, "no-such-model", "")
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body.String()).To(ContainSubstring("no-such-model"))
		})
	})

	Context("model name format validations", func() {
		It("should succeed when given a valid name format", func() {
			testValidModelName(randoms, &router, "meta-llama/Llama-2-7b-chat-hf")
			testValidModelName(randoms, &router, "Llama-2-7b-chat-hf")
			testValidModelName(randoms, &router, "microsoft/Phi-3-medium-128k-instruct")
			testValidModelName(randoms, &router, "mistralai/mamba-codestral-7B-v0.1")
		})

		It("should fail when given an invalid name format", func() {
			testInvalidModelName(randoms, &router, "model name with spaces")
			testInvalidModelName(randoms, &router, " meta-llama/Llama-2-7b-chat-hf")
			testInvalidModelName(randoms, &router, "meta-llama/Llama-2-7b-chat-hf ")
			testInvalidModelName(randoms, &router, "meta-llama/ Llama-2-7b-chat-hf")
			testInvalidModelName(randoms, &router, "Llama-2-7b-chat-hf/")
			testInvalidModelName(randoms, &router, "/meta-llama/Llama-2-7b-chat-hf")
			testInvalidModelName(randoms, &router, "microsoft.dot.com/model")
		})
	})

	Context("token format validations", func() {
		It("should succeed when given a valid token format", func() {
			testValidHfToken(randoms, &router, "hf_aabbccDDEEFFgghhiiKKLLMMnnooPP1234")
			testValidHfToken(randoms, &router, "hf_123")
		})

		It("should fail when given an invalid token format", func() {
			testInvalidHfToken(randoms, &router, "invalid_token_format")
			testInvalidHfToken(randoms, &router, " hf_aabbccDDEEFFgghhiiKKLLMMnnooPP1234")
			testInvalidHfToken(randoms, &router, "hf_aabbccDDEEFFgghhiiKKLLMMnnooPP1234 ")
			testInvalidHfToken(randoms, &router, "hf_aabbccDDEEFFgghhii KKLLMMnnooPP1234")
			testInvalidHfToken(randoms, &router, "hf")
			testInvalidHfToken(randoms, &router, "hf_")
			testInvalidHfToken(randoms, &router, "hf_aabbccDDEEFFgghhiiKKLLMMnnooPP.dot")
		})
	})

	Context("too many concurrent requests", func() {
		It("should succeed up to the capacity, then fail", func() {
			config.Init()
			maxConcurrency := runai_config.Get[*config.HuggingFaceConfig]().MaxConcurrentRequests

			// use a WaitGroup to ensure all goroutines complete before the test ends
			var wg sync.WaitGroup

			// start maxConcurrency requests as slow requests, in order for the extra one to get rejected
			for i := 0; i < maxConcurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					callHuggingFaceModelInfoApi(randoms, &router, "slow-response-model", "")
				}()
			}

			// the extra requests that should fail

			time.Sleep(1 * time.Second) // wait for the slow requests to start

			resp := callHuggingFaceModelInfoApi(randoms, &router, "slow-response-model", "")
			Expect(resp.Code).To(Equal(http.StatusServiceUnavailable))

			// wait for all goroutines to complete before the test ends
			wg.Wait()
		})
	})
})
