/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package field_types

// +k8s:deepcopy-gen=true
type MpiField struct {
	Values *MpiMembers `json:"values,omitempty"`
}

// +k8s:deepcopy-gen=true
type MpiMembers struct {
	// Specifies the number of slots per worker used in hostfile. Defaults to 1.
	SlotsPerWorker *IntField `json:"slotsPerWorker,omitempty"`
	// Specifies the directory where SSH keys are mounted.
	SshAuthMountPath *StringField `json:"sshAuthMountPath,omitempty"`
	// Specifies if the MPI launcher will start with the workers or only after they are ready (supported from mpi-operator v0.5).
	LauncherCreationPolicy *StringField `json:"launcherCreationPolicy,omitempty"`
}

func (self *MpiField) Deref() MpiMembers {
	if self != nil && self.Values != nil {
		return *self.Values
	} else {
		return MpiMembers{}
	}
}
