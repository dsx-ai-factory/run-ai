/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	sasl "github.com/emersion/go-sasl"
	go_smtp "github.com/emersion/go-smtp"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/internal/sinks/sinkerrors"
)

const (
	verifyTimeout = 30 * time.Second
	// maxVerifyRecipients bounds the sequential RCPT TO probes a single
	// verification request can trigger.
	maxVerifyRecipients = 10
)

type SmtpClienter interface {
	Auth(sasl.Client) error
	Mail(string, *go_smtp.MailOptions) error
	Rcpt(string, *go_smtp.RcptOptions) error
	SendMail(string, []string, io.Reader) error
	Quit() error
	Close() error
}

type ClientProvider func(withDeadline bool) (SmtpClienter, error)

type SmtpEmailSender struct {
	config         *SmtpConfig
	auth           sasl.Client
	clientProvider ClientProvider
}

func NewSmtpEmailSender(config *SmtpConfig, auth sasl.Client) *SmtpEmailSender {
	s := &SmtpEmailSender{config: config, auth: auth}
	s.clientProvider = s.openClient
	return s
}

func (s *SmtpEmailSender) SendMail(from string, to []string, msg []byte) error {
	client, err := s.clientProvider(false)
	if err != nil {
		return err
	}
	defer closeClient(client)
	return client.SendMail(from, to, bytes.NewReader(msg))
}

func (s *SmtpEmailSender) Verify(from string, to []string) error {
	client, err := s.clientProvider(true)
	if err != nil {
		return err
	}
	defer closeClient(client)

	return verifyEnvelope(client, s.config.SmtpHost, from, to)
}

func (s *SmtpEmailSender) openClient(withDeadline bool) (SmtpClienter, error) {
	addr := net.JoinHostPort(s.config.SmtpHost, fmt.Sprintf("%d", s.config.SmtpPort))
	tlsCfg := &tls.Config{ServerName: s.config.SmtpHost, MinVersion: tls.VersionTLS12}

	var client *go_smtp.Client
	var err error
	if s.config.SmtpTlsEnabled {
		client, err = go_smtp.DialStartTLS(addr, tlsCfg)
		if err != nil {
			if classified := classifyDialErr(err); errors.Is(classified, sinkerrors.ErrCannotReachHost) {
				return nil, classified
			}
			return nil, sinkerrors.ErrTLS
		}
	} else {
		client, err = go_smtp.Dial(addr)
		if err != nil {
			return nil, classifyDialErr(err)
		}
	}

	if withDeadline {
		client.CommandTimeout = verifyTimeout
	}

	if s.auth == nil {
		return client, nil
	}

	if err := client.Auth(s.auth); err != nil {
		closeClient(client)
		return nil, classifyAuthErr(err)
	}

	return client, nil
}

func closeClient(client SmtpClienter) {
	if client.Quit() != nil {
		_ = client.Close()
	}
}

func verifyEnvelope(client SmtpClienter, host, from string, to []string) error {
	if err := client.Mail(from, nil); err != nil {
		detail := fmt.Errorf("smtp server %q rejected sender %q: %w", host, from, err)
		return errors.Join(classifySenderErr(err), detail)
	}

	if len(to) > maxVerifyRecipients {
		log.Warn().Msgf("verifying only the first %d of %d recipients", maxVerifyRecipients, len(to))
		to = to[:maxVerifyRecipients]
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient, nil); err != nil {
			detail := fmt.Errorf("smtp server %q rejected recipient %q: %w", host, recipient, err)
			return errors.Join(classifyRecipientErr(err), detail)
		}
	}
	return nil
}

func classifySenderErr(err error) error {
	e := smtpError(err)
	switch {
	case e == nil:
		return nil
	case isAuthRequired(e):
		return sinkerrors.ErrAuthRequired
	case e.Code >= 500:
		return sinkerrors.ErrSenderRejected
	default:
		return nil
	}
}

func classifyRecipientErr(err error) error {
	e := smtpError(err)
	switch {
	case e == nil:
		return nil
	case isRelayDenied(e):
		return sinkerrors.ErrRelayDenied
	case e.Code >= 500:
		return sinkerrors.ErrRecipientRejected
	default:
		return nil
	}
}

func smtpError(err error) *go_smtp.SMTPError {
	var e *go_smtp.SMTPError
	errors.As(err, &e)
	return e
}

func isAuthRequired(e *go_smtp.SMTPError) bool {
	return e.Code == 530 || e.Code == 538
}

func isRelayDenied(e *go_smtp.SMTPError) bool {
	if e.Code == 530 {
		return true
	}
	msg := strings.ToLower(e.Message)
	return (e.Code == 550 || e.Code == 553 || e.Code == 554) && strings.Contains(msg, "relay")
}

func classifyDialErr(err error) error {
	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return sinkerrors.ErrCannotReachHost
	}
	return err
}

func classifyAuthErr(err error) error {
	if errors.Is(err, go_smtp.ErrAuthUnsupported) || errors.Is(err, go_smtp.ErrAuthFailed) {
		return sinkerrors.ErrInvalidCredentials
	}
	smtpErr := smtpError(err)
	if smtpErr != nil && smtpErr.Code == 535 {
		return sinkerrors.ErrInvalidCredentials
	}
	return err
}
