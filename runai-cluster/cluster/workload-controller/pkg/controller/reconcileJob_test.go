/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	v1alpha12 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	runaiv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	fields "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	controllerutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/controller-utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	workloadresource "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/finalizers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	resourceUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	job "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	common2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	issue "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	jobBuilder "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/jobs/runaijob"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/managers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/aggregator"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/status/update"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	twv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	projectName  = "test-project"
	namespace    = "runai-test-project"
	nodepoolName = "test-nodepool"
	pvcSize      = "1Gi"
	jobFailedMsg = "I failed"
)

var _ = Describe("Reconcile runai jobs workloads", func() {
	const jobName = "test-job"
	var (
		k8sClient         k8sclient.Client
		fakeClientPtr     *client.FakeClientWithErrorRules
		reconcile         *JobController
		ctx               context.Context
		nodepool          *v1alpha12.NodePool
		nodepoolKai       *kaiorg.NodePool
		project           *runaiv2.Project
		kaiProject        *kaiorg.Project
		namespaceObj      *v1.Namespace
		recorder          *record.FakeRecorder
		defaultPodLabels  map[string]string
		defaultAnnotation map[string]string
	)
	scheme := runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		project = resourceUtils.CreateProject(projectName)
		kaiProject = resourceUtils.CreateKaiProject(projectName)
		namespaceObj = resourceUtils.CreateNamespace(namespace, projectName)

		// Register both the legacy run.ai project and the kai.resources project so the
		// reconcile flow finds its fixture.
		trackedObjects := []k8sclient.Object{
			&runaiJobV1.RunaiJob{}, &iw.InteractiveWorkload{}, &twv1.TrainingWorkload{},
			&v2alpha1.DistributedWorkload{}, &v1.Service{},
		}
		fakeClientPtr = client.NewFakeClientWithErrorRules(scheme, trackedObjects, nil, project, kaiProject, namespaceObj)
		k8sClient = fakeClientPtr

		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
		recorder = record.NewFakeRecorder(200)
	})
	Context("Training Workload", func() {
		var (
			trainingWorkload *twv1.TrainingWorkload
			jobWorkload      *job.JobWorkload
			statusUpdater    update.StatusUpdater
		)
		BeforeEach(func() {
			statusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.TrainingWorkload)
			Expect(err).ToNot(HaveOccurred())
			statusUpdater = statusUpdaterFactory.CreateStatusUpdater(k8sClient)
			finalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.TrainingWorkload)
			Expect(err).ToNot(HaveOccurred())
			reconcile = CreateJobController(k8sClient, statusUpdater,
				finalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder, defaultPodLabels, defaultAnnotation)
		})
		Context("Sanity test", func() {
			Context("Job is Mark for deletion", func() {
				Context("backward compatibility ", func() {

					BeforeEach(func() {

						jobWorkload := workload.NewJobWorkloadBuilder().WithTraining().WithNamespace(namespace).WithImage("test").WithName(jobName).Build()
						jobWorkload.DeletionTimestamp = &metav1.Time{Time: time.Now()}
						trainingWorkload = twv1.NewFromJobWorkload(jobWorkload)
						controllerutils.AddFinalizer(trainingWorkload, workloadCommon.FinalizerName)
					})
				})
			})
			Context("First workload reconcile", func() {

				BeforeEach(func() {
					builder := workload.NewJobWorkloadBuilder()
					builder.WithName(jobName).WithNamespace(namespace).WithTraining()
					jobWorkload = builder.WithImage("test-image").WithCommand("test-command").Build()
					trainingWorkload = twv1.NewFromJobWorkload(jobWorkload)
					pvcs := map[string]fields.PvcsItemAndRules{}
					pvcs["pvc1"] = fields.PvcsItemAndRules{
						Value: &fields.PvcsItemValue{
							ClaimName: ptr.To("regularclaimname"),
							Size:      ptr.To("1Gi"),
							Path:      ptr.To("/path"),
							ReadOnly:  ptr.To(true),
							Ephemeral: ptr.To(false),
						},
					}
					pvcs["pvc2"] = fields.PvcsItemAndRules{
						Value: &fields.PvcsItemValue{
							ClaimName: ptr.To("ephemeralClaimName"),
							Size:      ptr.To("1Gi"),
							Path:      ptr.To("/path/cyber"),
							ReadOnly:  ptr.To(true),
							Ephemeral: ptr.To(true),
						},
					}
					trainingWorkload.Spec.Pvcs = fields.NewPvcsField(pvcs)
					err := k8sClient.Create(ctx, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Should set creationStatus to submitted", func() {
					tmp := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmp.Status.CreationStatus).To(Equal(common.StatusSubmitted))
				})
				It("Should set first Ack a job", func() {
					tmp := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmp.Status.FirstAcknowledgedTime).ToNot(BeNil())
				})
				It("All Condition should be false", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Ready, "")
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Succeeded, "")
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Failed, "")
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Suspended, "")
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Degraded, "")
				})
				Context("Backward compatibility - MPI job", func() {
					const MPIJobName = "test-mpi-job"
					BeforeEach(func() {
						jobWorkload = workload.NewJobWorkloadBuilder().WithTraining().WithNamespace(namespace).WithOldMPI().WithName(MPIJobName).Build()
						trainingWorkload = twv1.NewFromJobWorkload(jobWorkload)
						err := k8sClient.Create(ctx, trainingWorkload)
						Expect(err).ToNot(HaveOccurred())
						trainingWorkload.CopyTo(jobWorkload)
						_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
						Expect(err).ToNot(HaveOccurred())
					})
					It("Should skip reconcile", func() {
						trainingWorkload = getTrainingJob(ctx, k8sClient, MPIJobName, namespace)
						Expect(trainingWorkload.Status.CreatedResources).To(HaveLen(0))
						Expect(trainingWorkload.Status.FirstAcknowledgedTime.IsZero()).To(BeTrue())
					})

				})

			})
			Context("Init context", func() {
				BeforeEach(func() {
					builder := workload.NewJobWorkloadBuilder()
					builder.WithName(jobName).WithNamespace(namespace).WithTraining()
					jobWorkload = builder.WithImage("test-image").WithCommand("test-command").Build()
					trainingWorkload = twv1.NewFromJobWorkload(jobWorkload)
					err := k8sClient.Create(ctx, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					trainingWorkload.CopyTo(jobWorkload)
				})

				It("does not fail on mnnvl", func() {
					jobWorkload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Required

					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("Min required ready replica", func() {
				BeforeEach(func() {
					jobWorkload = workload.NewJobWorkloadBuilder().WithTraining().WithNamespace(namespace).WithImage("test-image").WithName(jobName).Build()
					trainingWorkload = twv1.NewFromJobWorkload(jobWorkload)
					err := k8sClient.Create(ctx, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Should set min required ready replica", func() {
					tmp := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmp.Status.MinRequiredReadyPods).To(Equal(int32(1)))
				})
			})
			Context("Job uses NodePool", func() {

				BeforeEach(func() {
					nodepool = resourceUtils.CreateNodePool(nodepoolName, "key", "value")
					Expect(k8sClient.Create(ctx, nodepool)).To(Succeed())
					nodepoolKai = resourceUtils.CreateKaiNodePool(nodepoolName, "key", "value")
					Expect(k8sClient.Create(ctx, nodepoolKai)).To(Succeed())

					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					trainingWorkload.Spec.NodePools = fields.NewStringField(nodepoolName)
					Expect(k8sClient.Update(ctx, trainingWorkload)).To(Succeed())

					jobWorkload := &job.JobWorkload{}
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

				})
				It("runai job should have node affinity", func() {
					checkRunaiNodeAffinity(ctx, jobName, namespace, "key", "value", k8sClient)
				})
				It("runai job should have the same node affinity even if nodepool changed", func() {

					nodepool.Spec.LabelKey = "key2"
					nodepool.Spec.LabelValue = "value2"
					Expect(k8sClient.Update(ctx, nodepool)).To(Succeed())

					// Mirror the mutation onto the kai twin, so this spec still exercises an actual
					// pool change under both values of the compile-time UseKaiNodePools flag.
					// TODO: drop once constants.UseKaiNodePools=true and the legacy fixture goes
					nodepoolKai.Spec.LabelKey = "key2"
					nodepoolKai.Spec.LabelValue = "value2"
					Expect(k8sClient.Update(ctx, nodepoolKai)).To(Succeed())
					trainingWorkload = getTrainingJob(ctx, k8sClient, jobName, namespace)
					jobWorkload := &job.JobWorkload{}
					trainingWorkload.CopyTo(jobWorkload)

					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					checkRunaiNodeAffinity(ctx, jobName, namespace, "key", "value", k8sClient)
				})
				Context("Upgrading workload without status", func() {
					BeforeEach(func() {
						trainingWorkload = getTrainingJob(ctx, k8sClient, jobName, namespace)
						trainingWorkload.Status = twv1.TrainingWorkloadStatus{}
						Expect(k8sClient.Update(ctx, trainingWorkload)).To(Succeed())
						jobWorkload := &job.JobWorkload{}
						trainingWorkload.CopyTo(jobWorkload)
						// updating node pool
						nodepool.Spec.LabelKey = "key3"
						nodepool.Spec.LabelValue = "value3"
						Expect(k8sClient.Update(ctx, nodepool)).To(Succeed())

						// Mirror the mutation onto the kai twin, so this spec still exercises an actual
						// pool change under both values of the compile-time UseKaiNodePools flag.
						// TODO: drop once constants.UseKaiNodePools=true and the legacy fixture goes
						nodepoolKai.Spec.LabelKey = "key3"
						nodepoolKai.Spec.LabelValue = "value3"
						Expect(k8sClient.Update(ctx, nodepoolKai)).To(Succeed())
						// first reconcile
						_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
						Expect(err).ToNot(HaveOccurred())

					})
					It("runai job should have node affinity", func() {
						checkRunaiNodeAffinity(ctx, jobName, namespace, "key", "value", k8sClient)
					})
				})

			})
			Context("Job is Suspended", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job suspended

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, true, 0, 0, 0, nil)
					trainingWorkload := getTrainingJob(ctx, k8sClient, jobName, namespace)
					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Suspended condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Suspended)
				})
				It("Pending condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Initialized, "")
				})
				It("Number of running pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.RunningPods).To(Equal(int32(0)))
				})
				It("Number of pending pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.SucceededPods).To(Equal(int32(0)))
				})
				It("Number of failed pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.FailedPods).To(Equal(int32(0)))
				})
				It("CompletionTime should be nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.CompletionTime).To(BeNil())
				})
				It("StartTime should be nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.StartTime).To(BeNil())
				})
				It("Ready condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Ready, "job suspended")
				})
				It("Succeeded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Succeeded, "")
				})
				It("Failed condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Failed, "")
				})
				It("Degraded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Degraded, "")
				})
			})
			// in case the annotation does not update yey
			Context("Job is half Completed", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job completed half completed

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, false, 1, 0, 2, nil)
					trainingWorkload := getTrainingJob(ctx, k8sClient, jobName, namespace)
					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Succeeded condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Succeeded)
				})
				It("Pending condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Initialized)
				})
				It("Failed condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Failed, "")
				})
				It("Suspended condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Suspended, "")
				})
				It("CompletionTime should be set", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.CompletionTime).ToNot(BeNil())
				})
				It("Number of running pods should be 1", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.RunningPods).To(Equal(int32(1)))
				})
				It("Number of pending pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be at last 1", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.SucceededPods).To(BeNumerically(">=", 1))
				})
				It("Number of allocated GPUs should be 2", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.TotalAllocatedGPU).To(Equal("2"))
				})
				It("Number of allocated GPU memory should be 2400", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.TotalAllocatedGPUMemory).To(Equal(int64(2400)))
				})
				It("Degraded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Degraded, "")
				})

				Context("job completed full", func() {
					BeforeEach(func() {
						jobWorkload := &job.JobWorkload{}

						updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 1, 0, 0, nil)
						trainingWorkload := getTrainingJob(ctx, k8sClient, jobName, namespace)
						trainingWorkload.CopyTo(jobWorkload)
						_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
						Expect(err).ToNot(HaveOccurred())
					})
					It("Succeeded condition should be True", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Succeeded)
					})
					It("Pending condition should be False", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
					})
					It("Initialized condition should be False", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Initialized, "")
					})
					It("Failed condition should be False", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Failed, "")
					})
					It("Suspended condition should be False", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Suspended, "")
					})
					It("CompletionTime should be set", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.CompletionTime).ToNot(BeNil())
					})
					It("Number of running pods should be 0", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.RunningPods).To(Equal(int32(0)))
					})
					It("Number of pending pods should be 0", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.PendingPods).To(Equal(int32(0)))
					})
					It("Number of succeeded pods should be at last 1", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.SucceededPods).To(BeNumerically(">=", 1))
					})
					It("Number of allocated GPUs should be 0", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.TotalAllocatedGPU).To(Equal("0"))
					})
					It("Number of allocated GPU memory should be 0", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
					})
					It("Number of ready pods should be 0", func() {
						tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpTrain.Status.ReadyPods).To(Equal(int32(0)))
					})
				})
			})
			Context("Job is Failed", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job failed

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 0, 1, 0, nil)
					trainingWorkload := getTrainingJob(ctx, k8sClient, jobName, namespace)
					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Suspended condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Suspended, "")
				})
				It("Pending condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Initialized, "")

				})
				It("Number of running pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.RunningPods).To(Equal(int32(0)))
				})
				It("Number of pending pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.SucceededPods).To(Equal(int32(0)))
				})
				It("Number of failed pods should at least 1", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.FailedPods).To(BeNumerically(">=", 1))
				})
				It("CompletionTime should be not nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.CompletionTime.IsZero()).To(BeFalse())
				})
				It("failed condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Failed)

				})
				It("Ready condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Ready, aggregator.WorkloadHasIssuesMessage)
				})
				It("Succeeded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Succeeded, "")
				})
				It("Runtime issues are calculated", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpTrain.Status.Issues.Compute).To(HaveLen(1))
					computeIssue := tmpTrain.Status.Issues.Compute[0]

					Expect(computeIssue.Kind).To(Equal(constants.KindRunaiJob))
					Expect(computeIssue.Name).To(Equal(jobName))
					Expect(computeIssue.Name).To(Equal(jobName))
					Expect(computeIssue.Description).To(ContainSubstring(jobName))
					Expect(computeIssue.Description).To(ContainSubstring(jobFailedMsg))
				})
				It("Degraded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Job is Running", func() {
				BeforeEach(func() {
					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					jobWorkload := &job.JobWorkload{}
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set running

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, false, 0, 0, 2, nil)
					err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(trainingWorkload), trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					Expect(jobWorkload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
				})
				It("number of running pods should be at least 1", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.RunningPods).To(BeNumerically(">=", 1))
				})
				It("number of gpu should be 2", func() {
					tmpTrina := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrina.Status.TotalAllocatedGPU).To(Equal("2"))
				})
				It("number of gpu memory should be 2400", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.TotalAllocatedGPUMemory).To(Equal(int64(2400)))
				})
				It("Initialized condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Initialized)
				})
				It("Pending condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Pending, "")
				})
				It("Succeeded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Succeeded, "")
				})
				It("Failed condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Failed, "")
				})
				It("Number of running pods should be at least 1", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpTrain.Status.RunningPods).To(BeNumerically(">=", 1))
				})
				It("Number of pending pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpTrain.Status.PendingPods).To(Equal(int32(0)))
				})
				It("StartTime should not be nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpTrain.Status.StartTime).ToNot(BeNil())
				})
				It("CompletionTime should be nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpTrain.Status.CompletionTime).To(BeNil())
				})
				It("Workload should restore job spec", func() {
					spec := runaiJobV1.JobSpec{Parallelism: ptr.To[int32](3), Completions: ptr.To[int32](3)}

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, true, 0, 0, 0, &spec)
					err := k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					tmpJobWorkload := &job.JobWorkload{}
					trainingWorkload.CopyTo(tmpJobWorkload)

					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, tmpJobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					runaiJob := runaiJobV1.RunaiJob{}
					err = k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, &runaiJob)
					Expect(err).ToNot(HaveOccurred())
					Expect(*runaiJob.Spec.Parallelism).To(Equal(int32(1)))

				})
				It("Ready condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Ready)
				})
				It("Degraded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Job is Pending", func() {
				BeforeEach(func() {
					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					jobWorkload := &job.JobWorkload{}
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set pending

					updateRunaiJob(ctx, k8sClient, jobName, 1, 0, 0, false, 0, 0, 0, nil)
					err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(trainingWorkload), trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

					trainingWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Pending condition should be True", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpTrain.Status.Conditions, common.Pending)
				})
				It("Initialized condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Initialized, "")
				})
				It("Suspend condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Suspended, "")
				})
				It("Failed condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Failed, "")
				})
				It("Succeeded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Succeeded, "")
				})
				It("Number of running pods should be 0", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpTrain.Status.RunningPods).To(Equal(int32(0)))
				})
				It("completionTime should be nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.CompletionTime).To(BeNil())
				})
				It("startTime should be nil", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.StartTime).To(BeNil())
				})
				It("Degraded condition should be False", func() {
					tmpTrain := getTrainingJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpTrain.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Min ready Required replicas status", func() {
				BeforeEach(func() {
					trainingWorkload = createTrainingJob(ctx, jobName, k8sClient)
					jobWorkload := &job.JobWorkload{}
					trainingWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					// second reconcile
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())

				})
				It("runai job should have min ready replicas status", func() {
					tmpTraining := getTrainingJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTraining.Status.MinRequiredReadyPods).To(Equal(int32(1)))
				})
			})
			Context("LoadBalancer Service Status", func() {
				It("should update status with LoadBalancer IPs/hostnames for LoadBalancer service", func() {
					// Build workload with LoadBalancer service
					workloadBuilder := workload.NewJobWorkloadBuilder().
						WithName(jobName).
						WithNamespace(namespace).
						WithTraining().
						WithImage("test-image").
						WithCommand("test-command").
						WithPorts([]string{"8080"}).
						WithDefaultServiceType(constants.ServiceTypeLoadBalancer)
					jobWorkload := workloadBuilder.Build()
					trainingWorkload := twv1.NewFromJobWorkload(jobWorkload)
					Expect(k8sClient.Create(ctx, trainingWorkload)).To(Succeed())
					trainingWorkload.CopyTo(jobWorkload)

					// Run reconcile twice
					for i := 0; i < 2; i++ {
						_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
						Expect(err).ToNot(HaveOccurred())
					}
					srvList := v1.ServiceList{}
					err := k8sClient.List(ctx, &srvList, k8sclient.InNamespace(namespace))
					Expect(err).ToNot(HaveOccurred())
					Expect(srvList.Items).To(HaveLen(1))
					srv := &srvList.Items[0]
					// Update its status to simulate LoadBalancer ingress
					srv.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
						{IP: "2.2.2.2"},
						{Hostname: "host-b"},
						{IP: "1.1.1.1", Hostname: "host-a"},
					}
					err = k8sClient.Status().Update(ctx, srv)
					Expect(err).ToNot(HaveOccurred())
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					// Fetch updated workload
					updated := getTrainingJob(ctx, k8sClient, jobName, namespace)
					// Find the ServiceStatus for the service
					Expect(updated.Status.ServiceStatuses).To(HaveLen(1))
					serviceStatus := updated.Status.ServiceStatuses[0]

					Expect(serviceStatus.LoadBalancer.Ingress).To(ConsistOf(
						And(HaveField("IP", "1.1.1.1"), HaveField("Hostname", "host-a")),
						And(HaveField("IP", "2.2.2.2"), HaveField("Hostname", "")),
						And(HaveField("IP", ""), HaveField("Hostname", "host-b")),
					))
					//  remove the service status and check if the service status is empty
					err = k8sClient.List(ctx, &srvList, k8sclient.InNamespace(namespace))
					Expect(err).ToNot(HaveOccurred())
					Expect(srvList.Items).To(HaveLen(1))
					srv = &srvList.Items[0]
					// Update its status to simulate LoadBalancer ingress
					srv.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{}
					err = k8sClient.Status().Update(ctx, srv)
					Expect(err).ToNot(HaveOccurred())
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
					Expect(err).ToNot(HaveOccurred())
					// Fetch updated workload
					updated = getTrainingJob(ctx, k8sClient, jobName, namespace)
					// Find the ServiceStatus for the service
					Expect(updated.Status.ServiceStatuses).To(HaveLen(1))
					serviceStatus = updated.Status.ServiceStatuses[0]
					Expect(serviceStatus.LoadBalancer.Ingress).To(BeEmpty())
				})
			})
		})
	})
	Context("Interactive Workload", func() {
		var (
			interactiveWorkload *iw.InteractiveWorkload
			jobWorkload         *job.JobWorkload
			statusUpdater       update.StatusUpdater
		)
		BeforeEach(func() {
			statusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.InteractiveWorkload)
			Expect(err).ToNot(HaveOccurred())
			statusUpdater = statusUpdaterFactory.CreateStatusUpdater(k8sClient)
			finalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.InteractiveWorkload)
			Expect(err).ToNot(HaveOccurred())
			reconcile = CreateJobController(k8sClient, statusUpdater,
				finalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
				defaultPodLabels, defaultAnnotation)
		})

		Context("Sanity test", func() {
			Context("Job is Mark for deletion", func() {
			})
			Context("Workload failed building sub resource ", func() {
				Context("Nodepool does not exits", func() {

					BeforeEach(func() {

						jobWorkload := workload.NewJobWorkloadBuilder().WithName(jobName).
							WithInteractive().WithNamespace(namespace).WithNodePools("non-existing-pool").Build()
						interactiveWorkload = iw.NewFromJobWorkload(jobWorkload)
						err := k8sClient.Create(ctx, interactiveWorkload)
						Expect(err).ToNot(HaveOccurred())
						// workload with non-existing nodepool
						interactiveWorkload.CopyTo(jobWorkload)
						_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("node pool non-existing-pool does not exist"))
					})
					It("Worklaod should be failed", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						Expect(tmpInteractive.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to a Internal issue"))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.Issues.Internal).To(HaveLen(1))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.Issues.Internal[0].Description).To(ContainSubstring("failed to build job spec: error: node pool non-existing-pool does not exist"))
					})
					It("Suspended condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
					})
					It("Pending condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
					})
					It("Initialized condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
					})
					It("Number of running pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
					})
					It("Number of pending pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
					})
					It("Number of succeeded pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
					})
					It("Number of failed pods should be 0 ", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.FailedPods).To(Equal(int32(0)))
					})
					It("CompletionTime should not be nil", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.CompletionTime.IsZero()).To(BeFalse())
					})
					It("acknowledge time should be not nil", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())
					})
					It("failed condition should be True", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Failed)
					})
					It("Ready condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, aggregator.WorkloadHasIssuesMessage)
					})
					It("Degraded condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
					})
				})
				Context("Existing Pvc does not exits", func() {
					var (
						pvcNonExists *v1.PersistentVolumeClaim
					)
					const pvcName = "test-pvc-1"
					BeforeEach(func() {
						pvcNonExists = resourceUtils.CreatePvc(pvcName, projectName, resource.MustParse(pvcSize),
							"", "", "", nil, nil, nil)
						jobWorkload := workload.NewJobWorkloadBuilder().WithName(jobName).WithImage("ubi-minimal:latest").
							WithInteractive().WithNamespace(namespace).WithPVC(pvcNonExists, true, false).Build()
						interactiveWorkload = iw.NewFromJobWorkload(jobWorkload)
						err := k8sClient.Create(ctx, interactiveWorkload)
						Expect(err).ToNot(HaveOccurred())
						// workload with non-existing nodepool
						interactiveWorkload.CopyTo(jobWorkload)
						_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("persistentvolumeclaims \"%s\" not found", pvcName)))
					})
					It("Worklaod should be failed", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						Expect(tmpInteractive.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to a Internal issue"))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.Issues.Internal).To(HaveLen(1))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.Issues.Internal[0].Description).To(ContainSubstring(fmt.Sprintf("persistentvolumeclaims \"%s\" not found", pvcName)))
					})
					It("Suspended condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
					})
					It("Pending condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
					})
					It("Initialized condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
					})
					It("Number of running pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
					})
					It("Number of pending pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
					})
					It("Number of succeeded pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
					})
					It("Number of failed pods should be 0 ", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.FailedPods).To(Equal(int32(0)))
					})
					It("CompletionTime should not be nil", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.CompletionTime.IsZero()).To(BeFalse())
					})
					It("acknowledge time should be not nil", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())
					})
					It("Ready condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, aggregator.WorkloadHasIssuesMessage)
					})
					It("Degraded condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
					})
				})
				Context("New Pvc already exists by other workload", func() {
					var (
						newPvc *v1.PersistentVolumeClaim
					)
					const pvcName = "test-pvc-1"
					BeforeEach(func() {
						otherUid := iw.NewFromJobWorkload(workload.NewJobWorkloadBuilder().
							WithInteractive().WithName("other").
							WithNamespace(namespace).WithUid("other").Build())
						newPvc = resourceUtils.CreatePvc(pvcName, projectName, resource.MustParse(pvcSize),
							otherUid.UID, constants.KindInteractiveWorkload, "", nil, nil, nil)
						Expect(controllerutils.SetControllerReference(otherUid, newPvc, k8sClient.Scheme())).To(Succeed())
						Expect(k8sClient.Create(ctx, newPvc)).To(Succeed())
						jobWorkload := workload.NewJobWorkloadBuilder().WithName(jobName).WithImage("ubi-minimal:latest").WithUid(jobName).
							WithInteractive().WithNamespace(namespace).WithPVC(newPvc, true, false).Build()
						interactiveWorkload = iw.NewFromJobWorkload(jobWorkload)
						err := k8sClient.Create(ctx, interactiveWorkload)
						Expect(err).ToNot(HaveOccurred())
						// workload with non-existing nodepool
						interactiveWorkload.CopyTo(jobWorkload)
						_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("persistentvolumeclaims \"%s\" not found", pvcName)))
					})
					It("Worklaod should be failed", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						Expect(tmpInteractive.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to a Internal issue"))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.Issues.Internal).To(HaveLen(1))
						Expect(tmpInteractive.Status.WorkloadCommonStatus.Issues.Internal[0].Description).To(ContainSubstring(fmt.Sprintf("persistentvolumeclaims \"%s\" not found", pvcName)))
					})
					It("Suspended condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
					})
					It("Pending condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
					})
					It("Initialized condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
					})
					It("Number of running pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
					})
					It("Number of pending pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
					})
					It("Number of succeeded pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
					})
					It("Number of failed pods should be 0 ", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.FailedPods).To(Equal(int32(0)))
					})
					It("CompletionTime should not be nil", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.CompletionTime.IsZero()).To(BeFalse())
					})
					It("acknowledge time should be not nil", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.FirstAcknowledgedTime.IsZero()).To(BeFalse())
					})
					It("failed condition should be True", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Failed)
					})
					It("Ready condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, aggregator.WorkloadHasIssuesMessage)
					})
					It("Degraded condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
					})
				})
			})
			Context("First workload reconcile", func() {

				BeforeEach(func() {
					err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
					Expect(err).ToNot(HaveOccurred())
					builder := workload.NewJobWorkloadBuilder()
					builder.WithName(jobName).WithNamespace(namespace).WithTraining()
					jobWorkload = builder.WithImage("test-image").WithCommand("test-command").SetOldJupyter().Build()
					interactiveWorkload = iw.NewFromJobWorkload(jobWorkload)
					pvcs := map[string]fields.PvcsItemAndRules{}
					pvcs["pvc1"] = fields.PvcsItemAndRules{
						Value: &fields.PvcsItemValue{
							ClaimName: ptr.To("regularclaimname"),
							Size:      ptr.To("1Gi"),
							Path:      ptr.To("/path"),
							ReadOnly:  ptr.To(true),
							Ephemeral: ptr.To(false),
						},
					}
					pvcs["pvc2"] = fields.PvcsItemAndRules{
						Value: &fields.PvcsItemValue{
							ClaimName: ptr.To("ephemeralClaimName"),
							Size:      ptr.To("1Gi"),
							Path:      ptr.To("/path/cyber"),
							ReadOnly:  ptr.To(true),
							Ephemeral: ptr.To(true),
						},
					}
					interactiveWorkload.Spec.Pvcs = fields.NewPvcsField(pvcs)
					err = k8sClient.Create(ctx, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
					interactiveWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

				})
				AfterEach(func() {
					err := os.Unsetenv(workloadconfig.EnvVarClusterDomain)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Should set first Ack a job", func() {
					tmp := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmp.Status.FirstAcknowledgedTime).ToNot(BeNil())
				})
				It("Should set creationStatus to submitted", func() {
					tmp := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmp.Status.CreationStatus).To(Equal(common.StatusSubmitted))
				})
				It("All Condition should be false", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, "")
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
				It("Phase should be 'Creating'", func() {
					tmp := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmp.Status.Phase).To(Equal(common.PhaseCreating.String()))
				})

			})
			Context("Init context", func() {
				BeforeEach(func() {
					err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
					Expect(err).ToNot(HaveOccurred())
					builder := workload.NewJobWorkloadBuilder()
					builder.WithName(jobName).WithNamespace(namespace).WithTraining()
					jobWorkload = builder.WithImage("test-image").WithCommand("test-command").SetOldJupyter().Build()
					interactiveWorkload = iw.NewFromJobWorkload(jobWorkload)

					err = k8sClient.Create(ctx, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
					interactiveWorkload.CopyTo(jobWorkload)
				})

				It("does not fail on mnnvl", func() {
					jobWorkload.Labels[mnnvl.FeatureLabelKey] = mnnvl.Required

					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
			})
			Context("Job is Suspending", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job suspended
					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, true, 0, 0, 0, nil)
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					tmpInteractive.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Suspended condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Suspended)
				})
				It("Pending condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
				})
				It("Number of running pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
				})
				It("Number of pending pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
				})
				It("Number of failed pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.FailedPods).To(Equal(int32(0)))
				})
				It("CompletionTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.CompletionTime).To(BeNil())
				})
				It("StartTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.StartTime).To(BeNil())
				})
				It("Ready condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, "job suspended")
				})
				It("Succeeded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
				})
				It("Failed condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Job is Suspending", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job not suspended
					updateRunaiJob(ctx, k8sClient, jobName, 1, 1, 0, false, 0, 0, 0, nil)
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					// set workload to suspended
					tmpInteractive.Spec.Active = fields.NewOptionalBoolField(ptr.To(false))
					// update the workload
					Expect(k8sClient.Update(ctx, tmpInteractive)).To(Succeed())
					tmpInteractive.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Suspended condition should be false", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
				})
				It("Pending condition should be true", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Pending)
				})
				It("Initialized condition should be false", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
				})
				It("Number of running pods should be 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(1)))
				})
				It("Number of pending pods should be 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(1)))
				})
				It("Number of succeeded pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
				})
				It("Number of failed pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.FailedPods).To(Equal(int32(0)))
				})
				It("CompletionTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.CompletionTime).To(BeNil())
				})
				It("StartTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.StartTime).To(BeNil())
				})
				It("Ready condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, "workload is suspending")
				})
				It("Succeeded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
				})
				It("Failed condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
				It("workload phase should be suspending", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseSuspending.String()))
				})
			})
			Context("Job is Resuming", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job not suspended
					updateRunaiJob(ctx, k8sClient, jobName, 1, 1, 0, false, 0, 0, 0, nil)
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					// set workload to suspended
					tmpInteractive.Spec.Active = fields.NewOptionalBoolField(ptr.To(false))
					// update the workload
					Expect(k8sClient.Update(ctx, tmpInteractive)).To(Succeed())
					tmpInteractive.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
					tmpInteractive = getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseSuspending.String()))
					// set job to suspended
					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, true, 0, 0, 0, nil)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					tmpInteractive = getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseSuspended.String()))
					tmpInteractive.Spec.Active = fields.NewOptionalBoolField(ptr.To(true))
					// set job to resuming
					Expect(k8sClient.Update(ctx, tmpInteractive)).To(Succeed())
					tmpInteractive.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Suspended condition should be true", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Suspended)
				})
				It("Pending condition should be false", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be false", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
				})
				It("Number of running pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
				})
				It("Number of pending pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
				})
				It("Number of failed pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.FailedPods).To(Equal(int32(0)))
				})
				It("CompletionTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.CompletionTime).To(BeNil())
				})
				It("StartTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.StartTime).To(BeNil())
				})
				It("Ready condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, "job suspended")
				})
				It("Succeeded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
				})
				It("Failed condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
				It("workload phase should be resuming", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseResuming.String()))
				})
			})
			Context("Job is half Completed", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job half completed

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, false, 1, 0, 2, nil)
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					tmpInteractive.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Succeeded condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Succeeded)
				})
				It("Pending condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Initialized)
				})
				It("Failed condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
				})
				It("Suspended condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
				})
				It("CompletionTime should be set", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.CompletionTime).ToNot(BeNil())
				})
				It("Number of running pods should be 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(1)))
				})
				It("Number of pending pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be at last 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.SucceededPods).To(BeNumerically(">=", 1))
				})
				It("Number of allocated GPUs should be 2", func() {
					tmpTrain := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.TotalAllocatedGPU).To(Equal("2"))
				})
				It("Number of allocated GPU memory should be 2400", func() {
					tmpTrain := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpTrain.Status.TotalAllocatedGPUMemory).To(Equal(int64(2400)))
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})

				Context("job completed", func() {
					BeforeEach(func() {
						jobWorkload := &job.JobWorkload{}

						updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 1, 0, 0, nil)
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						tmpInteractive.CopyTo(jobWorkload)
						_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
						Expect(err).ToNot(HaveOccurred())
					})
					It("Succeeded condition should be True", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Succeeded)
					})
					It("Pending condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
					})
					It("Initialized condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
					})
					It("Failed condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
					})
					It("Suspended condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
					})
					It("CompletionTime should be set", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.CompletionTime).ToNot(BeNil())
					})
					It("Number of running pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
					})
					It("Number of pending pods should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
					})
					It("Number of succeeded pods should be at last 1", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.SucceededPods).To(BeNumerically(">=", 1))
					})
					It("Number of allocated GPUs should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.TotalAllocatedGPU).To(Equal("0"))
					})
					It("Number of allocated GPU memory should be 0", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
					})
					It("Degraded condition should be False", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
					})
				})
			})
			Context("Job is Failed", func() {
				BeforeEach(func() {
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set job failed

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 0, 1, 0, nil)
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					tmpInteractive.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, tmpInteractive)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Suspended condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
				})
				It("Pending condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
				})
				It("Initialized condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")

				})
				It("Number of running pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
				})
				It("Number of pending pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
				})
				It("Number of succeeded pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.SucceededPods).To(Equal(int32(0)))
				})
				It("Number of failed pods should at least 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.FailedPods).To(BeNumerically(">=", 1))
				})
				It("CompletionTime should be not nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.CompletionTime.IsZero()).To(BeFalse())
				})
				It("failed condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Failed)

				})
				It("Ready condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Ready, aggregator.WorkloadHasIssuesMessage)
				})
				It("Succeeded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Job is Running", func() {
				BeforeEach(func() {
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set running
					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, false, 0, 0, 2, nil)
					err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(interactiveWorkload), interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					interactiveWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("number of running pods should be at least 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.RunningPods).To(BeNumerically(">=", 1))
				})
				It("number of gpu should be 2", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.TotalAllocatedGPU).To(Equal("2"))
				})
				It("number of gpu memory should be 2400", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.TotalAllocatedGPUMemory).To(Equal(int64(2400)))
				})
				It("Initialized condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Initialized)
				})
				It("Pending condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Pending, "")
				})
				It("Succeeded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
				})
				It("Failed condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
				})
				It("Number of running pods should be at least 1", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpInteractive.Status.RunningPods).To(BeNumerically(">=", 1))
				})
				It("Number of pending pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpInteractive.Status.PendingPods).To(Equal(int32(0)))
				})
				It("StartTime should not be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpInteractive.Status.StartTime).ToNot(BeNil())
				})
				It("CompletionTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpInteractive.Status.CompletionTime).To(BeNil())
				})
				It("Workload should restore job spec", func() {
					spec := runaiJobV1.JobSpec{Parallelism: ptr.To[int32](3), Completions: ptr.To[int32](3)}

					updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 1, 0, 0, &spec)
					err := k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
					tmpJobWorkload := &job.JobWorkload{}
					interactiveWorkload.CopyTo(tmpJobWorkload)

					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, tmpJobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
					runaiJob := runaiJobV1.RunaiJob{}
					err = k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, &runaiJob)
					Expect(err).ToNot(HaveOccurred())
					Expect(*runaiJob.Spec.Parallelism).To(Equal(int32(1)))

				})
				It("Ready condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Ready)
				})
				It("'Running' workflow turns to 'Creating' if deleting its job", func() {
					runaiJob := &runaiJobV1.RunaiJob{}
					err := k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, runaiJob)
					Expect(err).ToNot(HaveOccurred())

					err = k8sClient.Delete(ctx, runaiJob)
					Expect(err).ToNot(HaveOccurred())

					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseRunning.String()))

					tmpJobWorkload := &job.JobWorkload{}
					interactiveWorkload.CopyTo(tmpJobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, tmpJobWorkload, tmpInteractive)
					Expect(err).ToNot(HaveOccurred())

					tmpInteractive = getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseCreating.String()))
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Job is Pending", func() {
				BeforeEach(func() {
					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					// set running
					updateRunaiJob(ctx, k8sClient, jobName, 1, 0, 0, false, 0, 0, 0, nil)
					err = k8sClient.Get(ctx, k8sclient.ObjectKeyFromObject(interactiveWorkload), interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

					interactiveWorkload.CopyTo(jobWorkload)
					_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
				})
				It("Pending condition should be True", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsTrue(tmpInteractive.Status.Conditions, common.Pending)
				})
				It("Initialized condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Initialized, "")
				})
				It("Suspend condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Suspended, "")
				})
				It("Failed condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Failed, "")
				})
				It("Succeeded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Succeeded, "")
				})
				It("Number of running pods should be 0", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)

					Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(0)))
				})
				It("completionTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.CompletionTime).To(BeNil())
				})
				It("startTime should be nil", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					Expect(tmpInteractive.Status.StartTime).To(BeNil())
				})
				It("Degraded condition should be False", func() {
					tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
					workload.CheckConditionIsFalseOrUnset(tmpInteractive.Status.Conditions, common.Degraded, "")
				})
			})
			Context("Job uses NodePool", func() {

				BeforeEach(func() {
					nodepool = resourceUtils.CreateNodePool(nodepoolName, "key", "value")
					Expect(k8sClient.Create(ctx, nodepool)).To(Succeed())
					nodepoolKai = resourceUtils.CreateKaiNodePool(nodepoolName, "key", "value")
					Expect(k8sClient.Create(ctx, nodepoolKai)).To(Succeed())

					interactiveWorkload = createSimpleInteractiveJob(jobName, k8sClient, ctx)
					interactiveWorkload.Spec.NodePools = fields.NewStringField(nodepoolName)
					Expect(k8sClient.Update(ctx, interactiveWorkload)).To(Succeed())

					jobWorkload := &job.JobWorkload{}
					interactiveWorkload.CopyTo(jobWorkload)
					// first reconcile
					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())

				})
				It("runai job should have node affinity", func() {
					checkRunaiNodeAffinity(ctx, jobName, namespace, "key", "value", k8sClient)
				})
				It("runai job should have the same node affinity even if nodepool changed", func() {

					nodepool.Spec.LabelKey = "key2"
					nodepool.Spec.LabelValue = "value2"
					Expect(k8sClient.Update(ctx, nodepool)).To(Succeed())

					// Mirror the mutation onto the kai twin, so this spec still exercises an actual
					// pool change under both values of the compile-time UseKaiNodePools flag.
					// TODO: drop once constants.UseKaiNodePools=true and the legacy fixture goes
					nodepoolKai.Spec.LabelKey = "key2"
					nodepoolKai.Spec.LabelValue = "value2"
					Expect(k8sClient.Update(ctx, nodepoolKai)).To(Succeed())
					interactiveWorkload = getInteractiveJob(ctx, k8sClient, jobName, namespace)
					jobWorkload := &job.JobWorkload{}
					interactiveWorkload.CopyTo(jobWorkload)

					_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
					Expect(err).ToNot(HaveOccurred())
					checkRunaiNodeAffinity(ctx, jobName, namespace, "key", "value", k8sClient)
				})
				Context("Upgrading workload without status", func() {
					BeforeEach(func() {
						interactiveWorkload = getInteractiveJob(ctx, k8sClient, jobName, namespace)
						interactiveWorkload.Status = iw.InteractiveWorkloadStatus{}
						Expect(k8sClient.Update(ctx, interactiveWorkload)).To(Succeed())
						jobWorkload := &job.JobWorkload{}
						interactiveWorkload.CopyTo(jobWorkload)
						// updating node pool
						nodepool.Spec.LabelKey = "key3"
						nodepool.Spec.LabelValue = "value3"
						Expect(k8sClient.Update(ctx, nodepool)).To(Succeed())

						// Mirror the mutation onto the kai twin, so this spec still exercises an actual
						// pool change under both values of the compile-time UseKaiNodePools flag.
						// TODO: drop once constants.UseKaiNodePools=true and the legacy fixture goes
						nodepoolKai.Spec.LabelKey = "key3"
						nodepoolKai.Spec.LabelValue = "value3"
						Expect(k8sClient.Update(ctx, nodepoolKai)).To(Succeed())
						// first reconcile
						_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
						Expect(err).ToNot(HaveOccurred())

					})
					It("runai job should have node affinity", func() {
						checkRunaiNodeAffinity(ctx, jobName, namespace, "key", "value", k8sClient)
					})
				})
				Context("Min ready Required replicas status", func() {
					BeforeEach(func() {
						interactiveWorkload = getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(k8sClient.Update(ctx, interactiveWorkload)).To(Succeed())
						jobWorkload := &job.JobWorkload{}
						interactiveWorkload.CopyTo(jobWorkload)
						// first reconcile
						_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
						Expect(err).ToNot(HaveOccurred())

					})
					It("runai job should have min ready replicas status", func() {
						tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
						Expect(tmpInteractive.Status.MinRequiredReadyPods).To(Equal(int32(1)))
					})
				})

			})
		})
	})
	Context("Workload Status Phase validation", func() {
		var (
			interactiveWorkload *iw.InteractiveWorkload
			//jobWorkload         *job.JobWorkload
			statusUpdater update.StatusUpdater
		)
		BeforeEach(func() {
			statusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.InteractiveWorkload)
			Expect(err).ToNot(HaveOccurred())
			statusUpdater = statusUpdaterFactory.CreateStatusUpdater(k8sClient)
			finalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.InteractiveWorkload)
			Expect(err).ToNot(HaveOccurred())
			reconcile = CreateJobController(k8sClient, statusUpdater,
				finalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
				defaultPodLabels, defaultAnnotation)
		})

		Context("Job is Pending", func() {
			BeforeEach(func() {
				interactiveWorkload = createInteractiveJobWithPorts(jobName, k8sClient, ctx)
				jobWorkload := &job.JobWorkload{}
				interactiveWorkload.CopyTo(jobWorkload)

				Expect(interactiveWorkload.Status.Phase).To(Equal(""))
				// first reconcile
				_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Phase - Pending: reconcile executed twice (no changes)", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 0, 0, 0, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseCreating.String()))
			})
			It("Phase - Pending: reconcile with pending job", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 1, 0, 0, false, 0, 0, 0, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhasePending.String()))
			})
			It("Phase - Initializing: reconcile with running attributes", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 1, 0, false, 0, 0, 2, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseInitializing.String()))
				Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(1)))
				Expect(tmpInteractive.Status.ReadyPods).To(Equal(int32(0)))
			})
			It("Phase - Running: reconcile with running attributes", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, false, 0, 0, 2, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseRunning.String()))
			})
			It("Phase - Degraded: reconcile with running transition to degraded state", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 1, false, 0, 0, 2, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.RunningPods).To(Equal(int32(1)))
				// Phase switch to degraded phase after being degraded for at least 1 minute
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseRunning.String()))

				rules := []client.ResourceFailingRule{
					{
						ResourceKind: constants.KindService,
						ResourceName: fmt.Sprintf("iw-%s-0", jobName),
						StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: "Bla", Code: http.StatusBadRequest}},
						ClientAction: client.Update,
					},
				}
				fakeClientPtr.UpdateRules(rules)

				jobWorkload := &job.JobWorkload{}
				tmpInteractive.CopyTo(jobWorkload)

				_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, tmpInteractive)
				Expect(err).To(HaveOccurred())

				tmpInteractive = getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseDegraded.String()))
				Expect(tmpInteractive.Status.PhaseMessage).To(Equal("Workload is in degraded state due to a Network issue"))
			})
			It("Phase - Suspended: reconcile with suspended job", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, true, 0, 0, 0, nil)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				tmpInteractive.Spec.Active = fields.NewOptionalBoolField(ptr.To(false))
				err := k8sClient.Update(ctx, tmpInteractive)
				Expect(err).ToNot(HaveOccurred())
				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive = getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseSuspended.String()))
			})
			It("Phase - Complete: reconcile with complete job", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 1, 0, 0, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseCompleted.String()))
			})
			It("Phase - Failed: reconcile with failed job", func() {

				updateRunaiJob(ctx, k8sClient, jobName, 0, 0, 0, false, 0, 1, 0, nil)

				executeReconcile(ctx, reconcile, k8sClient, jobName)
				tmpInteractive := getInteractiveJob(ctx, k8sClient, jobName, namespace)
				Expect(tmpInteractive.Status.Phase).To(Equal(common.PhaseFailed.String()))
			})
		})
	})
	Context("Workload Resource List", func() {
		var (
			jobWorkload   *job.JobWorkload
			statusUpdater update.StatusUpdater
		)
		BeforeEach(func() {
			statusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.InteractiveWorkload)
			Expect(err).ToNot(HaveOccurred())
			statusUpdater = statusUpdaterFactory.CreateStatusUpdater(k8sClient)
			finalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.InteractiveWorkload)
			Expect(err).ToNot(HaveOccurred())
			reconcile = CreateJobController(k8sClient, statusUpdater,
				finalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
				defaultPodLabels, defaultAnnotation)
		})

		Context("Dangling Resource Clearance - Interactive", func() {
			var (
				interactiveWorkload *iw.InteractiveWorkload
			)

			BeforeEach(func() {

				interactiveWorkload = createInteractiveJobWithPorts(jobName, k8sClient, ctx)
				jobWorkload = &job.JobWorkload{}
				interactiveWorkload.CopyTo(jobWorkload)

				Expect(interactiveWorkload.Status.Phase).To(Equal(""))
				// first reconcile
				_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Creating and removing dangling resource", func() {
				// Service existence validation
				service := &v1.Service{}
				serviceName := fmt.Sprintf("iw-%s-0", interactiveWorkload.Name)
				err := k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: interactiveWorkload.Namespace, Name: serviceName}, service)
				Expect(err).ToNot(HaveOccurred())

				// Ingress existence validation
				ingress := &network.Ingress{}
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: interactiveWorkload.Namespace, Name: serviceName}, ingress)
				Expect(err).ToNot(HaveOccurred())

				// Creating new service with labels match workload resource
				labels := map[string]string{constants.LabelRelease: interactiveWorkload.Name}
				serviceOne := resourceUtils.CreateDumbService("not_in_list", interactiveWorkload.Namespace, interactiveWorkload.UID, interactiveWorkload.Kind, labels)

				err = k8sClient.Create(ctx, serviceOne)
				Expect(err).ToNot(HaveOccurred())

				// Service existence validation
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: interactiveWorkload.Namespace, Name: "not_in_list"}, service)
				Expect(err).ToNot(HaveOccurred())

				// reconcile will clear it
				_, err = reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
				Expect(err).ToNot(HaveOccurred())

				// Service clear validation
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: interactiveWorkload.Namespace, Name: "not_in_list"}, service)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("services \"not_in_list\" not found"))
			})
		})
		Context("Dangling Resource Clearance - Training", func() {
			var (
				trainingWorkload *twv1.TrainingWorkload
				trainReconcile   *JobController
			)
			BeforeEach(func() {
				trainStatusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.TrainingWorkload)
				Expect(err).ToNot(HaveOccurred())
				statusUpdater = trainStatusUpdaterFactory.CreateStatusUpdater(k8sClient)
				trainFinalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.TrainingWorkload)
				Expect(err).ToNot(HaveOccurred())
				trainReconcile = CreateJobController(k8sClient, statusUpdater,
					trainFinalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
					defaultPodLabels, defaultAnnotation)

				trainingWorkload = createTrainingJobWithPorts(jobName, k8sClient, ctx)
				jobWorkload = &job.JobWorkload{}
				trainingWorkload.CopyTo(jobWorkload)

				Expect(trainingWorkload.Status.Phase).To(Equal(""))
				// first reconcile
				_, err = trainReconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Creating and removing dangling resource", func() {
				// Service existence validation
				service := &v1.Service{}
				serviceName := fmt.Sprintf("tw-%s-0", trainingWorkload.Name)
				err := k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: trainingWorkload.Namespace, Name: serviceName}, service)
				Expect(err).ToNot(HaveOccurred())

				// Ingress existence validation
				ingress := &network.Ingress{}
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: trainingWorkload.Namespace, Name: serviceName}, ingress)
				Expect(err).ToNot(HaveOccurred())

				// Creating new service with labels match workload resource
				labels := map[string]string{constants.LabelRelease: trainingWorkload.Name}
				serviceOne := resourceUtils.CreateDumbService("not_in_list", trainingWorkload.Namespace, trainingWorkload.UID, trainingWorkload.Kind, labels)

				err = k8sClient.Create(ctx, serviceOne)
				Expect(err).ToNot(HaveOccurred())

				// Service existence validation
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: trainingWorkload.Namespace, Name: "not_in_list"}, service)
				Expect(err).ToNot(HaveOccurred())

				// reconcile will clear it
				_, err = trainReconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, trainingWorkload)
				Expect(err).ToNot(HaveOccurred())

				// Service clear validation
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: trainingWorkload.Namespace, Name: "not_in_list"}, service)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("services \"not_in_list\" not found"))
			})
		})
		Context("Dangling Resource Clearance - Distributed", func() {
			var (
				distributedWorkload *v2alpha1.DistributedWorkload
				distReconcile       *JobController
			)

			BeforeEach(func() {
				//labels := map[string]string{constants.LabelRelease: jobName}
				distStatusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.DistributedWorkload)
				Expect(err).ToNot(HaveOccurred())
				statusUpdater = distStatusUpdaterFactory.CreateStatusUpdater(k8sClient)
				distFinalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.DistributedWorkload)
				Expect(err).ToNot(HaveOccurred())
				distReconcile = CreateJobController(k8sClient, statusUpdater,
					distFinalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
					defaultPodLabels, defaultAnnotation)

				distributedWorkload = createDistributedJobWithPorts(jobName, k8sClient, ctx)
				jobWorkload = &job.JobWorkload{}
				distributedWorkload.CopyTo(jobWorkload)

				Expect(distributedWorkload.Status.Phase).To(Equal(""))
				// first reconcile
				_, err = distReconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, distributedWorkload)
				Expect(err).ToNot(HaveOccurred())
			})

			It("Creating and removing dangling resource", func() {
				// Service existence validation
				service := &v1.Service{}
				serviceName := fmt.Sprintf("dw-%s-0", distributedWorkload.Name)
				err := k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: distributedWorkload.Namespace, Name: serviceName}, service)
				Expect(err).ToNot(HaveOccurred())

				// Ingress existence validation
				ingress := &network.Ingress{}
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: distributedWorkload.Namespace, Name: serviceName}, ingress)
				Expect(err).ToNot(HaveOccurred())

				// Creating new service with labels match workload resource
				labels := map[string]string{constants.LabelRelease: distributedWorkload.Name}
				serviceOne := resourceUtils.CreateDumbService("not_in_list", distributedWorkload.Namespace, distributedWorkload.UID, distributedWorkload.Kind, labels)

				err = k8sClient.Create(ctx, serviceOne)
				Expect(err).ToNot(HaveOccurred())

				// Service existence validation
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: distributedWorkload.Namespace, Name: "not_in_list"}, service)
				Expect(err).ToNot(HaveOccurred())

				// reconcile will clear it
				_, err = distReconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, distributedWorkload)
				Expect(err).ToNot(HaveOccurred())

				// Service clear validation
				err = k8sClient.Get(context.TODO(), k8sclient.ObjectKey{Namespace: distributedWorkload.Namespace, Name: "not_in_list"}, service)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("services \"not_in_list\" not found"))
			})
		})
	})
	Context("Workload Failed by Workload", func() {
		var (
			reconciler *JobController
		)

		BeforeEach(func() {
			fakeClientPtr = client.NewFakeClientWithErrorRules(scheme, []k8sclient.Object{&runaiJobV1.RunaiJob{}, &twv1.TrainingWorkload{}}, nil, project, kaiProject, namespaceObj)
			k8sClient = fakeClientPtr

			trainStatusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.TrainingWorkload)
			Expect(err).ToNot(HaveOccurred())
			statusUpdater := trainStatusUpdaterFactory.CreateStatusUpdater(k8sClient)
			trainFinalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.TrainingWorkload)
			Expect(err).ToNot(HaveOccurred())
			reconciler = CreateJobController(k8sClient, statusUpdater, trainFinalizerRemoverFactory.CreateRemoveFinalizer(k8sClient),
				recorder, defaultPodLabels, defaultAnnotation)
		})

		Context("duplicate ingress url", func() {
			const (
				firstName = "w1"
				dupName   = "w2"
			)

			BeforeEach(func() {
				k8sClient = client.NewFakeClientWithIngressUrlErrors(scheme, []k8sclient.Object{&runaiJobV1.RunaiJob{}, &twv1.TrainingWorkload{}}, project, kaiProject, namespaceObj)
				trainStatusUpdaterFactory, err := update.GetStatusUpdaterFactory(workloadCommon.TrainingWorkload)
				Expect(err).ToNot(HaveOccurred())
				statusUpdater := trainStatusUpdaterFactory.CreateStatusUpdater(k8sClient)
				trainFinalizerRemoverFactory, err := finalizers.GetRemoveFinalizerFactory(workloadCommon.TrainingWorkload)
				Expect(err).ToNot(HaveOccurred())
				reconciler = CreateJobController(k8sClient, statusUpdater,
					trainFinalizerRemoverFactory.CreateRemoveFinalizer(k8sClient), recorder,
					defaultPodLabels, defaultAnnotation)

				// create a workload with exposed custom url
				firstWorkload := createTrainingJobWithPorts(firstName, k8sClient, ctx)
				firstJob := &job.JobWorkload{}
				firstWorkload.CopyTo(firstJob)

				_, err = reconciler.Reconcile(ctx, ctrl.Request{}, firstJob, firstWorkload)
				Expect(err).NotTo(HaveOccurred())
			})
			It("Fail workload if ingress url already exist", func() {
				dupWorkload := createTrainingJobWithPorts(dupName, k8sClient, ctx)
				dupJob := &job.JobWorkload{}
				dupWorkload.CopyTo(dupJob)

				// first time - should turn to failed
				// second time - only calc status
				// third time - only calc status (making sure idempotent)
				for i := 1; i <= 3; i++ {
					_, err := reconciler.Reconcile(ctx, ctrl.Request{}, dupJob, dupWorkload)
					if i == 1 {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(expectedUpsertErrorMsg))
					} else {
						Expect(err).NotTo(HaveOccurred())
					}

					dupWorkload = getTrainingJob(ctx, k8sClient, dupName, namespace)

					Expect(dupWorkload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))
					Expect(dupWorkload.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to a Network issue"))
					Expect(dupWorkload.Status.CreationStatus).To(Equal(common.StatusFailed))
					Expect(dupWorkload.Status.CompletionTime).NotTo(BeNil())
					Expect(dupWorkload.Status.OwnedResources).To(BeNil())
					Expect(dupWorkload.Status.Issues.Network).To(HaveLen(1))
					Expect(dupWorkload.Status.Issues.Network[0].Description).To(ContainSubstring(fmt.Sprintf(strings.Split(client.UrlAlreadyExistsMsg, ":")[1], fmt.Sprintf("%s/%s", namespace, "tw-"+firstName+"-0"))))

					workload.CheckConditionIsTrue(dupWorkload.Status.Conditions, common.Failed)
					workload.CheckConditionIsFalseOrUnset(dupWorkload.Status.Conditions, common.Ready, aggregator.WorkloadHasIssuesMessage)

					if err != nil {
						workload.CheckConditionIsFalseOrUnset(dupWorkload.Status.Conditions, common.LastReconcileSucceeded, "upsert workload resources error")
					} else {
						workload.CheckConditionIsTrue(dupWorkload.Status.Conditions, common.LastReconcileSucceeded)
					}
				}

				Expect(recorder.Events).NotTo(BeNil())
				Eventually(recorder.Events).WithTimeout(3 * time.Second).
					Should(Receive(ContainSubstring("Workload failed due to a Network issue")))
			})
		})
		Context("node port already allocated", func() {
			const (
				workloadName = "w1"
			)

			It("Fail workload if node port already allocated", func() {
				serviceName := fmt.Sprintf("tw-%s-0-nodeport", workloadName)
				errMsg := fmt.Sprintf("Service \"%s\" is invalid: spec.ports[0].nodePort: Invalid value: 30000: provided port is already allocated", serviceName)
				rules := []client.ResourceFailingRule{
					{
						ResourceKind: constants.KindService,
						ResourceName: serviceName,
						StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: errMsg, Code: http.StatusBadRequest}},
					},
				}
				fakeClientPtr.UpdateRules(rules)

				jw := workload.NewJobWorkloadBuilder().
					WithName(workloadName).WithNamespace(namespace).
					WithPorts([]string{"30080:8080", "30060:6060"}).
					WithDefaultServiceType(constants.ServiceTypeNodePort).
					WithTraining().
					WithImage("test-image").
					WithCommand("test-command").
					Build()
				tw := twv1.NewFromJobWorkload(jw)

				err := k8sClient.Create(ctx, tw)
				Expect(err).ToNot(HaveOccurred())
				err = k8sClient.Status().Update(ctx, tw)
				Expect(err).ToNot(HaveOccurred())

				// first time - should turn to failed
				// second time - only calc status
				// third time - only calc status (making sure idempotent)
				for i := 1; i <= 3; i++ {
					_, err := reconciler.Reconcile(ctx, ctrl.Request{}, jw, tw)
					if i == 1 {
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring(expectedUpsertErrorMsg))
					} else {
						Expect(err).NotTo(HaveOccurred())
					}

					tw = getTrainingJob(ctx, k8sClient, workloadName, namespace)

					Expect(tw.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))
					Expect(tw.Status.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to a Network issue"))
					Expect(tw.Status.CreationStatus).To(Equal(common.StatusFailed))
					Expect(tw.Status.CompletionTime).NotTo(BeNil())
					Expect(tw.Status.OwnedResources).To(BeNil())
					Expect(tw.Status.Issues.Network).To(HaveLen(1))
					Expect(tw.Status.Issues.Network[0].Description).To(ContainSubstring(errMsg))
					Expect(tw.Status.Issues.Network[0].Description).To(ContainSubstring(serviceName))

					if err != nil {
						workload.CheckConditionIsFalseOrUnset(tw.Status.Conditions, common.LastReconcileSucceeded, "upsert workload resources error")
					} else {
						workload.CheckConditionIsTrue(tw.Status.Conditions, common.LastReconcileSucceeded)
					}
				}

				Expect(recorder.Events).NotTo(BeNil())
				Eventually(recorder.Events).WithTimeout(3 * time.Second).
					Should(Receive(ContainSubstring("Workload failed due to a Network issue")))
			})
		})
		Context("FailWorkload", func() {
			const (
				workloadName     = "w1"
				expectedPhaseMsg = "just a phase msg"
			)

			var (
				tw         *twv1.TrainingWorkload
				originalJw *job.JobWorkload
				aggr       *workloadresource.Aggregator
			)

			BeforeEach(func() {
				tw = createTrainingJobWithPorts(workloadName, k8sClient, ctx)

				originalJw = &job.JobWorkload{}
				tw.CopyTo(originalJw)

				_, err := reconciler.Reconcile(ctx, ctrl.Request{}, originalJw, tw)
				Expect(err).NotTo(HaveOccurred())

				tw = getTrainingJob(ctx, k8sClient, workloadName, namespace)
				tw.Status.PhaseMessage = expectedPhaseMsg

				jw := &job.JobWorkload{}
				tw.CopyTo(jw)

				aggr = workloadresource.NewAggregator(jw)
			})

			Context("sub-resources cleanup", func() {
				It("cleanup success", func() {
					err := reconciler.FailWorkload(ctx, aggr, originalJw, tw)
					Expect(err).NotTo(HaveOccurred())

					tw = getTrainingJob(ctx, k8sClient, workloadName, namespace)
					Expect(tw.Status.Issues).To(BeNil())

					Expect(recorder.Events).NotTo(BeNil())
					Eventually(recorder.Events).WithTimeout(3 * time.Second).
						Should(Receive(ContainSubstring(expectedPhaseMsg)))
				})
				It("cleanup error", func() {
					jobErrMsg := "runaijob delete error"
					expectedErrMsg := fmt.Sprintf("failed to cleanup workload resources. error message: %s", jobErrMsg)

					rules := []client.ResourceFailingRule{
						{
							ResourceKind: constants.KindRunaiJob,
							ResourceName: workloadName,
							StatusError:  &k8serrors.StatusError{ErrStatus: metav1.Status{Message: jobErrMsg, Code: http.StatusBadRequest}},
							ClientAction: client.Delete,
						},
					}
					fakeClientPtr.UpdateRules(rules)

					err := reconciler.FailWorkload(ctx, aggr, originalJw, tw)
					Expect(err).To(HaveOccurred())

					tw = getTrainingJob(ctx, k8sClient, workloadName, namespace)
					Expect(tw.Status.Issues).NotTo(BeNil())
					Expect(tw.Status.Issues.Internal).To(HaveLen(1))
					Expect(tw.Status.Issues.Internal[0].Name).To(Equal(workloadName))
					Expect(tw.Status.Issues.Internal[0].Kind).To(Equal(constants.KindTrainingWorkload))
					Expect(tw.Status.Issues.Internal[0].Description).To(ContainSubstring(expectedErrMsg))

					Expect(recorder.Events).NotTo(BeNil())
					Eventually(recorder.Events).WithTimeout(3 * time.Second).
						Should(Receive(ContainSubstring(expectedErrMsg)))
					Eventually(recorder.Events).WithTimeout(3 * time.Second).
						Should(Receive(ContainSubstring(expectedPhaseMsg)))
				})
			})

		})
	})
})

func executeReconcile(ctx context.Context, reconcile *JobController, k8sClient k8sclient.Client, jobName string) {
	interactiveWorkload := getInteractiveJob(ctx, k8sClient, jobName, namespace)
	jobWorkload := &job.JobWorkload{}
	interactiveWorkload.CopyTo(jobWorkload)

	_, err := reconcile.Reconcile(ctx, ctrl.Request{}, jobWorkload, interactiveWorkload)
	Expect(err).ToNot(HaveOccurred())
}

func checkRunaiNodeAffinity(ctx context.Context, jobName, namespace, key, value string, k8sClient k8sclient.Client) {
	runaiJob := runaiJobV1.RunaiJob{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, &runaiJob)
	Expect(err).ToNot(HaveOccurred())
	Expect(runaiJob.Spec.Template.Spec.Affinity.NodeAffinity).ToNot(BeNil())
	Expect(runaiJob.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
	NodeSelectorTerms := runaiJob.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0]
	Expect(NodeSelectorTerms.MatchExpressions).ToNot(BeEmpty())
	Expect(NodeSelectorTerms.MatchExpressions).To(HaveLen(1))
	Expect(NodeSelectorTerms.MatchExpressions[0].Key).To(Equal(key))
	Expect(NodeSelectorTerms.MatchExpressions[0].Operator).To(Equal(v1.NodeSelectorOpIn))
	Expect(NodeSelectorTerms.MatchExpressions[0].Values).To(HaveLen(1))
	Expect(NodeSelectorTerms.MatchExpressions[0].Values[0]).To(Equal(value))
}

func getTrainingJob(ctx context.Context, client k8sclient.Client, name, namespace string) *twv1.TrainingWorkload {
	tmp := &twv1.TrainingWorkload{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, tmp)
	Expect(err).ToNot(HaveOccurred())
	return tmp
}

func getInteractiveJob(ctx context.Context, client k8sclient.Client, name, namespace string) *iw.InteractiveWorkload {
	tmp := &iw.InteractiveWorkload{}
	err := client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, tmp)
	Expect(err).ToNot(HaveOccurred())
	return tmp
}

func createTrainingJob(ctx context.Context, jobName string, k8sClient k8sclient.Client) *twv1.TrainingWorkload {
	jobWorkload := buildTestWorkloadJob(jobName)
	trainingWorkload := twv1.NewFromJobWorkload(jobWorkload).DeepCopy()
	err := k8sClient.Create(ctx, trainingWorkload)
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.Status().Update(ctx, trainingWorkload)
	Expect(err).ToNot(HaveOccurred())
	return trainingWorkload
}

func createSimpleInteractiveJob(jobName string, k8sClient k8sclient.Client, ctx context.Context) *iw.InteractiveWorkload {
	jobWorkload := buildTestWorkloadJob(jobName)
	interactiveJob := iw.NewFromJobWorkload(jobWorkload).DeepCopy()
	err := k8sClient.Create(ctx, interactiveJob)
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.Status().Update(ctx, interactiveJob)
	Expect(err).ToNot(HaveOccurred())
	return interactiveJob
}

func createInteractiveJobWithPorts(jobName string, k8sClient k8sclient.Client, ctx context.Context) *iw.InteractiveWorkload {
	jobWorkload := buildTestWorkloadJobWithPorts(jobName, constants.KindInteractiveWorkload)
	interactiveJob := iw.NewFromJobWorkload(jobWorkload).DeepCopy()
	err := k8sClient.Create(ctx, interactiveJob)
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.Status().Update(ctx, interactiveJob)
	Expect(err).ToNot(HaveOccurred())
	return interactiveJob
}

func createTrainingJobWithPorts(jobName string, k8sClient k8sclient.Client, ctx context.Context) *twv1.TrainingWorkload {
	jobWorkload := buildTestWorkloadJobWithPorts(jobName, constants.KindTrainingWorkload)
	trainingJob := twv1.NewFromJobWorkload(jobWorkload).DeepCopy()
	err := k8sClient.Create(ctx, trainingJob)
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.Status().Update(ctx, trainingJob)
	Expect(err).ToNot(HaveOccurred())
	return trainingJob
}

func createDistributedJobWithPorts(jobName string, k8sClient k8sclient.Client, ctx context.Context) *v2alpha1.DistributedWorkload {
	jobDist := buildTestWorkloadJobWithPorts(jobName, constants.KindDistributedWorkload)
	distributedJob := v2alpha1.NewFromJobWorkload(jobDist).DeepCopy()
	err := k8sClient.Create(ctx, distributedJob)
	Expect(err).ToNot(HaveOccurred())
	err = k8sClient.Status().Update(ctx, distributedJob)
	Expect(err).ToNot(HaveOccurred())
	return distributedJob
}

func buildTestWorkloadJobWithPorts(jobName string, kind string) *job.JobWorkload {
	workloadBuilder := workload.NewJobWorkloadBuilder()
	workloadBuilder.WithName(jobName).WithNamespace(namespace)
	switch kind {
	case constants.KindTrainingWorkload:
		workloadBuilder.WithTraining()
	case constants.KindInteractiveWorkload:
		workloadBuilder.WithInteractive()
	case constants.KindDistributedWorkload:
		workloadBuilder.WithDistribution()
	}
	workload := workloadBuilder.
		WithImage("test-image").
		WithCommand("test-command").
		WithExternalUrl([]string{"http://why.like.this/"}, []int{1234}).
		Build()
	workload.Status.FirstAcknowledgedTime = nil
	blaBlaTime := &metav1.Time{Time: time.Now()} // fake time
	workload.Status.CompletionTime = nil
	workload.Status.StartTime = blaBlaTime
	workload.Status.OwnedResources = nil
	workload.Status.RunningPods = -1
	workload.Status.Conditions = []common.WorkloadCondition{}
	return workload
}

func buildTestWorkloadJob(jobName string) *job.JobWorkload {
	workloadBuilder := workload.NewJobWorkloadBuilder()
	workloadBuilder.WithName(jobName).WithNamespace(namespace).WithTraining()
	workload := workloadBuilder.WithImage("test-image").WithCommand("test-command").Build()
	workload.Status.FirstAcknowledgedTime = nil
	blaBlaTime := &metav1.Time{Time: time.Now()} // fake time
	workload.Status.CompletionTime = nil
	workload.Status.StartTime = blaBlaTime
	workload.Status.OwnedResources = nil
	workload.Status.RunningPods = -1
	workload.Status.Conditions = []common.WorkloadCondition{{
		Type:   common.Pending,
		Status: v1.ConditionTrue,
	},
		{
			Type:   common.Initialized,
			Status: v1.ConditionTrue,
		},
		{
			Type:   common.Suspended,
			Status: v1.ConditionTrue,
		},
		{
			Type:   common.Failed,
			Status: v1.ConditionFalse,
		},
		{
			Type:   common.Ready,
			Status: v1.ConditionTrue,
		},
		{
			Type:   common.Succeeded,
			Status: v1.ConditionFalse,
		},
	}
	return workload
}

func updateRunaiJob(ctx context.Context, k8sClient k8sclient.Client, jobName string,
	pending, initializing, ready int32,
	suspended bool, succeeded, failed int32, gpu int,
	spec *runaiJobV1.JobSpec) {
	runaiJob := &runaiJobV1.RunaiJob{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: jobName, Namespace: namespace}, runaiJob)
	Expect(err).ToNot(HaveOccurred())
	if gpu > 0 {
		runaiJob.Annotations = map[string]string{common.RunaiCurrentAllocatedGpusAnnotation: strconv.Itoa(gpu),
			common.RunaiCurrentAllocatedGpuMemoryAnnotation: "2400"}
	} else {
		runaiJob.Annotations = map[string]string{common.RunaiCurrentAllocatedGpusAnnotation: "0",
			common.RunaiCurrentAllocatedGpuMemoryAnnotation: "0"}
	}
	Expect(k8sClient.Update(ctx, runaiJob)).To(Succeed())

	runaiJob.Status.Pending = pending
	runaiJob.Status.Initialized = initializing
	runaiJob.Status.Ready = ready
	runaiJob.Status.Running = ready + initializing

	if suspended || succeeded > 0 || failed > 0 {
		runaiJob.Status.Conditions = make([]runaiJobV1.JobCondition, 0)
	}
	if suspended {
		runaiJob.Status.Conditions = append(runaiJob.Status.Conditions, runaiJobV1.JobCondition{
			Type:   runaiJobV1.JobSuspended,
			Status: v1.ConditionTrue,
		})
	}
	if succeeded > 0 {
		runaiJob.Status.Succeeded = succeeded
		runaiJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
		runaiJob.Status.Conditions = append(runaiJob.Status.Conditions, runaiJobV1.JobCondition{
			Type:   runaiJobV1.JobComplete,
			Status: v1.ConditionTrue,
		})
	}
	if failed > 0 {
		runaiJob.Status.Failed = failed
		runaiJob.Status.Conditions = append(runaiJob.Status.Conditions, runaiJobV1.JobCondition{
			Type:    runaiJobV1.JobFailed,
			Status:  v1.ConditionTrue,
			Message: jobFailedMsg,
		})
	}

	if spec != nil {
		err = k8sClient.Update(ctx, runaiJob)
		Expect(err).ToNot(HaveOccurred())
	}

	err = k8sClient.Status().Update(ctx, runaiJob)
	Expect(err).ToNot(HaveOccurred())

}

func CreateJobController(client k8sclient.Client, updater update.StatusUpdater,
	finalizer finalizers.RemoveFinalizer, recorder record.EventRecorder,
	podLabels map[string]string, podAnnotations map[string]string) *JobController {
	runaiBuilder := jobBuilder.NewRunaiJobBuilder(client, podLabels, podAnnotations, true)
	pvcBuilder := common2.NewPvcBuilder(client)
	serviceBuilder := common2.NewServiceBuilder(client)
	urlExporterBuilder := common2.NewExternalsUrlsBuilder(client, workloadconfig.IsOpenshift())
	commonBuilder := common2.NewShardResourcesBuilder(serviceBuilder, urlExporterBuilder, pvcBuilder)
	jobController := NewJobController(client, runaiBuilder,
		commonBuilder, update.NewSimpleStatusUpdaterManager(updater), finalizer, aggregator.NewStatusCalculator(false), managers.NewResourceManager(client, client, &issue.ComputeResourceGrouper{}),
		managers.ResourceOptions{}, recorder, managers.NewCleanupManager(client, managers.GetDefaultFailureCleanupPolicy(), recorder), true)
	return jobController
}
