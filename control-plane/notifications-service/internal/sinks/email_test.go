/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinks

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/events"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks/smtp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type SmtpClientMock struct {
	MsgSent  string
	verified bool
}

func (s *SmtpClientMock) SendMail(from string, to []string, msg []byte) error {
	s.MsgSent = string(msg)
	return nil
}

func (s *SmtpClientMock) Verify(from string, to []string) error {
	s.verified = true
	return nil
}

var (
	validUser         = "test@run.ai"
	from              = "user@run.ai"
	fromDisplayName   = "NVIDIA Run:ai"
	logoUrl           = "https://s3.amazonaws.com/www.run.ai/runai-logo.png"
	recipientsDefault = []string{
		"user@run.ai",
	}

	eventExample = events.Event{
		UID: "test",
		Properties: map[string]interface{}{
			"user_id":             validUser,
			"reason":              "Scheduled",
			"namespace":           "test-namespace",
			"type":                "test-name",
			"category":            "Pod",
			"note":                "test message",
			"cluster":             "test-cluster",
			"workloadDisplayName": "test-workload-display-name",
		},
	}
)

func TestNewEmailSink(t *testing.T) {
	templatesFolder := "../../sinks_templates/email"
	templateFile := "generic_notification.tpl"

	// To stop server use Stop() method
	testCases := []struct {
		desc          string
		config        EmailSinkConfig
		event         events.Event
		expectedErr   error
		expectedInMsg string
	}{
		{
			desc: "test happy flow",
			config: EmailSinkConfig{
				FromDisplayName:     fromDisplayName,
				From:                from,
				TemplatesFolder:     templatesFolder,
				TemplateFile:        templateFile,
				LogoUrl:             logoUrl,
				DirectNotifications: false,
				Recipients:          recipientsDefault,
			},
			event: eventExample,

			expectedErr:   nil,
			expectedInMsg: "",
		},
		{
			desc: "test no recipients",
			config: EmailSinkConfig{
				FromDisplayName:     fromDisplayName,
				From:                from,
				TemplatesFolder:     templatesFolder,
				TemplateFile:        templateFile,
				LogoUrl:             logoUrl,
				DirectNotifications: false,
				Recipients:          []string{},
			},
			event:         eventExample,
			expectedErr:   nil,
			expectedInMsg: "",
		},
		{
			desc: "test subject",
			config: EmailSinkConfig{
				FromDisplayName:     fromDisplayName,
				From:                from,
				TemplatesFolder:     templatesFolder,
				TemplateFile:        templateFile,
				LogoUrl:             logoUrl,
				DirectNotifications: false,
				Recipients:          recipientsDefault,
			},
			event: eventExample,

			expectedErr:   nil,
			expectedInMsg: "From: NVIDIA Run:ai  <user@run.ai>\r\nTo: user@run.ai <user@run.ai>\r\nSubject:  \r\n",
		},
		{
			desc: "test subject from event",
			config: EmailSinkConfig{
				FromDisplayName:     fromDisplayName,
				From:                from,
				TemplatesFolder:     templatesFolder,
				TemplateFile:        templateFile,
				LogoUrl:             logoUrl,
				DirectNotifications: false,
				Recipients:          recipientsDefault,
			},
			event: events.Event{Properties: map[string]interface{}{"subject": "test"}},

			expectedErr:   nil,
			expectedInMsg: "Subject: test \r\n",
		},
		{
			desc: "test long link",
			config: EmailSinkConfig{
				FromDisplayName:     fromDisplayName,
				From:                from,
				TemplatesFolder:     templatesFolder,
				TemplateFile:        "keycloak_reset.tpl",
				LogoUrl:             logoUrl,
				DirectNotifications: false,
				Recipients:          recipientsDefault,
			},
			event: events.Event{Properties: map[string]interface{}{"link": "https://test-env.example.com/auth/realms/runai/login-actions/action-token?key=token123token123token123token123token123token123token123token123token123token123token123token123token123token123token123token123token123token123token123token123"}},

			expectedErr:   nil,
			expectedInMsg: "a href=3D\"https://test-env.example.com/auth/realms/runai/login-act=\r\nions/action-token?key=3Dtoken123token123token123token123token123token123tok=\r\nen123token123token123token123token123token123token123token123token123token1=\r\n23token123token123token123token123\"",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc := tc
			fmt.Printf("#### test case: %s #### \n", tc.desc)
			// To test with real smtp server use SmtpClientImpl{} instead of SmtpClientMock{}
			sinker := EmailSink{
				config:     &tc.config,
				smtpClient: &SmtpClientMock{},
				encoder:    &smtp.BodyEncoder{},
			}

			err := sinker.Sink(&tc.event)
			assert.ErrorIs(t, err, tc.expectedErr)
			assert.Contains(t, sinker.smtpClient.(*SmtpClientMock).MsgSent, tc.expectedInMsg)
		})
	}
}

func TestTemplate_ParsesSubject(t *testing.T) {
	templatesFolder := "../../sinks_templates/email"
	templateFile := "generic_notification.tpl"
	config := &EmailSinkConfig{
		FromDisplayName:     fromDisplayName,
		From:                from,
		TemplatesFolder:     templatesFolder,
		TemplateFile:        templateFile,
		LogoUrl:             logoUrl,
		DirectNotifications: false,
		Recipients:          recipientsDefault,
		Subject:             "this is the subject {{ .test }}",
	}
	sinker := EmailSink{
		config:     config,
		smtpClient: &SmtpClientMock{},
		encoder:    &smtp.BodyEncoder{},
	}
	event := &events.Event{
		Properties: map[string]interface{}{"test": "test"},
	}
	err := sinker.Sink(event)
	assert.Nil(t, err)
	assert.Contains(t, sinker.smtpClient.(*SmtpClientMock).MsgSent, "this is the subject test")
}

func TestHeaderInjection(t *testing.T) {
	testCases := []struct {
		desc            string
		event           events.Event
		subject         string
		expectedSubject string
	}{
		{
			desc:            "crlf in subject property",
			event:           events.Event{Properties: map[string]interface{}{"subject": "hello\r\nBcc: attacker@evil.com"}},
			expectedSubject: "hello  Bcc: attacker@evil.com",
		},
		{
			desc:            "crlf in a property interpolated into the configured subject",
			event:           events.Event{Properties: map[string]interface{}{"test": "hello\r\nBcc: attacker@evil.com"}},
			subject:         "notification {{ .test }}",
			expectedSubject: "notification hello  Bcc: attacker@evil.com",
		},
		{
			desc:            "bare lf in subject property",
			event:           events.Event{Properties: map[string]interface{}{"subject": "hello\nBcc: attacker@evil.com"}},
			expectedSubject: "hello Bcc: attacker@evil.com",
		},
		{
			desc:            "nul in subject property",
			event:           events.Event{Properties: map[string]interface{}{"subject": "hello\x00truncated"}},
			expectedSubject: "hellotruncated",
		},
		{
			desc:            "template directive in subject property is not executed",
			event:           events.Event{Properties: map[string]interface{}{"subject": "{{ .secret }}", "secret": "leaked"}},
			subject:         "unused",
			expectedSubject: "{{ .secret }}",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			sinker := EmailSink{
				config: &EmailSinkConfig{
					FromDisplayName: fromDisplayName,
					From:            from,
					TemplatesFolder: "../../sinks_templates/email",
					TemplateFile:    "generic_notification.tpl",
					LogoUrl:         logoUrl,
					Recipients:      recipientsDefault,
					Subject:         tc.subject,
				},
				smtpClient: &SmtpClientMock{},
				encoder:    &smtp.BodyEncoder{},
			}

			require.NoError(t, sinker.Sink(&tc.event))

			lines := strings.Split(sinker.smtpClient.(*SmtpClientMock).MsgSent, HeadersDelimiter)
			require.Greater(t, len(lines), 3)
			assert.True(t, strings.HasPrefix(lines[0], "From: "))
			assert.True(t, strings.HasPrefix(lines[1], "To: "))
			assert.Equal(t, "Subject: "+tc.expectedSubject+" ", lines[2])
			assert.False(t, strings.HasPrefix(lines[3], "Bcc:"), "no header was injected after Subject")
		})
	}
}

func TestVerifyFuncInvoked(t *testing.T) {
	smtpClientMock := &SmtpClientMock{}
	sinker := EmailSink{
		config:     &EmailSinkConfig{},
		smtpClient: smtpClientMock,
		encoder:    &smtp.BodyEncoder{},
	}

	require.NoError(t, sinker.Verify())
	require.True(t, smtpClientMock.verified)
}

func TestNewEmailSinkConfigDefaults(t *testing.T) {
	sinker, err := NewEmailSinkConfig(viper.New(), "id")

	require.NoError(t, err)
	require.Equal(t, "nvidia-runai@nvidia.com", sinker.From)
	require.Equal(t, "NVIDIA Run:ai", sinker.FromDisplayName)
	require.Equal(t, "https://s3.amazonaws.com/www.run.ai/OS_NVIDIA_Runai_black.svg", sinker.LogoUrl)
	require.Equal(t, "", sinker.Smtp.SmtpHost)
	require.Equal(t, 587, sinker.Smtp.SmtpPort)
}

func TestSanitizeDescriptionHTML_WorkAsExpected_DifferentScenarios(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// safe HTML tags- output remain unchanged
		{"simple description", "simple description"},
		{"<div><h1>Hi!</h1></div>", "<div><h1>Hi!</h1></div>"},
		{"<p>This is <strong>bold</strong> text.</p>", "<p>This is <strong>bold</strong> text.</p>"},
		{`<p style="font-size: bold;">Click me</p>`, `<p style="font-size: bold;">Click me</p>`},
		{"<ul><li>Item 1</li><li>Item 2</li></ul>", "<ul><li>Item 1</li><li>Item 2</li></ul>"},
		{`<div><h1 style="height: 96px;">title</h1><p style="color: red;">content!</p></div>`, `<div><h1 style="height: 96px;">title</h1><p style="color: red;">content!</p></div>`},
		// filter out unsafe HTML tags
		{`<script>alert("alerting!")</script>`, ""},
		{`<iframe src="http://example.com"></iframe>`, ""},
		{`<div onmouseover="alert('alerting!')">Hover me!</div>`, "<div>Hover me!</div>"},
		// mixed safe and unsafe tags
		{`<div><h1>Title</h1><script>alert('XSS!')</script></div>`, "<div><h1>Title</h1></div>"},
		{`<a href="javascript:alert('XSS!')">Click me</a>`, "Click me"},
		{`<a onblur="alert(secret)" href="http://www.google.com">Google</a>`, `<a href="http://www.google.com" rel="nofollow">Google</a>`},
		// not existing tags
		{`<notatag>Some text</notatag>`, "Some text"},
		// full html documents
		{
			`<html>
			<head>
				<title>some title</title>
			</head>
			<body style="display:flex; align-items:center; justify-content:center;"></body>
			</html>`,
			"",
		},
		{
			`<html><body><h1>Header</h1><button onclick="alert('XSS!')">Press</button><script>alert('Hack!')</script></body></html>`,
			`<h1>Header</h1>Press`,
		},
	}

	for _, tc := range testCases {
		output := sanitizeDescriptionHTML(tc.input)
		trimmedOutput := strings.TrimSpace(output)
		trimmedExpectedOutput := strings.TrimSpace(tc.expected)
		require.Equal(t, trimmedOutput, trimmedExpectedOutput)
	}
}
