/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-service/internal/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Event properties are deserialized from untrusted JSON, so any value may be of
// any type. These tests pin that non-string values are ignored rather than
// panicking the process.
var malformedUserIds = map[string]interface{}{
	"number": float64(42),
	"bool":   true,
	"object": map[string]interface{}{"email": validUser},
	"array":  []interface{}{validUser},
	"null":   nil,
}

func TestEmailGetRecipients_NonStringUserId(t *testing.T) {
	for name, userId := range malformedUserIds {
		t.Run(name, func(t *testing.T) {
			sink := &EmailSink{config: &EmailSinkConfig{
				DirectNotifications: true,
				Recipients:          recipientsDefault,
			}}
			event := &events.Event{Properties: map[string]interface{}{"user_id": userId}}

			require.NotPanics(t, func() {
				assert.Equal(t, recipientsDefault, sink.getRecipients(event))
			})
		})
	}
}

func TestSlackGetRecipients_NonStringUserId(t *testing.T) {
	for name, userId := range malformedUserIds {
		t.Run(name, func(t *testing.T) {
			sink := &SlackSink{config: &SlackSinkConfig{
				DirectNotifications: true,
				Recipients:          recipientIds,
			}}
			event := &events.Event{Properties: map[string]interface{}{"user_id": userId}}

			require.NotPanics(t, func() {
				assert.Equal(t, recipientIds, sink.getRecipients(event))
			})
		})
	}
}

func TestEmailGetTemplateData_NonStringDescription(t *testing.T) {
	for name, description := range malformedUserIds {
		t.Run(name, func(t *testing.T) {
			sink := &EmailSink{config: &EmailSinkConfig{
				TemplatesFolder: "../../sinks_templates/email",
				TemplateFile:    "generic_notification.tpl",
			}}
			event := &events.Event{Properties: map[string]interface{}{"description": description}}

			require.NotPanics(t, func() {
				assert.Empty(t, sink.getTemplateData(event).Description)
			})
		})
	}
}

func TestEmailGetTemplateData_StringDescriptionIsSanitized(t *testing.T) {
	sink := &EmailSink{config: &EmailSinkConfig{}}
	event := &events.Event{Properties: map[string]interface{}{
		"description": "<b>hello</b><script>alert(1)</script>",
	}}

	assert.Equal(t, "<b>hello</b>", string(sink.getTemplateData(event).Description))
}
