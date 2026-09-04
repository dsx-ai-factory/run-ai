/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sinkerrors

import "errors"

var (
	ErrCannotReachHost    = errors.New("cannot reach host")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTLS                = errors.New("tls error")
	ErrAuthRequired       = errors.New("authentication required")
	ErrSenderRejected     = errors.New("sender rejected")
	ErrRelayDenied        = errors.New("relay denied")
	ErrRecipientRejected  = errors.New("recipient rejected")
)
