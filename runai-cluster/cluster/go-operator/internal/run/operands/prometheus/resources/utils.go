/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
)

func findRemoteWrite(remoteWrites []manifests.RemoteWriteSpec, remoteWrite manifests.RemoteWriteSpec) (bool, int) {
	for index, rw := range remoteWrites {
		if rw.Name == remoteWrite.Name {
			return true, index
		}
	}
	return false, -1
}
