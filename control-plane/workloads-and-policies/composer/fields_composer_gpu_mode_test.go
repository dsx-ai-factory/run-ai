/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"context"
	"net/http"
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPU compute mode", func() {
	DescribeTable("validates cluster support only for explicit GPU modes",
		func(mode *string, devices int32, available bool, expectedCode int32, emitted bool) {
			fc := FieldsComposer{Options: FieldsComposerOptions{GpuComputeModeSupported: available}}
			annotations := map[string]string{"unrelated": "preserved"}
			err := fc.ComposeGpuComputeMode(mode, &GpuSizingInfo{NumDevices: &devices}, "actual-container", &annotations, StringToMapString)()
			if expectedCode == 0 {
				Expect(err).To(BeNil())
			} else {
				Expect(err).NotTo(BeNil())
				Expect(err.Code).To(Equal(expectedCode))
			}
			Expect(annotations).To(HaveKeyWithValue("unrelated", "preserved"))
			if emitted {
				Expect(annotations).To(HaveKeyWithValue("nvidia.com/container.actual-container.gpu-compute.mode", *mode))
			} else {
				Expect(annotations).To(HaveLen(1))
			}
		},
		Entry("legacy omission on old cluster", nil, int32(1), false, int32(0), false),
		Entry("CPU zero devices", convutils.Ref("sm-sharing"), int32(0), false, int32(0), false),
		Entry("explicit time slicing", convutils.Ref("time-slicing"), int32(1), true, int32(0), true),
		Entry("SM sharing on multiple devices", convutils.Ref("sm-sharing"), int32(2), true, int32(0), true),
		Entry("time slicing on old cluster", convutils.Ref("time-slicing"), int32(1), false, int32(http.StatusBadRequest), false),
		Entry("SM sharing on old cluster", convutils.Ref("sm-sharing"), int32(1), false, int32(http.StatusBadRequest), false),
		Entry("unknown enum", convutils.Ref("unsupported"), int32(1), true, int32(400), false),
	)
	DescribeTable("validates sizing before evaluating the compute mode",
		func(memory *string, devices int32, code int32, emitted bool) {
			fc := FieldsComposer{Options: FieldsComposerOptions{GpuComputeModeSupported: true}}
			var sizing GpuSizingInfo
			annotations := map[string]string{}
			runner := &StepRunner{}
			runner.Batch(
				fc.ComposeGpuRequestsAndLimits(context.Background(), &devices, nil, nil, nil, memory, nil,
					func(*GpuSizingInfo, any, any) *workloadsApi.Error { return nil }, nil, nil, &sizing),
				fc.ComposeGpuComputeMode(convutils.Ref("time-slicing"), &sizing, "main", &annotations, StringToMapString),
			)
			if code == 0 {
				Expect(runner.Error()).To(BeNil())
			} else {
				Expect(runner.Error()).NotTo(BeNil())
				Expect(runner.Error().Code).To(Equal(code))
			}
			Expect(len(annotations) > 0).To(Equal(emitted))
		},
		Entry("zero memory is rejected before mode composition", convutils.Ref("0"), int32(0), int32(http.StatusBadRequest), false),
		Entry("zero devices stays CPU-only", nil, int32(0), int32(0), false),
		Entry("positive memory emits the mode", convutils.Ref("1Gi"), int32(1), int32(0), true),
	)

	It("validates partial policy defaults without requiring a running cluster", func() {
		fc := FieldsComposer{Options: OptionsForPolicy()}
		var annotations map[string]string
		Expect(fc.ComposeGpuComputeMode(convutils.Ref("sm-sharing"), nil, "main", &annotations, StringToMapString)()).To(BeNil())
	})
	It("rejects an annotation key exceeding the Kubernetes limit", func() {
		fc := FieldsComposer{Options: OptionsForPolicy()}
		var annotations map[string]string
		err := fc.ComposeGpuComputeMode(convutils.Ref("sm-sharing"), nil, strings.Repeat("a", 50), &annotations, StringToMapString)()
		Expect(err.Code).To(Equal(int32(400)))
	})
})
