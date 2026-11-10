// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package completion

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestGPUTypesCompleteCaseInsensitively(t *testing.T) {
	fn := GetFlagCompletionFn(GPUType, func(*cobra.Command) error { return nil }, func(*cobra.Command) ([]string, error) { return []string{"NVIDIA-A100", "Tesla-V100"}, nil })
	got, directive := fn(&cobra.Command{}, nil, "nViDiA")
	if !reflect.DeepEqual(got, []string{"NVIDIA-A100"}) || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("got %v directive %v", got, directive)
	}
	got, directive = fn(&cobra.Command{}, nil, "missing")
	if len(got) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("got %v directive %v", got, directive)
	}
}
