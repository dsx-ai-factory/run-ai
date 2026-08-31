/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type RunaiClaims struct {
	Issuer   string           `json:"iss,omitempty"`
	Subject  string           `json:"sub,omitempty"`
	Audience jwt.ClaimStrings `json:"aud,omitempty"`

	ExpiresAt *jwt.NumericDate `json:"exp,omitempty"`
	NotBefore *jwt.NumericDate `json:"nbf,omitempty"`
	IssuedAt  *jwt.NumericDate `json:"iat,omitempty"`

	TenantId            flexInt  `json:"tenant_id,omitempty"`
	ClientId            string   `json:"client_id,omitempty"`
	Groups              []string `json:"groups,omitempty"`
	SupplementaryGroups []int    `json:"supplementarygroups,omitempty"`
	ClientType          string   `json:"client_type,omitempty"`
	UID                 string   `json:"uid,omitempty"`
	GID                 string   `json:"gid,omitempty"`
	SID                 string   `json:"sid,omitempty"`
	IdentityProvider    string   `json:"identity_provider,omitempty"`
}

func (c *RunaiClaims) valid() error {
	now := time.Now()
	if c.ExpiresAt != nil && now.After(c.ExpiresAt.Time) {
		return fmt.Errorf("token is expired")
	}
	if c.NotBefore != nil && now.Before(c.NotBefore.Time) {
		return fmt.Errorf("token is not valid yet")
	}
	if c.IssuedAt != nil && now.Before(c.IssuedAt.Time) {
		return fmt.Errorf("token used before issued")
	}
	return nil
}

// flexInt unmarshals a JSON value that is either a bare integer or a quoted integer string.
type flexInt int

func (f *flexInt) UnmarshalJSON(b []byte) error {
	var i int
	if err := json.Unmarshal(b, &i); err == nil {
		*f = flexInt(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = flexInt(i)
	return nil
}
