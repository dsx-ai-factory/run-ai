/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	resourcev1 "k8s.io/api/resource/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
)

var _ = Describe("ParseDeviceClassToServerScheme", func() {
	buildDeviceClass := func(name string) *resourcev1.DeviceClass {
		return &resourcev1.DeviceClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		}
	}

	It("maps the device class name onto the tracker payload", func() {
		result := parsers.ParseDeviceClassToServerScheme(buildDeviceClass("gpu.nvidia.com"))
		Expect(result.DeviceClassName).To(Equal("gpu.nvidia.com"))
	})

	It("sets the cluster id from the configured cluster uuid", func() {
		result := parsers.ParseDeviceClassToServerScheme(buildDeviceClass("mig.nvidia.com"))
		Expect(result.ClusterId).To(HaveValue(Equal(config.Config().ClusterUuid)))
	})

	It("ignores everything on the device class except the name", func() {
		dc := buildDeviceClass("compute-domain-default-channel.nvidia.com")
		dc.Labels = map[string]string{"unused": "value"}
		dc.Annotations = map[string]string{"unused": "value"}
		dc.Spec = resourcev1.DeviceClassSpec{
			Selectors: []resourcev1.DeviceSelector{{}},
		}

		result := parsers.ParseDeviceClassToServerScheme(dc)
		Expect(result.DeviceClassName).To(Equal("compute-domain-default-channel.nvidia.com"))
		Expect(result.ClusterId).To(HaveValue(Equal(config.Config().ClusterUuid)))
	})
})
