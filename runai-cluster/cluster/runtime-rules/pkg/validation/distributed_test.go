/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/controller"
	kfv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation Test Suite", func() {
	var (
		validator *WorkloadValidator
		iw, dw    *jobs.JobWorkload
	)

	BeforeEach(func() {
		validator = NewWorkloadValidator(NewDistributedValidator(getTestFrameworksStatus()))
		jobStartTime := &metav1.Time{Time: time.Now()}
		iw = &jobs.JobWorkload{
			TypeMeta:   metav1.TypeMeta{Kind: constants.KindInteractiveWorkload},
			ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "iw"},
		}
		iw.Status.StartTime = jobStartTime
		dw = &jobs.JobWorkload{
			TypeMeta:   metav1.TypeMeta{Kind: constants.KindDistributedWorkload},
			ObjectMeta: metav1.ObjectMeta{Namespace: "namespace", Name: "dw"},
		}
		dw.Status.StartTime = jobStartTime
	})

	Context("Distributed", func() {
		Context("validation success", func() {
			It("not distributed", func() {
				Expect(validator.Validate(iw)).To(BeNil())
			})

			It("framework enabled and suspendable", func() {
				dw.Spec.JobType = constants.KindTFJob
				Expect(validator.Validate(dw)).To(BeNil())
			})
		})
		Context("validation failure", func() {
			It("framework enabled and not suspendable", func() {
				dw.Spec.JobType = constants.KindPyTorchJob
				err := validator.Validate(dw)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Cluster prerequisites were not met for %s framework. Contact your administrator for help", dw.Spec.JobType)))
			})

			It("framework not enabled", func() {
				dw.Spec.JobType = constants.KindMpiJob
				err := validator.Validate(dw)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Cluster prerequisites were not met for %s framework. Contact your administrator for help", dw.Spec.JobType)))
			})
		})

	})
})

func getTestFrameworksStatus() map[string]controller.DistributedJobFrameworksStatus {
	enabledFrameworks := make(map[string]controller.DistributedJobFrameworksStatus)
	enabledFrameworks[constants.KindTFJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.TFJob{}, SuspendEnabled: true}
	enabledFrameworks[constants.KindPyTorchJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.PyTorchJob{}, SuspendEnabled: false}

	return enabledFrameworks
}
