/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	iw "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/finalizers"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/status/update"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	expectedUpsertErrorMsg = "upsert workload resources error"
)

var _ = Describe("Issue Structure Verification", func() {
	const jobName = "test-job1"
	var (
		fakeClientPtr         *client.FakeClientWithErrorRules
		k8sClient             k8sclient.Client
		reconcile             *JobController
		ctx                   context.Context
		project               *runaiv2.Project
		namespaceObj          *corev1.Namespace
		statusUpdater         update.StatusUpdater
		defaultPodLabels      map[string]string
		defaultPodAnnotations map[string]string
	)
	scheme := runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		project = resourceUtils.CreateProject(projectName)
		kaiProject := resourceUtils.CreateKaiProject(projectName)
		namespaceObj = resourceUtils.CreateNamespace(namespace, projectName)

		fakeClientPtr = client.NewFakeClientWithErrorRules(scheme, []k8sclient.Object{&runaiJobV1.RunaiJob{}, &iw.InteractiveWorkload{}}, nil, project, kaiProject, namespaceObj)
		k8sClient = fakeClientPtr

		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
		recorder := record.NewFakeRecorder(200)

		statusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.InteractiveWorkload)
		Expect(err).ToNot(HaveOccurred())
		statusUpdater = statusUpdaterFactory.CreateStatusUpdater(k8sClient)
		finalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.InteractiveWorkload)
		Expect(err).ToNot(HaveOccurred())
		reconcile = CreateJobController(k8sClient, statusUpdater,
			finalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
			defaultPodLabels, defaultPodAnnotations)
	})

	Context("Workspace building with no issue", Ordered, func() {
		BeforeEach(func() {
			interactiveWorkload := createInteractiveJobWithPorts(jobName, k8sClient, ctx)
			jobWorkload := &job.JobWorkload{}
			interactiveWorkload.CopyTo(jobWorkload)

			Expect(interactiveWorkload.Status.Phase).To(Equal(""))
			// first reconcile
			_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Duplicate ingress issues found", func() {
			rules := []client.ResourceFailingRule{
				{
					ResourceKind: constants.KindIngress,
					ResourceName: "iw-test-job2-0",
					StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: client.UrlAlreadyExistsMsg, Code: http.StatusBadRequest}},
				},
			}
			fakeClientPtr.UpdateRules(rules)

			jobName2 := "test-job2"
			dupWorkload := createInteractiveJobWithPorts(jobName2, k8sClient, ctx)
			dupJob := &job.JobWorkload{}
			dupWorkload.CopyTo(dupJob)

			_, err := reconcile.Reconcile(ctx, ctrl.Request{}, dupJob, dupWorkload)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedUpsertErrorMsg))

			first := getInteractiveJob(ctx, k8sClient, jobName, namespace)
			Expect(first).ToNot(BeNil())
			second := getInteractiveJob(ctx, k8sClient, jobName2, namespace)
			Expect(second).ToNot(BeNil())

			// first job has owned resources and issues list is nil
			Expect(first.Status.OwnedResources).To(HaveLen(3))
			Expect(first.Status.Issues).To(BeNil())
			// second job owned resource is empty, and issues contain a network issue
			Expect(second.Status.OwnedResources).To(HaveLen(0))
			Expect(second.Status.Issues).ToNot(BeNil())
			Expect(second.Status.Issues.Network).To(HaveLen(1))
			Expect(second.Status.Issues.Network[0].Description).To(ContainSubstring("is already defined in ingress"))

			// reconcile once again
			_, err = reconcile.Reconcile(ctx, ctrl.Request{}, dupJob, dupWorkload)
			Expect(err).ToNot(HaveOccurred())

			// second job owned resource is empty, and issues contain a network issue - issues is kept
			second = getInteractiveJob(ctx, k8sClient, jobName2, namespace)
			Expect(second).ToNot(BeNil())
			Expect(second.Status.OwnedResources).To(HaveLen(0))
			Expect(second.Status.Issues).ToNot(BeNil())
			Expect(second.Status.Issues.Network).To(HaveLen(1))
			Expect(second.Status.Issues.Network[0].Description).To(ContainSubstring("is already defined in ingress"))
		})

		It("General ingress issue found", func() {
			rules := []client.ResourceFailingRule{
				{
					ResourceKind: constants.KindIngress,
					ResourceName: "iw-test-random-job2-0",
				},
			}
			fakeClientPtr.UpdateRules(rules)

			randomName := "test-random-job2"
			randomWorkload := createInteractiveJobWithPorts(randomName, k8sClient, ctx)
			job := &job.JobWorkload{}
			randomWorkload.CopyTo(job)

			_, err := reconcile.Reconcile(ctx, ctrl.Request{}, job, randomWorkload)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(expectedUpsertErrorMsg))

			tmp := getInteractiveJob(ctx, k8sClient, randomName, namespace)
			Expect(tmp).ToNot(BeNil())
			Expect(tmp.Status.Phase).To(Equal(common.PhaseFailed.String()))
			Expect(tmp.Status.PhaseMessage).To(Equal("Workload failed due to a Network issue"))
			Expect(tmp.Status.Issues).ToNot(BeNil())
			Expect(tmp.Status.Issues.Network).To(HaveLen(1))
			Expect(tmp.Status.Issues.Network[0].Kind).To(Equal(constants.KindIngress))
			Expect(tmp.Status.Issues.Network[0].Description).To(ContainSubstring(fmt.Sprintf(client.GeneralErrorMsg, "")))
			Expect(tmp.Status.Issues.Network[0].Description).To(ContainSubstring(randomName))
		})
	})
})
