/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/flow"
)

// ContainsValue reports whether needle is present in values, used for asserting
// on helm --set arguments captured by FakeHelmInstaller.
func ContainsValue(values []string, needle string) bool {
	for _, v := range values {
		if v == needle {
			return true
		}
	}
	return false
}

// UseFastPollInterval shrinks flow.DefaultPollInterval for the duration of the
// test so that the FlowManager's status-polling ticker fires quickly instead of
// the 30s production default. Restored on test cleanup.
func UseFastPollInterval(t *testing.T) {
	t.Helper()
	orig := flow.DefaultPollInterval
	flow.DefaultPollInterval = 100 * time.Millisecond
	t.Cleanup(func() { flow.DefaultPollInterval = orig })
}
