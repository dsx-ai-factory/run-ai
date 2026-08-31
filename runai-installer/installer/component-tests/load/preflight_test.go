/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package load_test

import (
	"errors"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_PreflightProbesForPushAccess(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	res := run(t, loadConfig(t, bundle), false)
	require.NoError(t, res.err)

	assert.Contains(t, res.pusher.Checks, "registry.local:5000/runai/runai-airgap-probe:probe")
	assert.Contains(t, res.pusher.Checks, "registry.local:5000/runai/runai-control-plane/assets-service:2.26.13")
	assert.Contains(t, res.pusher.Checks, "registry.local:5000/runai/runai-control-plane/postgresql:16")
}

func TestLoad_PreflightFailures(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{
			name:    "unreachable",
			err:     errors.New("connection refused"),
			message: "connection refused",
		},
		{
			name:    "credentials rejected",
			err:     errors.New("UNAUTHORIZED"),
			message: "UNAUTHORIZED",
		},
		{
			name:    "no push access",
			err:     errors.New("DENIED"),
			message: "DENIED",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bundle := harness.CreateLoadBundle(t, harness.TestLock())

			f := newFakes()
			f.pusher.CheckError = tc.err

			res := runFakes(t, loadConfig(t, bundle), f)

			require.ErrorIs(t, res.err, specrun.ErrPreflightFailed)
			assert.Empty(t, res.pusher.Pushes, "nothing is pushed when preflight fails")
			assert.Contains(t, res.stdout, tc.message)
			assert.Contains(t, res.stdout, "provide credentials with push access")
		})
	}
}

func TestLoad_PreflightOnlyStopsBeforeTheBundle(t *testing.T) {
	bundle := harness.CreateLoadBundle(t, harness.TestLock())

	f := newFakes()
	res := runStaged(t, loadConfig(t, bundle), f, func(o *specrun.ExecOptions) { o.PreflightOnly = true })

	require.NoError(t, res.err)
	assert.Len(t, f.pusher.Checks, 1)
	assert.Empty(t, res.pusher.Pushes)
	assert.Nil(t, findEvent(res.events, "step_succeeded", "extract-bundle"))
}
