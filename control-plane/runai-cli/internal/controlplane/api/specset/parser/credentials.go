/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parser

// Personal-credential attribute keys used in submit (e.g. --env-my-credentials type=...,name=...,credential-name=...).
const (
	PersonalCredentialAttributeType     = "type"
	PersonalCredentialAttributeName     = "name"
	PersonalCredentialAttributeCredName = "credential-name"
	PersonalCredentialAttributeKey      = "key"
)

const (
	ErrorPersonalCredentialKeyRequired   = "key is required for type=%s"
	ErrorPersonalCredentialKeyDisallowed = "key is not allowed for type=%s"
	ErrorPersonalCredentialInvalidType   = "invalid type %q: must be one of %s, %s, %s"
)
