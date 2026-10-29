/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimcp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// fakeOpenAPIError mimics a generated-client GenericOpenAPIError: it carries the raw
// response body and a decoded model (nil when the client failed to decode).
type fakeOpenAPIError struct {
	msg   string
	body  []byte
	model any
}

func (e fakeOpenAPIError) Error() string { return e.msg }
func (e fakeOpenAPIError) Body() []byte  { return e.body }
func (e fakeOpenAPIError) Model() any    { return e.model }

func TestMapHTTPError_SurfacesBodyOnDecodeFailure(t *testing.T) {
	resp := func(code int) *http.Response { return &http.Response{StatusCode: code} }
	decodeErr := fakeOpenAPIError{
		msg:  "json: cannot unmarshal string into Go value of type map[string]interface {}",
		body: []byte(`"can not enable node level scheduling while gpu-over-provisioning flag is disabled"`),
	}

	t.Run("bare_string_body_replaces_decode_noise", func(t *testing.T) {
		err := MapHTTPError(resp(http.StatusBadRequest), decodeErr)
		assert.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "gpu-over-provisioning")
		assert.NotContains(t, err.Error(), "unmarshal")
	})
	t.Run("status_still_classified", func(t *testing.T) {
		err := MapHTTPError(resp(http.StatusForbidden), decodeErr)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Contains(t, err.Error(), "gpu-over-provisioning")
	})
	t.Run("nil_response_wraps_upstream", func(t *testing.T) {
		err := MapHTTPError(nil, decodeErr)
		assert.ErrorIs(t, err, ErrUpstream)
		assert.Contains(t, err.Error(), "gpu-over-provisioning")
	})
	t.Run("decoded_model_keeps_its_message", func(t *testing.T) {
		withModel := fakeOpenAPIError{msg: "quota exceeded", body: []byte(`{"message":"x"}`), model: struct{}{}}
		err := MapHTTPError(resp(http.StatusBadRequest), withModel)
		assert.Contains(t, err.Error(), "quota exceeded")
	})
	t.Run("decoded_model_with_only_status_in_error", func(t *testing.T) {
		err := MapHTTPError(&http.Response{StatusCode: http.StatusBadRequest, Status: "400 Bad Request"}, fakeOpenAPIError{
			msg: "400 Bad Request", model: struct{}{},
			body: []byte(`{"code":400,"message":"To apply a policy, there must be at least one default or rule"}`),
		})
		assert.ErrorIs(t, err, ErrValidation)
		assert.Contains(t, err.Error(), "at least one default or rule")
	})
	t.Run("plain_error_untouched", func(t *testing.T) {
		err := MapHTTPError(resp(http.StatusNotFound), errors.New("boom"))
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Contains(t, err.Error(), "boom")
	})
	t.Run("long_body_truncated", func(t *testing.T) {
		long := strings.Repeat("x", maxUpstreamBodyLen+50)
		err := MapHTTPError(resp(http.StatusInternalServerError), fakeOpenAPIError{msg: "decode fail", body: []byte(long)})
		assert.ErrorIs(t, err, ErrUpstream)
		assert.Contains(t, err.Error(), "…")
		assert.Less(t, len(err.Error()), len(long))
	})
	t.Run("multibyte_body_truncated_on_rune_boundary", func(t *testing.T) {
		long := strings.Repeat("é", maxUpstreamBodyLen+50)
		err := MapHTTPError(resp(http.StatusInternalServerError), fakeOpenAPIError{msg: "decode fail", body: []byte(long)})
		assert.True(t, utf8.ValidString(err.Error()), "truncation must not split a multibyte rune")
		assert.Contains(t, err.Error(), "…")
	})
}

func TestMapHTTPError(t *testing.T) {
	resp := func(code int) *http.Response { return &http.Response{StatusCode: code} }

	for _, tc := range []struct {
		status int
		want   error
	}{
		{http.StatusBadRequest, ErrValidation},
		{http.StatusUnprocessableEntity, ErrValidation},
		{http.StatusUnauthorized, ErrAuth},
		{http.StatusForbidden, ErrForbidden},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusConflict, ErrConflict},
		{http.StatusTooManyRequests, ErrUpstreamThrottled},
		{http.StatusBadGateway, ErrUpstream}, // unmapped status → upstream
	} {
		err := MapHTTPError(resp(tc.status), errors.New("boom"))
		assert.ErrorIs(t, err, tc.want, "status %d", tc.status)
		assert.Contains(t, err.Error(), "boom", "status %d keeps the detail", tc.status)
	}

	// The generated-client placeholder carries no signal and must not leak.
	clean := MapHTTPError(resp(http.StatusForbidden), errors.New("undefined response type"))
	assert.ErrorIs(t, clean, ErrForbidden)
	assert.NotContains(t, clean.Error(), "undefined response type")

	assert.ErrorIs(t, MapHTTPError(nil, errors.New("x")), ErrUpstream)
	assert.Nil(t, MapHTTPError(resp(http.StatusForbidden), nil))
}

func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want errorCategory
	}{
		{"nil", nil, catOK},
		{"auth", fmt.Errorf("x: %w", ErrAuth), catAuth},
		{"forbidden", fmt.Errorf("x: %w", ErrForbidden), catForbidden},
		{"not_found", fmt.Errorf("x: %w", ErrNotFound), catNotFound},
		{"throttled", fmt.Errorf("x: %w", ErrUpstreamThrottled), catRateLimit},
		{"upstream", fmt.Errorf("x: %w", ErrUpstream), catUpstream},
		{"confirmation_required", fmt.Errorf("x: %w", ErrConfirmationRequired), catConfirmationRequired},
		{"declined", fmt.Errorf("x: %w", ErrDeclined), catDeclined},
		{"timeout", context.DeadlineExceeded, catTimeout},
		{"canceled", context.Canceled, catCanceled},
		{"validation", validationError{inner: errors.New("bad")}, catValidation},
		{"internal", ErrInternal, catInternal},
		{"fallback", errors.New("nothing in particular"), catError},
		{"auth_wins_over_upstream", fmt.Errorf("%w: %w", ErrAuth, ErrUpstream), catAuth},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, classify(tc.err))
		})
	}
}
