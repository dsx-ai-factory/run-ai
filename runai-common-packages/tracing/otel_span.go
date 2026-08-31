/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing

import (
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type SpanStatus string

const (
	SpanStatusOK    SpanStatus = "Ok"
	SpanStatusError SpanStatus = "Error"
)

// OtelSpan implements the OtelSpan interface.
type OtelSpan struct {
	span trace.Span
}

// End ends the span.
func (s *OtelSpan) End() {
	s.span.End()
}

// SetAttributes sets attributes on the span.
func (s *OtelSpan) SetAttribute(key, value string) Span {
	s.span.SetAttributes(attribute.String(key, value))
	return s
}

// SetStatus sets status on the span.
func (s *OtelSpan) SetStatus(status SpanStatus, description string) Span {
	s.span.SetStatus(getStatusAndDescription(status, description))
	return s
}

// RecordError records error.
func (s *OtelSpan) RecordError(err error) Span {
	s.span.RecordError(err)
	return s
}

// GetSpanID returns the span ID.
func (s *OtelSpan) GetSpanID() string {
	return s.span.SpanContext().SpanID().String()
}

func getStatusAndDescription(status SpanStatus, description string) (codes.Code, string) {
	switch status {
	case SpanStatusError:
		return codes.Error, description
	case SpanStatusOK:
		return codes.Ok, description
	}
	if description != "" {
		return codes.Error, fmt.Sprintf("undefined status [%s]: %s", status, description)
	}
	return codes.Error, fmt.Sprintf("undefined status [%s]", status)
}
