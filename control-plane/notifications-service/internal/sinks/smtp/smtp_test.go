/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"testing"

	sasl "github.com/emersion/go-sasl"
	go_smtp "github.com/emersion/go-smtp"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/sinkerrors"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/smtp/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newSender(cfg *SmtpConfig, auth sasl.Client, session SmtpClienter) *SmtpEmailSender {
	return &SmtpEmailSender{
		config:         cfg,
		auth:           auth,
		clientProvider: func(bool) (SmtpClienter, error) { return session, nil },
	}
}

func newSenderWithErr(cfg *SmtpConfig, sessionErr error) *SmtpEmailSender {
	return &SmtpEmailSender{
		config:         cfg,
		clientProvider: func(bool) (SmtpClienter, error) { return nil, sessionErr },
	}
}

func testCfg(authType AuthType) *SmtpConfig {
	return &SmtpConfig{SmtpHost: "fake.smtp", SmtpPort: 587, AuthType: authType, User: "alice", Password: "secret"}
}

func TestSendMail_NoneAuth(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("SendMail", "from@example.com", []string{"to@example.com"}, mock.Anything).Return(nil)
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).SendMail("from@example.com", []string{"to@example.com"}, []byte("hello"))
	assert.NoError(t, err)
}

func TestSendMail_SessionOpenError_Propagates(t *testing.T) {
	err := newSenderWithErr(testCfg(AuthNone), sinkerrors.ErrCannotReachHost).
		SendMail("from@example.com", []string{"to@example.com"}, []byte("hello"))
	assert.ErrorIs(t, err, sinkerrors.ErrCannotReachHost)
}

func TestSendMail_ForceTLS_STARTTLSFailure_ReturnsErrTLS(t *testing.T) {
	cfg := testCfg(AuthNone)
	cfg.SmtpTlsEnabled = true
	err := newSenderWithErr(cfg, sinkerrors.ErrTLS).
		SendMail("from@example.com", []string{"to@example.com"}, []byte("hello"))
	assert.ErrorIs(t, err, sinkerrors.ErrTLS)
}

func TestSendMail_AuthFailure_ReturnsErrInvalidCredentials(t *testing.T) {
	// Auth runs inside openClient; a failed open surfaces as ErrInvalidCredentials.
	err := newSenderWithErr(testCfg(AuthLogin), sinkerrors.ErrInvalidCredentials).
		SendMail("from@example.com", []string{"to@example.com"}, []byte("hello"))
	assert.ErrorIs(t, err, sinkerrors.ErrInvalidCredentials)
}

func TestSendMail_PassesNoDeadline(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("SendMail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	session.On("Quit").Return(nil)

	var withDeadline bool
	sender := &SmtpEmailSender{
		config: testCfg(AuthNone),
		clientProvider: func(d bool) (SmtpClienter, error) {
			withDeadline = d
			return session, nil
		},
	}
	require.NoError(t, sender.SendMail("from@example.com", []string{"to@example.com"}, []byte("hello")))
	assert.False(t, withDeadline)
}

func TestVerify_NoneAuth(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", nil)
	assert.NoError(t, err)
}

func TestVerify_LoginAuth(t *testing.T) {
	// Auth already completed in openClient; Verify only probes the envelope.
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Quit").Return(nil)

	auth := sasl.NewLoginClient("alice", "secret")
	err := newSender(testCfg(AuthLogin), auth, session).Verify("from@example.com", nil)
	assert.NoError(t, err)
}

func TestVerify_PassesDeadline(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", mock.Anything, mock.Anything).Return(nil)
	session.On("Quit").Return(nil)

	var withDeadline bool
	sender := &SmtpEmailSender{
		config: testCfg(AuthNone),
		clientProvider: func(d bool) (SmtpClienter, error) {
			withDeadline = d
			return session, nil
		},
	}
	require.NoError(t, sender.Verify("from@example.com", nil))
	assert.True(t, withDeadline)
}

// --- verifyEnvelope ---

func TestVerifyEnvelope_MailRejected_ReturnsSenderErr(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "bad@sender.com", (*go_smtp.MailOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 550, Message: "sender rejected"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("bad@sender.com", nil)
	assert.ErrorIs(t, err, sinkerrors.ErrSenderRejected)
}

func TestVerifyEnvelope_MailAuthRequired_ReturnsErrAuthRequired(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 530, Message: "Authentication required"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", nil)
	assert.ErrorIs(t, err, sinkerrors.ErrAuthRequired)
}

func TestVerifyEnvelope_SingleRecipientAccepted(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", "to@example.com", (*go_smtp.RcptOptions)(nil)).Return(nil)
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", []string{"to@example.com"})
	assert.NoError(t, err)
}

func TestVerifyEnvelope_RecipientRejected_ReturnsErrRecipientRejected(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", "bad@domain.invalid", (*go_smtp.RcptOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 550, Message: "user unknown"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", []string{"bad@domain.invalid"})
	assert.ErrorIs(t, err, sinkerrors.ErrRecipientRejected)
}

func TestVerifyEnvelope_RelayDenied_ReturnsErrRelayDenied(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", "to@external.com", (*go_smtp.RcptOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 550, Message: "Relay access denied"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", []string{"to@external.com"})
	assert.ErrorIs(t, err, sinkerrors.ErrRelayDenied)
}

func TestVerifyEnvelope_RelayDeniedVia530_ReturnsErrRelayDenied(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", "to@external.com", (*go_smtp.RcptOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 530, Message: "Authentication required"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", []string{"to@external.com"})
	assert.ErrorIs(t, err, sinkerrors.ErrRelayDenied)
}

func TestVerifyEnvelope_TruncatesAtMaxRecipients(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Quit").Return(nil)

	recipients := make([]string, maxVerifyRecipients+3)
	for i := range recipients {
		recipients[i] = "r@example.com"
	}
	session.On("Rcpt", "r@example.com", (*go_smtp.RcptOptions)(nil)).
		Return(nil).Times(maxVerifyRecipients)

	err := newSender(testCfg(AuthNone), nil, session).Verify("from@example.com", recipients)
	assert.NoError(t, err)
}

func TestVerifyEnvelope_StopsAtFirstRecipientError(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", "from@example.com", (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", "good@example.com", (*go_smtp.RcptOptions)(nil)).Return(nil)
	session.On("Rcpt", "bad@example.com", (*go_smtp.RcptOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 550, Message: "user unknown"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).
		Verify("from@example.com", []string{"good@example.com", "bad@example.com"})
	assert.ErrorIs(t, err, sinkerrors.ErrRecipientRejected)
}

// --- classifyAuthErr ---

func TestClassifyAuthErr_Code535_ReturnsErrInvalidCredentials(t *testing.T) {
	err := classifyAuthErr(&go_smtp.SMTPError{Code: 535, Message: "Authentication failed"})
	assert.ErrorIs(t, err, sinkerrors.ErrInvalidCredentials)
}

func TestClassifyAuthErr_ErrAuthFailed_ReturnsErrInvalidCredentials(t *testing.T) {
	err := classifyAuthErr(go_smtp.ErrAuthFailed)
	assert.ErrorIs(t, err, sinkerrors.ErrInvalidCredentials)
}

func TestClassifyAuthErr_ErrAuthUnsupported_ReturnsErrInvalidCredentials(t *testing.T) {
	err := classifyAuthErr(go_smtp.ErrAuthUnsupported)
	assert.ErrorIs(t, err, sinkerrors.ErrInvalidCredentials)
}

// --- isAuthRequired / isRelayDenied ---

func TestIsAuthRequired_Code530(t *testing.T) {
	assert.True(t, isAuthRequired(&go_smtp.SMTPError{Code: 530}))
}

func TestIsAuthRequired_Code538(t *testing.T) {
	assert.True(t, isAuthRequired(&go_smtp.SMTPError{Code: 538}))
}

func TestIsAuthRequired_OtherCode(t *testing.T) {
	assert.False(t, isAuthRequired(&go_smtp.SMTPError{Code: 550}))
}

func TestIsRelayDenied_Code530(t *testing.T) {
	assert.True(t, isRelayDenied(&go_smtp.SMTPError{Code: 530, Message: "auth required"}))
}

func TestIsRelayDenied_Code550WithRelayMsg(t *testing.T) {
	assert.True(t, isRelayDenied(&go_smtp.SMTPError{Code: 550, Message: "Relay access denied"}))
}

func TestIsRelayDenied_Code550WithoutRelayMsg(t *testing.T) {
	assert.False(t, isRelayDenied(&go_smtp.SMTPError{Code: 550, Message: "User unknown"}))
}

// Compile-time interface checks.
var _ SmtpClienter = (*mocks.SmtpClienter)(nil)
var _ SmtpClient = (*SmtpEmailSender)(nil)
