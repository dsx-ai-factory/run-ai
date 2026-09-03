/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaijob

import (
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("projectLimits extractors", func() {
	const (
		interactiveTimeLimit   = int64(120)
		interactiveIdle        = int64(1200)
		interactivePreemptIdle = int64(1800)
		trainingTimeLimit      = int64(90)
		trainingIdle           = int64(2400)
	)

	expected := projectLimits{
		interactiveTimeLimitSecs:          interactiveTimeLimit,
		interactiveMaxIdleDurationSecs:    interactiveIdle,
		interactivePreemptibleMaxIdleSecs: interactivePreemptIdle,
		trainingTimeLimitSecs:             trainingTimeLimit,
		trainingMaxIdleDurationSecs:       trainingIdle,
	}

	It("reads the limits from the kai project annotations", func() {
		kaiProject := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.InteractiveJobTimeLimitSecondsAnnotation:                  strconv.FormatInt(interactiveTimeLimit, 10),
					constants.InteractiveJobMaxIdleDurationSecondsAnnotation:            strconv.FormatInt(interactiveIdle, 10),
					constants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation: strconv.FormatInt(interactivePreemptIdle, 10),
					constants.TrainingJobTimeLimitSecondsAnnotation:                     strconv.FormatInt(trainingTimeLimit, 10),
					constants.TrainingJobMaxIdleDurationSecondsAnnotation:               strconv.FormatInt(trainingIdle, 10),
				},
			},
		}
		Expect(limitsFromProject(kaiProject)).To(Equal(expected))
	})

	It("treats missing kai annotations as zero limits", func() {
		Expect(limitsFromProject(&kaiorg.Project{})).To(Equal(projectLimits{}))
	})

	It("treats an unparseable kai annotation as zero", func() {
		kaiProject := &kaiorg.Project{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.TrainingJobTimeLimitSecondsAnnotation: "not-a-number",
				},
			},
		}
		Expect(limitsFromProject(kaiProject)).To(Equal(projectLimits{}))
	})
})
