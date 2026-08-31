/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package io

import (
	"io"

	"github.com/spf13/cobra"
)

type Std struct {
	out io.Writer
	in  io.Reader
}

func NewStd(cmd *cobra.Command) *Std {
	return &Std{out: cmd.OutOrStdout(), in: cmd.InOrStdin()}
}

func (s *Std) Read(p []byte) (n int, err error) {
	return s.in.Read(p)
}

func (s *Std) Write(p []byte) (n int, err error) {
	return s.out.Write(p)
}

func (s *Std) Close() error {
	return nil
}
