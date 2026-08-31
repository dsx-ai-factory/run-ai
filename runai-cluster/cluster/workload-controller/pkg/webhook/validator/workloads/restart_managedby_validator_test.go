/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

func dwWithManagedByAndLabel(managedBy string, labelVal string) *jobs.JobWorkload {
	w := &jobs.JobWorkload{}
	w.Spec.ManagedBy = managedBy
	if labelVal != "" {
		if w.Labels == nil {
			w.Labels = make(map[string]string)
		}
		w.Labels[RestartRequestUIDLabel] = labelVal
	}
	return w
}

func TestValidateDistributedWorkloadRestartManagedBy(t *testing.T) {
	tests := []struct {
		name        string
		old         *jobs.JobWorkload
		new         *jobs.JobWorkload
		wantErr     bool
		errContains string
	}{
		{
			name:    "allow set managedBy to restart-controller with label",
			old:     dwWithManagedByAndLabel("", ""),
			new:     dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-123"),
			wantErr: false,
		},
		{
			name:    "allow clear managedBy and remove label",
			old:     dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-123"),
			new:     dwWithManagedByAndLabel("", ""),
			wantErr: false,
		},
		{
			name:        "deny managedBy to non-allowed value",
			old:         dwWithManagedByAndLabel("", ""),
			new:         dwWithManagedByAndLabel("other-controller", ""),
			wantErr:     true,
			errContains: "can only be",
		},
		{
			name:        "deny restart-controller without label",
			old:         dwWithManagedByAndLabel("", ""),
			new:         dwWithManagedByAndLabel(RestartControllerManagedBy, ""),
			wantErr:     true,
			errContains: "must be set with a non-empty value",
		},
		{
			name: "deny restart-controller with empty label value",
			old:  dwWithManagedByAndLabel("", ""),
			new: func() *jobs.JobWorkload {
				w := dwWithManagedByAndLabel(RestartControllerManagedBy, "x")
				w.Labels[RestartRequestUIDLabel] = ""
				return w
			}(),
			wantErr:     true,
			errContains: "must be set with a non-empty value",
		},
		{
			name:        "deny empty managedBy with label still present",
			old:         dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-123"),
			new:         dwWithManagedByAndLabel("", "uid-123"),
			wantErr:     true,
			errContains: "must be removed",
		},
		{
			name:        "deny modifying existing label value",
			old:         dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-123"),
			new:         dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-456"),
			wantErr:     true,
			errContains: "cannot be modified",
		},
		{
			name:    "allow same managedBy and label unchanged",
			old:     dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-123"),
			new:     dwWithManagedByAndLabel(RestartControllerManagedBy, "uid-123"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDistributedWorkloadRestartManagedBy(tt.old, tt.new)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDistributedWorkloadRestartManagedBy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("ValidateDistributedWorkloadRestartManagedBy() error = %v, want contains %q", err, tt.errContains)
			}
		})
	}
}
