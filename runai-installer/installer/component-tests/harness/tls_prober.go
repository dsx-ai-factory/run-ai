/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// TLSProbeCall records one ProbeCertificate invocation for assertions.
type TLSProbeCall struct {
	Host   string
	Port   int
	CAPath string
}

// FakeTLSProber is a configurable orchestrator2.TLSProber for tests. It returns
// Result (and Err) for every probe and records the calls.
type FakeTLSProber struct {
	Result orchestrator2.TLSProbeResult
	Err    error
	Calls  []TLSProbeCall
}

func (f *FakeTLSProber) ProbeCertificate(_ context.Context, host string, port int, caPEMPath string) (orchestrator2.TLSProbeResult, error) {
	f.Calls = append(f.Calls, TLSProbeCall{Host: host, Port: port, CAPath: caPEMPath})
	if f.Err != nil {
		return orchestrator2.TLSProbeResult{}, f.Err
	}
	return f.Result, nil
}
