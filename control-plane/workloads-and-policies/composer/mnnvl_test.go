/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func mnnvlLabel(name, value string, exclude bool) workloadsApi.Label {
	label := workloadsApi.Label{
		Name:  *workloadsApi.NewNullableString(convutils.Ref(name)),
		Value: *workloadsApi.NewNullableString(convutils.Ref(value)),
	}
	if exclude {
		label.Exclude = *workloadsApi.NewNullableBool(convutils.Ref(true))
	}
	return label
}

var _ = Describe("FieldsComposer.ComposeMNNVL", func() {
	fc := &FieldsComposer{}

	DescribeTable("writes the resolved MNNVL mode into the target configuration",
		func(labels []workloadsApi.Label, defaultMode string, expectSet bool, expected workloadsManagerApi.MnnvlMode) {
			config := workloadsManagerApi.WorkloadV2Configuration{}
			err := fc.ComposeMNNVL(labels, defaultMode, &config, MNNVLToWorkloadRunner, nil)()
			Expect(err).To(BeNil())
			Expect(config.Mnnvl.IsSet()).To(Equal(expectSet))
			if expectSet {
				Expect(config.GetMnnvl()).To(Equal(expected))
			}
		},
		Entry("no label + empty default leaves the field unset",
			nil, "", false, workloadsManagerApi.MnnvlMode("")),
		Entry("no label + None default emits None",
			nil, mnnvl.None, true, workloadsManagerApi.MNNVLMODE_NONE),
		Entry("no label + Required default emits Required",
			nil, mnnvl.Required, true, workloadsManagerApi.MNNVLMODE_REQUIRED),
		Entry("Required label emits Required",
			[]workloadsApi.Label{mnnvlLabel(mnnvl.FeatureLabelKey, mnnvl.Required, false)}, "", true, workloadsManagerApi.MNNVLMODE_REQUIRED),
		Entry("None label emits None",
			[]workloadsApi.Label{mnnvlLabel(mnnvl.FeatureLabelKey, mnnvl.None, false)}, "", true, workloadsManagerApi.MNNVLMODE_NONE),
		Entry("Preferred label is normalized to Required",
			[]workloadsApi.Label{mnnvlLabel(mnnvl.FeatureLabelKey, mnnvl.Preferred, false)}, "", true, workloadsManagerApi.MNNVLMODE_REQUIRED),
		Entry("excluded MNNVL label is ignored and the default applies",
			[]workloadsApi.Label{mnnvlLabel(mnnvl.FeatureLabelKey, mnnvl.Required, true)}, mnnvl.None, true, workloadsManagerApi.MNNVLMODE_NONE),
		Entry("non-MNNVL label is ignored and the default applies",
			[]workloadsApi.Label{mnnvlLabel("run.ai/other", mnnvl.Required, false)}, "", false, workloadsManagerApi.MnnvlMode("")),
	)

	It("rejects invalid MNNVL label values", func() {
		config := workloadsManagerApi.WorkloadV2Configuration{}
		labels := []workloadsApi.Label{mnnvlLabel(mnnvl.FeatureLabelKey, "bogus", false)}
		err := fc.ComposeMNNVL(labels, "", &config, MNNVLToWorkloadRunner, nil)()
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("Invalid MNNVL mode 'bogus'"))
		Expect(config.Mnnvl.IsSet()).To(BeFalse())
	})
})
