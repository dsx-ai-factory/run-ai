/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package secrets

import (
	"encoding/base64"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword()
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := base64.URLEncoding.DecodeString(password)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 32 {
		t.Fatalf("expected 32 random bytes, got %d", len(decoded))
	}
}
