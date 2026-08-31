/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	admission_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/admission-controller"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/binder"
	device_plugin "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/device-plugin"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/engine"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runai_container_toolkit "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-container-toolkit"
	runai_container_toolkit_reservation "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-container-toolkit-reservation"
	runai_node_exporter "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-node-exporter"
)

var _ = Describe("Deprecated RunaiConfig compatibility data", func() {
	It("survives defaulting, deep copy, and JSON round trips without being changed", func() {
		eventMask := []global.ToolkitEventReason{"gpuOOMKilled", "nodeScheduler"}
		legacy := RunaiConfigSpec{
			Engine: &engine.Engine{Enabled: ptr.To(false)},
			RunAIContainerToolkit: &runai_container_toolkit.RunAIContainerToolkit{
				RunAISharedDirectoryHostPath: ptr.To("/legacy/shared"),
				Metrics: &runai_container_toolkit.Metrics{
					Enabled:            ptr.To(true),
					CollectionInterval: ptr.To(17),
				},
			},
			RunAIContainerToolkitReservation: &runai_container_toolkit_reservation.RunAIContainerToolkitReservation{
				Enabled: ptr.To(false),
			},
			DevicePlugin: &device_plugin.DevicePlugin{
				Enabled: ptr.To(false),
				Args:    &device_plugin.Args{Slices: ptr.To(321)},
			},
			NodeExporter: &runai_node_exporter.NodeExporter{
				Enabled:                    ptr.To(false),
				RunAIProcDirectoryHostPath: "/legacy/proc",
			},
			Global: &global.Global{
				RunaiContainerRuntimeEnabled: ptr.To(false),
				DevicePluginBindings:         ptr.To(true),
				Core: &global.Core{
					ToolkitEvents: &global.ToolkitEvents{Enabled: ptr.To(true), Mask: &eventMask},
				},
			},
			AdmissionController: &admission_controller.AdmissionController{
				Args: &admission_controller.Args{
					RunAISharedDirectoryHostPath:                 "/legacy/shared",
					RunAIFractionalMinAllocationEnforcementBytes: 1234,
					RunAIMpsDirectoryHostPath:                    "/legacy/mps",
				},
			},
			Binder: &binder.Binder{
				Args: &binder.Args{
					RunAIFractionalMinAllocationEnforcementBytes: ptr.To(5678),
					RunAIMpsDirectoryHostPath:                    ptr.To("/legacy/mps"),
					NodeLevelScheduler:                           ptr.To(true),
					InjectRunaiResource:                          ptr.To(false),
				},
			},
		}
		config := &RunaiConfig{Spec: legacy}
		expected := config.DeepCopy()

		config.SetDefaultsWhereNeeded()
		Expect(config.Spec.Engine).To(Equal(expected.Spec.Engine))
		Expect(config.Spec.RunAIContainerToolkit).To(Equal(expected.Spec.RunAIContainerToolkit))
		Expect(config.Spec.RunAIContainerToolkitReservation).To(Equal(expected.Spec.RunAIContainerToolkitReservation))
		Expect(config.Spec.DevicePlugin).To(Equal(expected.Spec.DevicePlugin))
		Expect(config.Spec.NodeExporter).To(Equal(expected.Spec.NodeExporter))
		Expect(config.Spec.Global.Core).To(Equal(expected.Spec.Global.Core))
		Expect(config.Spec.Global.RunaiContainerRuntimeEnabled).To(Equal(expected.Spec.Global.RunaiContainerRuntimeEnabled))
		Expect(config.Spec.Global.DevicePluginBindings).To(Equal(expected.Spec.Global.DevicePluginBindings))
		Expect(config.Spec.AdmissionController.Args).To(Equal(expected.Spec.AdmissionController.Args))
		Expect(config.Spec.Binder.Args.RunAIFractionalMinAllocationEnforcementBytes).
			To(Equal(expected.Spec.Binder.Args.RunAIFractionalMinAllocationEnforcementBytes))
		Expect(config.Spec.Binder.Args.RunAIMpsDirectoryHostPath).
			To(Equal(expected.Spec.Binder.Args.RunAIMpsDirectoryHostPath))
		Expect(config.Spec.Binder.Args.NodeLevelScheduler).To(Equal(expected.Spec.Binder.Args.NodeLevelScheduler))
		Expect(config.Spec.Binder.Args.InjectRunaiResource).To(Equal(expected.Spec.Binder.Args.InjectRunaiResource))

		serialized, err := json.Marshal(config)
		Expect(err).ToNot(HaveOccurred())
		roundTripped := &RunaiConfig{}
		Expect(json.Unmarshal(serialized, roundTripped)).To(Succeed())
		Expect(roundTripped.Spec.Engine).To(Equal(expected.Spec.Engine))
		Expect(roundTripped.Spec.RunAIContainerToolkit).To(Equal(expected.Spec.RunAIContainerToolkit))
		Expect(roundTripped.Spec.RunAIContainerToolkitReservation).To(Equal(expected.Spec.RunAIContainerToolkitReservation))
		Expect(roundTripped.Spec.DevicePlugin).To(Equal(expected.Spec.DevicePlugin))
		Expect(roundTripped.Spec.NodeExporter).To(Equal(expected.Spec.NodeExporter))
		Expect(roundTripped.Spec.Global.Core).To(Equal(expected.Spec.Global.Core))
		Expect(roundTripped.Spec.Global.RunaiContainerRuntimeEnabled).
			To(Equal(expected.Spec.Global.RunaiContainerRuntimeEnabled))
		Expect(roundTripped.Spec.Global.DevicePluginBindings).To(Equal(expected.Spec.Global.DevicePluginBindings))
		Expect(roundTripped.Spec.AdmissionController.Args).To(Equal(expected.Spec.AdmissionController.Args))
		Expect(roundTripped.Spec.Binder.Args.RunAIFractionalMinAllocationEnforcementBytes).
			To(Equal(expected.Spec.Binder.Args.RunAIFractionalMinAllocationEnforcementBytes))
		Expect(roundTripped.Spec.Binder.Args.RunAIMpsDirectoryHostPath).
			To(Equal(expected.Spec.Binder.Args.RunAIMpsDirectoryHostPath))
		Expect(roundTripped.Spec.Binder.Args.NodeLevelScheduler).To(Equal(expected.Spec.Binder.Args.NodeLevelScheduler))
		Expect(roundTripped.Spec.Binder.Args.InjectRunaiResource).To(Equal(expected.Spec.Binder.Args.InjectRunaiResource))
	})
})
