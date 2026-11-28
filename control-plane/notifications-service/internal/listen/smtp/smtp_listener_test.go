/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseBodyToEvent_ValidJSON(t *testing.T) {
	listener := &SmtpListener{id: "test-listener"}
	jsonBody := `-----BEGIN NOTIFICATION-----{"foo": "bar", "num": 42}-----END NOTIFICATION-----`
	event := listener.parseBodyToEvent(jsonBody)
	require.NotNil(t, event)
	require.Equal(t, event.ListenerId, "test-listener")
	require.Equal(t, event.Properties["foo"], "bar")
	require.Equal(t, event.Properties["num"], float64(42))
}

func TestParseBodyToEvent_InvalidJSON(t *testing.T) {
	listener := &SmtpListener{id: "test-listener"}
	invalidBody := `-----BEGIN NOTIFICATION-----not a json-----END NOTIFICATION-----`
	event := listener.parseBodyToEvent(invalidBody)
	require.Nil(t, event)
}

func TestParseBodyToEvent_WithLineBreaksInField(t *testing.T) {
	listener := &SmtpListener{id: "test-listener"}
	jsonBody := "-----BEGIN NOTIFICATION-----{\"foo\": \"bar\r\nbar\"}-----END NOTIFICATION-----"
	event := listener.parseBodyToEvent(jsonBody)
	require.NotNil(t, event)
	require.Equal(t, "test-listener", event.ListenerId)
	require.Equal(t, "barbar", event.Properties["foo"])
}

func TestParseBodyToEvent_WithEscapedLineBreaksInField(t *testing.T) {
	listener := &SmtpListener{id: "test-listener"}
	jsonBody := "-----BEGIN NOTIFICATION-----{\"foo\": \"bar\\r\\nbar\"}-----END NOTIFICATION-----"
	event := listener.parseBodyToEvent(jsonBody)
	require.NotNil(t, event)
	require.Equal(t, "test-listener", event.ListenerId)
	require.Equal(t, "bar\r\nbar", event.Properties["foo"])
}

func TestParseBodyToEvent_WithSoftLineBreaksInField(t *testing.T) {
	listener := &SmtpListener{id: "test-listener"}
	jsonBody := "-----BEGIN NOTIFICATION-----{\"foo\": \"bar=\r\nbar\"}-----END NOTIFICATION-----"
	event := listener.parseBodyToEvent(jsonBody)
	require.NotNil(t, event)
	require.Equal(t, "test-listener", event.ListenerId)
	require.Equal(t, "barbar", event.Properties["foo"])
}

func TestParseBodyToEvent_WithHeaders(t *testing.T) {
	listener := &SmtpListener{id: "test-listener"}
	jsonBody := `-----BEGIN NOTIFICATION-----{"tenant_id": "foo", "user_id": "bar"}-----END NOTIFICATION-----`
	msg := fmt.Sprintf("From: frommmm\r\nTo: toooo\r\nSubject: Test JSON\r\n\r\n%s", jsonBody)
	event := listener.parseBodyToEvent(msg)
	require.NotNil(t, event)
	require.Equal(t, "test-listener", event.ListenerId)
	require.Equal(t, "foo", event.Properties["tenant_id"])
	require.Equal(t, "bar", event.Properties["user_id"])
}

func TestNewSmtpListenerAddr(t *testing.T) {
	for _, tc := range []struct {
		name string
		host string
		want string
	}{
		{name: "hostname", host: "localhost", want: "localhost:2525"},
		{name: "ipv4", host: "127.0.0.1", want: "127.0.0.1:2525"},
		{name: "ipv6", host: "::1", want: "[::1]:2525"},
		{name: "bracketed ipv6", host: "[::1]", want: "[::1]:2525"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			l := NewSmtpListener("id", nil, &SmtpListenerConfig{Host: tc.host, Port: 2525})
			require.Equal(t, tc.want, l.server.Addr)
		})
	}
}
