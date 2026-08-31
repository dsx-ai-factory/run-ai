/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

import (
	"fmt"
	"strings"

	sasl "github.com/emersion/go-sasl"
)

// loginAuth implements SASL LOGIN as used by Microsoft Office365 and MailHog:
// AUTH LOGIN with no initial response, then Username:/Password: challenges.
// sasl.NewLoginClient is unsuitable — it sends the username as an initial
// response and only accepts an exact "Password:" challenge.
type loginAuth struct {
	username, password string
}

func newLoginAuth(username, password string) sasl.Client {
	return &loginAuth{username: username, password: password}
}

func (a *loginAuth) Start() (mech string, ir []byte, err error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(challenge []byte) (response []byte, err error) {
	command := string(challenge)
	command = strings.TrimSpace(command)
	command = strings.TrimSuffix(command, ":")
	command = strings.ToLower(command)

	switch command {
	case "username":
		return []byte(a.username), nil
	case "password":
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("unexpected server challenge: %s", command)
	}
}
