/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads_service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
)

type partialErrorReader struct{}

func (partialErrorReader) Read(buffer []byte) (int, error) {
	count := copy(buffer, "synthetic-hf-prefix")
	return count, errors.New("synthetic read failure")
}

func TestRedactSensitiveValuesFromEncodedDownstreamBody(t *testing.T) {
	sensitiveValue := `synthetic-"quoted"-<html>-value`
	request := map[string]interface{}{
		"environmentVariables": []interface{}{
			map[string]interface{}{"name": "NGC_API_KEY", "value": sensitiveValue},
		},
	}
	body, err := json.Marshal(map[string]string{"message": "invalid request containing " + sensitiveValue})
	if err != nil {
		t.Fatalf("failed to build downstream body: %v", err)
	}

	redacted := redactSensitiveValuesFromDownstreamBody(body, request)

	if strings.Contains(string(redacted), sensitiveValue) || strings.Contains(string(redacted), `synthetic-\"quoted\"-\u003chtml\u003e-value`) {
		t.Fatalf("encoded sensitive value was returned in downstream error: %s", redacted)
	}
	var response map[string]string
	if err := json.Unmarshal(redacted, &response); err != nil {
		t.Fatalf("failed to parse redacted downstream body: %v", err)
	}
	if response["message"] != "invalid request containing [REDACTED]" {
		t.Fatalf("unexpected redacted message: %q", response["message"])
	}
}

func TestRedactSensitiveValuesFromHTMLEscapedDownstreamBody(t *testing.T) {
	request := map[string]interface{}{
		"environmentVariables": []interface{}{
			map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-<html>-value"},
		},
	}
	body := []byte(`{"message":"invalid synthetic-&lt;html&gt;-value"}`)

	redacted := string(redactSensitiveValuesFromDownstreamBody(body, request))

	if strings.Contains(redacted, "synthetic-&lt;html&gt;-value") || !strings.Contains(redacted, "[REDACTED]") {
		t.Fatalf("HTML-escaped sensitive value was not redacted: %s", redacted)
	}
}

func TestRedactSensitiveValuesFromDownstreamBody(t *testing.T) {
	request := map[string]interface{}{
		"spec": map[string]interface{}{
			"environmentVariables": []interface{}{
				map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-hf-value"},
				map[string]interface{}{"name": "NORMAL_VAR", "value": "visible"},
			},
		},
	}
	body := []byte(`{"message":"invalid request containing synthetic-hf-value and visible"}`)

	redacted := string(redactSensitiveValuesFromDownstreamBody(body, request))

	if strings.Contains(redacted, "synthetic-hf-value") {
		t.Fatalf("sensitive value was returned in downstream error: %s", redacted)
	}
	if !strings.Contains(redacted, "[REDACTED]") || !strings.Contains(redacted, "visible") {
		t.Fatalf("expected only the sensitive value to be redacted: %s", redacted)
	}
}

func TestRedactSensitiveValuesFromDownstreamBodyFailsClosed(t *testing.T) {
	body := []byte(`{"message":"downstream body that must not be returned"}`)

	redacted := string(redactSensitiveValuesFromDownstreamBody(body, make(chan int)))

	if strings.Contains(redacted, "downstream body that must not be returned") {
		t.Fatalf("downstream body was returned when the request could not be inspected: %s", redacted)
	}
	if !strings.Contains(redacted, "request could not be inspected") {
		t.Fatalf("expected a generic failure message, got: %s", redacted)
	}
}

func TestReadAndRedactSensitiveDownstreamBodyRejectsPartialReads(t *testing.T) {
	request := map[string]interface{}{
		"environmentVariables": []interface{}{
			map[string]interface{}{"name": "HF_TOKEN", "value": "synthetic-hf-prefix-and-suffix"},
		},
	}

	body, err := readAndRedactSensitiveDownstreamBody(partialErrorReader{}, request)

	if err == nil {
		t.Fatal("expected the partial read to fail")
	}
	if body != nil {
		t.Fatalf("partial downstream body must be discarded, got: %q", body)
	}
}

func TestSubmissionErrorFromDownstreamResponseRejectsNull(t *testing.T) {
	responseError := submissionErrorFromDownstreamResponse(
		http.StatusBadRequest,
		strings.NewReader("null"),
		map[string]interface{}{},
		"workspace",
	)

	if responseError == nil {
		t.Fatal("expected a non-nil error for a null downstream response")
	}
	if responseError.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, responseError.Code)
	}
	if !strings.Contains(responseError.Message, "invalid downstream error response") {
		t.Fatalf("unexpected error message: %q", responseError.Message)
	}
}
