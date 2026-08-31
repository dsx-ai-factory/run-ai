/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package huggingface

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
)

var languageModelTags = map[string]bool{
	"text-generation":          true, // Main tag for LLMs supported by vLLM
	"text2text-generation":     true, // For models that transform text to text
	"text-classification":      true, // For classification tasks
	"summarization":            true, // For text summarization
	"question-answering":       true, // For QA tasks
	"translation":              true, // For translation tasks
	"token-classification":     true, // For token-level classification
	"table-question-answering": true, // For table-based QA
	"zero-shot-classification": true, // For zero-shot classification
	"sentence-similarity":      true, // For semantic similarity
	"feature-extraction":       true, // For feature extraction
	"fill-mask":                true, // For masked language modeling
	"causal-lm":                true, // For causal language models
	"seq2seq-lm":               true, // For sequence-to-sequence models
	"conversational":           true, // For conversational models
	"chat":                     true, // For chat models
	"instruct":                 true, // For instruction-tuned models
	"code":                     true, // For code models
}

type ExternalApiInterface interface {
	modelApiCall(ctx context.Context, modelInfoUrl string, modelName string, token *string) (string, error)
	modelFileApiCall(ctx context.Context, modelFileUrl string, modelName string, token *string) (int, error)
}

type ExternalApi struct{}

type ModelInfoApiInterface interface {
	GetModelInfo(ctx context.Context, request api.HuggingFaceModelInfoRequest) (*api.HuggingFaceModelInfo, *api.Error)
}

type ModelInfoApi struct {
	ModelInfoUrl string
	ModelFileUrl string
	ExternalApi  ExternalApiInterface
}

func NewModelInfoApi(modelInfoUrl string, modelFileUrl string) ModelInfoApiInterface {
	return &ModelInfoApi{
		ModelInfoUrl: modelInfoUrl,
		ModelFileUrl: modelFileUrl,
		ExternalApi:  &ExternalApi{},
	}
}

// the following structs are used to parse the response from the Hugging Face model API
type hfModelInfo struct {
	Error       string               `json:"error"`
	Gated       BoolOrString         `json:"gated"`
	PipelineTag string               `json:"pipeline_tag"`
	Tags        []string             `json:"tags"`
	Config      hfApiModelInfoConfig `json:"config"`
}

// the "gated" response field is sometimes boolean sometimes string:
// both <false>, <"auto"> and <"manual"> are valid responses,
// so we need to handle it in a special way, by having a special unmarshal function
// that sets it to true iff its value as a string is not "false".
// (we don't care about the difference between "auto" and "manual" at this stage,
// they're both simply considered as true)
type BoolOrString bool

func (b *BoolOrString) UnmarshalJSON(data []byte) error {
	asString := string(data)
	*b = asString != "false"

	return nil
}

type hfApiModelInfoConfig struct {
	Architectures []string `json:"architectures"`
}

func (hf *ModelInfoApi) GetModelInfo(ctx context.Context, request api.HuggingFaceModelInfoRequest) (*api.HuggingFaceModelInfo, *api.Error) {
	// inputs (modelName and hfToken) are validated in the swagger level

	// call Hugging Face model API
	// in this call, we don't use the given HF token, if any, since if the token is invalid, HF will not return
	// the model info, but respond with an error instead. we prefer to get the info and only later handle the token
	modelInfo, err := hf.hfModelApi(ctx, hf.ModelInfoUrl, request.ModelName, nil)
	if err != nil || modelInfo.Error != "" {
		return nil, hf.parseHfModelApiError(request.ModelName, request.HfToken, err, modelInfo)
	}

	// call Hugging Face file API to check the token
	// note that previously, the "model API" didn't test the token, so we did it via the "model file API". now HF
	// have changed the behavior of the model API, but we prefer to keep this implementation as it seems more stable
	gated := bool(modelInfo.Gated)
	authorized := !gated
	if gated && request.HfToken != nil {
		var apiErr *api.Error
		authorized, apiErr = hf.hfModelFileApi(ctx, hf.ModelFileUrl, request.ModelName, request.HfToken)
		if apiErr != nil {
			return nil, apiErr
		}
	}

	info := api.HuggingFaceModelInfo{
		ModelName:  request.ModelName,
		Gated:      gated,
		Authorized: authorized,
		Language:   isLanguageModel(modelInfo.PipelineTag, modelInfo.Tags),
	}

	return &info, nil
}

func (hf *ModelInfoApi) hfModelApi(ctx context.Context, modelInfoUrl string, modelName string, token *string) (*hfModelInfo, error) {
	body, err := hf.ExternalApi.modelApiCall(ctx, modelInfoUrl, modelName, token)
	if err != nil {
		return nil, err
	}

	var modelInfo hfModelInfo
	if err := json.Unmarshal([]byte(body), &modelInfo); err != nil {
		return nil, err
	}

	return &modelInfo, nil
}

func (hf *ModelInfoApi) parseHfModelApiError(modelName string, token *string, err error, modelInfo *hfModelInfo) *api.Error {
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to call hugging face api: %s", err.Error()), nil)
	}

	errorMessage := fmt.Sprintf("failed to get model info: %s", modelInfo.Error)

	// note that we no longer receive all the below error messages from HF, but we keep handling them for
	// backward compatibility. due to a change in api behavior on HF side, they no longer validate the token and whether
	// it allows to access the given model. from our current experience, neither "You must be authenticated" and
	// "restricted and you are not in the authorized list" are returned, so our validation is lacking in this regard

	if strings.Contains(modelInfo.Error, "Invalid username or password") {
		errorMessage = "the provided model was not found on hugging face"
	} else if strings.Contains(modelInfo.Error, "You must be authenticated") {
		if token == nil {
			errorMessage = "access to the model is gated, please provide a token"
		} else {
			errorMessage = "access to the model is gated, please verify that the provided token is valid"
		}
	} else if strings.Contains(modelInfo.Error, "is restricted and you are not in the authorized list") {
		errorMessage = "access to the model is gated, please verify that the provided token is authorized for this model"
	}

	return api.NewError(http.StatusBadRequest, errorMessage, nil)
}

func (hf *ModelInfoApi) hfModelFileApi(ctx context.Context, modelFileUrl string, modelName string, token *string) (bool, *api.Error) {
	status, err := hf.ExternalApi.modelFileApiCall(ctx, modelFileUrl, modelName, token)
	if err != nil {
		return false, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to call hugging face file api: %s", err.Error()), nil)
	}

	// we expect one of the following statuses:
	// OK (200) - means the token is authorized for the model
	// Unauthorized (401) - means the token is invalid, i.e. not known by HF
	// Forbidden (403) - means the token is ok but is not authorized for the model
	// any other status should return error
	if status != http.StatusOK && status != http.StatusUnauthorized && status != http.StatusForbidden {
		return false, api.NewError(http.StatusServiceUnavailable, fmt.Sprintf("failed to call hugging face file api: %d", status), nil)
	}

	authorized := status == http.StatusOK

	return authorized, nil
}

// isLanguageModel returns true iff the given pipeline tag is one of the tags
// that are used by language models. the main tag for LLMs is "text-generation"
func isLanguageModel(pipelineTag string, tags []string) bool {
	if languageModelTags[pipelineTag] {
		return true
	}

	for _, tag := range tags {
		if languageModelTags[tag] {
			return true
		}
	}

	return false
}

func (ea *ExternalApi) modelApiCall(ctx context.Context, modelInfoUrl string, modelName string, token *string) (string, error) {
	url := fmt.Sprintf("%s/%s", modelInfoUrl, modelName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	if token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (ea *ExternalApi) modelFileApiCall(ctx context.Context, modelFileUrl string, modelName string, token *string) (int, error) {
	url := strings.Replace(modelFileUrl, "<HF_MODEL>", modelName, 1)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	if token != nil {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	return resp.StatusCode, nil
}

type ModelInfoApiMock struct{}

func NewModelInfoApiMock() ModelInfoApiInterface {
	return &ModelInfoApiMock{}
}

func (hf *ModelInfoApiMock) GetModelInfo(ctx context.Context, request api.HuggingFaceModelInfoRequest) (*api.HuggingFaceModelInfo, *api.Error) {
	modelInfo := api.HuggingFaceModelInfo{
		ModelName: request.ModelName,
	}

	if request.ModelName == "communication-error" {
		if request.HfToken == nil {
			return nil, api.NewError(http.StatusBadRequest, "communication error", nil)
		}
	}

	if request.ModelName == "slow-response-model" {
		// simulate a non-trivial latency
		time.Sleep(2 * time.Second)
	}

	return &modelInfo, nil
}
