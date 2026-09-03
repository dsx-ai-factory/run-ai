/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Status Manager", func() {
	Context("Status updater Manager", func() {
		var (
			ctx           context.Context
			statusManager *StatusUpdaterManager
			updater       *mocks.StatusUpdater
		)
		BeforeEach(func() {
			ctx = context.Background()
			updater = &mocks.StatusUpdater{}
			statusManager = NewStatusUpdaterManager(updater, mockIsStatusChanged, mockPreUpdateStatusCallback)
		})
		AfterEach(
			func() {
				ctx = context.Background()
			})

		const ReturnChanged = "returnChanged"
		const ReturnCallback = "returnCallback"
		It("should update status when status is changed", func() {
			ctx = context.WithValue(ctx, ReturnChanged, true)
			ctx = context.WithValue(ctx, ReturnCallback, nil)
			updater.On("UpdateStatus", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
				modifiedWorkload := args.Get(1).(*jobs.JobWorkload)
				modifiedWorkload.ResourceVersion = "Modified"
			})
			modifiedWorkload := &jobs.JobWorkload{

				ObjectMeta: metav1.ObjectMeta{
					Name:            "",
					ResourceVersion: "Original",
				},
			}
			err := statusManager.Update(ctx, modifiedWorkload, &jobs.JobWorkload{})
			Expect(err).To(BeNil())
			Expect(modifiedWorkload.ResourceVersion).To(Equal("Modified"))

		})
		It("should not update status when status is not changed", func() {
			ctx = context.WithValue(ctx, ReturnChanged, false)
			ctx = context.WithValue(ctx, ReturnCallback, nil)
			modifiedWorkload := &jobs.JobWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "",
					ResourceVersion: "Original",
				},
			}
			err := statusManager.Update(ctx, modifiedWorkload, &jobs.JobWorkload{})
			Expect(err).To(BeNil())
			Expect(modifiedWorkload.ResourceVersion).To(Equal("Original"))

		})
		It("should not update status when status is changed but pre update status callback failed", func() {
			ctx = context.WithValue(ctx, ReturnChanged, true)
			ctx = context.WithValue(ctx, ReturnCallback, fmt.Errorf("Failed to run callback"))
			modifiedWorkload := &jobs.JobWorkload{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "",
					ResourceVersion: "Original",
				},
			}
			err := statusManager.Update(ctx, modifiedWorkload, &jobs.JobWorkload{})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("Failed to run callback"))
			Expect(modifiedWorkload.ResourceVersion).To(Equal("Original"))

		})

	})
	Context("UpdateLastChangedStatus", func() {
		var (
			ctx              context.Context
			modifiedWorkload *jobs.JobWorkload
		)
		BeforeEach(func() {
			ctx = context.Background()
			modifiedWorkload = &jobs.JobWorkload{}
		})
		AfterEach(
			func() {
				ctx = context.Background()
			})

		It("should update last status change time", func() {
			err := UpdateLastChangedStatus(ctx, modifiedWorkload)
			Expect(err).To(BeNil())
			Expect(modifiedWorkload.Status.LastStatusChangeTime).ToNot(BeNil())
		})

	})
	Context("SimpleStatusUpdater", func() {
		var (
			ctx              context.Context
			modifiedWorkload *jobs.JobWorkload
			originalWorkload *jobs.JobWorkload
			updaterManager   *StatusUpdaterManager
			updater          *mocks.StatusUpdater
		)
		BeforeEach(func() {
			ctx = context.Background()
			originalWorkload = workloadUtils.NewJobWorkloadBuilder().WithName("original").
				WithTraining().SetActive().WithImage("ubuntu").WithGpuCount(1).Build()
			originalWorkload.ResourceVersion = "Original"
			modifiedWorkload = originalWorkload.DeepCopy()
			updater = &mocks.StatusUpdater{}
			updaterManager = NewSimpleStatusUpdaterManager(updater)
		})
		It("should not update status when status is not changed", func() {
			updater.On("UpdateStatus", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
				modifiedWorkload := args.Get(1).(*jobs.JobWorkload)
				modifiedWorkload.ResourceVersion = "Modified"
			})
			err := updaterManager.Update(ctx, modifiedWorkload, originalWorkload)
			Expect(err).To(BeNil())
			Expect(modifiedWorkload.ResourceVersion).To(Equal("Original"))
			Expect(modifiedWorkload.Status.LastStatusChangeTime).To(BeNil())
		})
		It("should update status when status is changed", func() {
			modifiedWorkload.Status.WorkloadCommonStatus.CompletionTime = &metav1.Time{Time: metav1.Now().Add(1 * time.Minute)}
			updater.On("UpdateStatus", ctx, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
				modifiedWorkload := args.Get(1).(*jobs.JobWorkload)
				modifiedWorkload.ResourceVersion = "Modified"
			})
			err := updaterManager.Update(ctx, modifiedWorkload, originalWorkload)
			Expect(err).To(BeNil())
			Expect(modifiedWorkload.ResourceVersion).To(Equal("Modified"))
			Expect(modifiedWorkload.Status.LastStatusChangeTime).ToNot(BeNil())
		})

	})
	Context("IsStatusChangedInWorkload", func() {
		var (
			ctx              context.Context
			modifiedWorkload *jobs.JobWorkload
			originalWorkload *jobs.JobWorkload
		)
		BeforeEach(func() {
			ctx = context.Background()
			modifiedWorkload = &jobs.JobWorkload{}
			originalWorkload = &jobs.JobWorkload{}
		})
		AfterEach(
			func() {
				ctx = context.Background()
			})

		It("should return true when status is changed", func() {
			modifiedWorkload.Status.WorkloadCommonStatus.CompletionTime = &metav1.Time{}
			Expect(IsStatusChangedInWorkload(ctx, modifiedWorkload, originalWorkload)).To(BeTrue())
		})
		It("should return false when status is not changed", func() {
			Expect(IsStatusChangedInWorkload(ctx, modifiedWorkload, originalWorkload)).To(BeFalse())
		})
	})
})

func mockIsStatusChanged(ctx context.Context, _, _ *jobs.JobWorkload) bool {

	value := ctx.Value("returnChanged")
	if value == nil {
		return false
	}
	return value.(bool)
}

func mockPreUpdateStatusCallback(ctx context.Context, _ *jobs.JobWorkload) error {
	value := ctx.Value("returnCallback")
	if value == nil {
		return nil
	}
	return value.(error)
}
