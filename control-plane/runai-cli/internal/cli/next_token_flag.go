/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/structs"
)

const (
	typeNextToken = "next_token"
)

type NextTokenFlag struct {
	nextToken *structs.NextToken
}

func NewNextTokenFlag(value *structs.NextToken) *NextTokenFlag {
	return &NextTokenFlag{nextToken: value}
}

func (nf *NextTokenFlag) String() string {
	if !nf.IsSet() {
		return ""
	}
	return nf.nextToken.String()
}

func (nf *NextTokenFlag) IsSet() bool {
	return nf.nextToken.IsSet()
}

func (nf *NextTokenFlag) Set(value string) error {
	nextToken, err := structs.NextTokenFromValue(value)
	if err != nil {
		return err
	}
	*nf.nextToken = *nextToken // Copies the value to the original address
	return nil
}

func (nf *NextTokenFlag) Get() *structs.NextToken {
	return nf.nextToken
}

// Type for CLI help value
func (nf *NextTokenFlag) Type() string {
	return typeNextToken
}
