/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
)

var _ = Describe("Utility functions", func() {
	Context("verify itemExist function for remoteWrite", func() {
		It("should return true if the item exists in the list", func() {
			remoteWrites := []manifests.RemoteWriteSpec{
				{
					URL:  "http://localhost:9090",
					Name: "test1",
				},
				{
					URL:  "http://localhost:9091",
					Name: "test2",
				},
			}
			remoteWrite := manifests.RemoteWriteSpec{
				URL:  "http://localhost:9090",
				Name: "test1",
			}
			exist, index := findRemoteWrite(remoteWrites, remoteWrite)
			Expect(exist).To(BeTrue())
			Expect(index).To(Equal(0))
		})
		It("should return false if the item does not exist in the list", func() {
			remoteWrites := []manifests.RemoteWriteSpec{
				{
					URL:  "http://localhost:9090",
					Name: "test1",
				},
				{
					URL:  "http://localhost:9091",
					Name: "test2",
				},
			}
			remoteWrite := manifests.RemoteWriteSpec{
				URL:  "http://localhost:9092",
				Name: "test3",
			}
			exist, index := findRemoteWrite(remoteWrites, remoteWrite)
			Expect(exist).To(BeFalse())
			Expect(index).To(Equal(-1))
		})
	})
})
