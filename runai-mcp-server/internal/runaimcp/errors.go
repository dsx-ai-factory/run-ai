/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaimcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrInternal             = errors.New("internal error")
	ErrValidation           = errors.New("validation error")
	ErrNotFound             = errors.New("not found")
	ErrForbidden            = errors.New("forbidden")
	ErrAuth                 = errors.New("authentication failed")
	ErrUpstreamThrottled    = errors.New("upstream throttled")
	ErrUpstream             = errors.New("upstream error")
	ErrConflict             = errors.New("conflict")
	ErrConfirmationRequired = errors.New("confirmation_required")
	ErrDeclined             = errors.New("declined")
)

type errorCategory string

const (
	catOK                   errorCategory = "ok"
	catValidation           errorCategory = "validation"
	catNotFound             errorCategory = "not_found"
	catForbidden            errorCategory = "forbidden"
	catAuth                 errorCategory = "auth"
	catRateLimit            errorCategory = "rate_limited"
	catUpstream             errorCategory = "upstream"
	catConflict             errorCategory = "conflict"
	catConfirmationRequired errorCategory = "confirmation_required"
	catDeclined             errorCategory = "declined"
	catTimeout              errorCategory = "timeout"
	catCanceled             errorCategory = "canceled"
	catInternal             errorCategory = "internal"
	catError                errorCategory = "error"
)

func classify(err error) errorCategory {
	switch {
	case err == nil:
		return catOK
	case errors.Is(err, ErrAuth):
		return catAuth
	case errors.Is(err, ErrForbidden):
		return catForbidden
	case errors.Is(err, ErrDeclined):
		return catDeclined
	case errors.Is(err, ErrConfirmationRequired):
		return catConfirmationRequired
	case errors.Is(err, ErrConflict):
		return catConflict
	case errors.Is(err, ErrNotFound):
		return catNotFound
	case errors.Is(err, ErrUpstreamThrottled):
		return catRateLimit
	case errors.Is(err, context.DeadlineExceeded):
		return catTimeout
	case errors.Is(err, context.Canceled):
		return catCanceled
	case errors.Is(err, ErrUpstream):
		return catUpstream
	case errors.Is(err, ErrValidation):
		return catValidation
	case errors.Is(err, ErrInternal):
		return catInternal
	default:
		return catError
	}
}

type validationError struct{ inner error }

func (v validationError) Error() string { return v.inner.Error() }
func (v validationError) Unwrap() error { return v.inner }
func (v validationError) Is(target error) bool {
	return target == ErrValidation
}

func MapHTTPError(resp *http.Response, err error) error {
	if err == nil {
		return nil
	}
	err = preferUpstreamBody(resp, err)
	sentinel := ErrUpstream
	if resp != nil {
		switch resp.StatusCode {
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			sentinel = ErrValidation
		case http.StatusUnauthorized:
			sentinel = ErrAuth
		case http.StatusForbidden:
			sentinel = ErrForbidden
		case http.StatusNotFound:
			sentinel = ErrNotFound
		case http.StatusConflict:
			sentinel = ErrConflict
		case http.StatusTooManyRequests:
			sentinel = ErrUpstreamThrottled
		}
	}

	// The generated OpenAPI clients return this exact placeholder when a response
	// body doesn't match a declared schema (common on error statuses whose body is
	// a plain string). It carries no signal, so surface only the classified
	// sentinel. Matched by equality, not Contains, to avoid scrubbing legitimate
	// text — this couples to the client's literal, so revisit if it ever wraps it.
	if err.Error() == "undefined response type" {
		return sentinel
	}
	return fmt.Errorf("%w: %w", sentinel, err)
}

// maxUpstreamBodyLen caps how much of a raw upstream error body we surface, so a
// stray HTML gateway page can't flood the LLM-facing error.
const maxUpstreamBodyLen = 300

// preferUpstreamBody recovers the API message when the generated client reports
// only an HTTP status or a decoding error.
func preferUpstreamBody(resp *http.Response, err error) error {
	var oe interface {
		Body() []byte
		Model() any
	}
	if !errors.As(err, &oe) {
		return err
	}
	// A decoded model can still lose its message if the client only formats the status.
	if oe.Model() != nil && (resp == nil || err.Error() != resp.Status) {
		return err
	}
	msg := strings.TrimSpace(string(oe.Body()))
	if msg == "" {
		return err
	}
	var structured struct {
		Message string `json:"message"`
	}
	var unquoted string
	if json.Unmarshal(oe.Body(), &structured) == nil && strings.TrimSpace(structured.Message) != "" {
		msg = strings.TrimSpace(structured.Message)
	} else if json.Unmarshal(oe.Body(), &unquoted) == nil && strings.TrimSpace(unquoted) != "" {
		msg = strings.TrimSpace(unquoted)
	}
	if r := []rune(msg); len(r) > maxUpstreamBodyLen {
		msg = string(r[:maxUpstreamBodyLen]) + "…"
	}
	return errors.New(msg)
}
