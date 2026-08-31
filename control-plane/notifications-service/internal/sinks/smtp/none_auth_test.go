/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"testing"

	go_smtp "github.com/emersion/go-smtp"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/sinkerrors"
	"github.com/run-ai/runai/control-plane/notifications-service/internal/sinks/smtp/mocks"
	"github.com/stretchr/testify/require"
)

const (
	probeSender    = "runai@example.com"
	probeRecipient = "alerts@example.com"
)

// TestNoneAuthVerify_ServerAdvertisesAuthButAllowsUnauthenticated guards against
// a regression: advertising AUTH must not fail auth_none verification when the
// server permits unauthenticated delivery (no Auth call; MAIL FROM succeeds).
func TestNoneAuthVerify_ServerAdvertisesAuthButAllowsUnauthenticated(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", probeSender, (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify(probeSender, nil)
	require.NoError(t, err)
}

// TestNoneAuthVerify_ServerRequiresAuth ensures auth_none verification fails
// when the server rejects the unauthenticated mail transaction.
func TestNoneAuthVerify_ServerRequiresAuth(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", probeSender, (*go_smtp.MailOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 530, Message: "Authentication required"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify(probeSender, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, sinkerrors.ErrAuthRequired)
}

// TestNoneAuthVerify_RecipientAccepted verifies that when recipients are
// provided, verification also probes RCPT TO and passes when accepted.
func TestNoneAuthVerify_RecipientAccepted(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", probeSender, (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", probeRecipient, (*go_smtp.RcptOptions)(nil)).Return(nil)
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify(probeSender, []string{probeRecipient})
	require.NoError(t, err)
}

// TestNoneAuthVerify_RelayDenied verifies that a server which accepts the sender
// but denies relaying to the recipient fails verification once recipients are provided.
func TestNoneAuthVerify_RelayDenied(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", probeSender, (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", probeRecipient, (*go_smtp.RcptOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 550, Message: "relay access denied - please authenticate"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify(probeSender, []string{probeRecipient})
	require.Error(t, err)
	require.ErrorIs(t, err, sinkerrors.ErrRelayDenied)
}

// TestNoneAuthVerify_RecipientRejected verifies that a non-relay recipient
// rejection is classified distinctly from a denied relay.
func TestNoneAuthVerify_RecipientRejected(t *testing.T) {
	session := mocks.NewSmtpClienter(t)
	session.On("Mail", probeSender, (*go_smtp.MailOptions)(nil)).Return(nil)
	session.On("Rcpt", probeRecipient, (*go_smtp.RcptOptions)(nil)).
		Return(&go_smtp.SMTPError{Code: 550, Message: "unroutable domain: asdas.com"})
	session.On("Quit").Return(nil)

	err := newSender(testCfg(AuthNone), nil, session).Verify(probeSender, []string{probeRecipient})
	require.Error(t, err)
	require.ErrorIs(t, err, sinkerrors.ErrRecipientRejected)
}
