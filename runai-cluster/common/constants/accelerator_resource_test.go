/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

import (
	"strings"
	"testing"
)

func TestAcceleratorResourceNodeLabel(t *testing.T) {
	if label := AcceleratorResourceNodeLabel("nvidia.com/gpu"); label != AcceleratorResourceLabelPrefix+"nvidia.com_gpu" {
		t.Errorf("unexpected label for a standard resource: %q", label)
	}
	longName := strings.Repeat("a", 48) + ".example/device"
	first := AcceleratorResourceNodeLabel(longName)
	_, name, found := strings.Cut(first, "/")
	if !found || len(name) > 63 || first != AcceleratorResourceNodeLabel(longName) || first == AcceleratorResourceNodeLabel(longName+"2") {
		t.Errorf("invalid or ambiguous long resource label %q", first)
	}
}
