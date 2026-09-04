/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package control_plane_details_downloader

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runaiv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

var _ = Describe("Types", func() {
	Describe("isLocalCluster", func() {
		cases := map[string]struct {
			cpdsd    *InstallationConfigurationData
			expected bool
		}{
			"local": {
				cpdsd: &InstallationConfigurationData{
					Internal: runaiv1_internal.Internal{
						Global: &runaiv1_internal.Global{
							ClusterURL: "http://localhost:8080",
							ControlPlane: &runaiv1_internal.ControlPlane{
								URL: "http://localhost:8080",
							},
						},
					},
				},
				expected: true,
			},
			"remote": {
				cpdsd: &InstallationConfigurationData{
					Internal: runaiv1_internal.Internal{
						Global: &runaiv1_internal.Global{
							ClusterURL: "http://localhost:8080",
							ControlPlane: &runaiv1_internal.ControlPlane{
								URL: "http://test.run.ai",
							},
						},
					},
				},
				expected: false,
			},
		}

		for name, c := range cases {
			c := c
			It(name, func() {
				isLocal, err := isLocalCluster(c.cpdsd.Internal.Global.ClusterURL, c.cpdsd.Internal.Global.ControlPlane.URL)
				Expect(err).To(BeNil())
				Expect(isLocal).To(Equal(c.expected))
			})
		}
	})
})
