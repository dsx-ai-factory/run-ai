/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregator

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	issue2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	network "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	statusUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	condition "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/status"
)

const (
	defaultNamespace  = "runai-test"
	pycharmApp        = "pycharm"
	httpPort          = 80
	service2          = "test-service-2"
	service1          = "test-service-1"
	httpPortName      = "http"
	mnnvlNotSupported = false
	mnnvlSupported    = true
	secret1           = "test-secret-1"
)

var timeHourBack = metav1.Time{Time: metav1.Now().Add(-1 * time.Hour)}

var _ = Describe("Workload StatusCalculator", func() {

	var (
		jobAdapter       *mocks.JobAdapter
		statusCalculator *StatusCalculator
		workload         *job.JobWorkload
		services         []*v1.Service
		secrets          []*v1.Secret
		aggr             *resource.Aggregator
	)

	BeforeEach(func() {
		services = []*v1.Service{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       service1,
					Namespace:  defaultNamespace,
					Generation: 1,
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeClusterIP,
					Ports: []v1.ServicePort{
						{
							Name:       httpPortName,
							Port:       httpPort,
							TargetPort: intstr.FromInt32(httpPort),
							Protocol:   v1.ProtocolTCP,
						},
					},
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:       service2,
					Namespace:  defaultNamespace,
					Generation: 1,
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
					Ports: []v1.ServicePort{
						{
							Name:       pycharmApp,
							TargetPort: intstr.FromInt32(3231),
							Port:       90,
							Protocol:   v1.ProtocolTCP,
							NodePort:   32131,
						},
					},
				},
			},
		}
		secrets = []*v1.Secret{
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Secret",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      secret1,
					Namespace: defaultNamespace,
				},
				Data: map[string][]byte{
					"key1": []byte("value1"),
					"key2": []byte("value2"),
				},
			},
		}
	})
	Context("Job include services", func() {
		BeforeEach(func() {
			statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			workload = &job.JobWorkload{}
			aggr = resource.NewAggregator(workload)
		})
		Context("workload status", func() {
			Context("when services status are nil", func() {
				It("should update the workload status with the services", func() {

					statusCalculator.WithServices(services)
					statusCalculator.WithSecrets(secrets)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
					Expect(workload.Status.ServiceStatuses).To(HaveLen(2))
					servieStatuses := workload.Status.ServiceStatuses
					// sort
					sort.Slice(servieStatuses, func(i, j int) bool {
						return strings.Compare(servieStatuses[i].Name, servieStatuses[j].Name) < 0
					})
					Expect(servieStatuses[0].Name).To(Equal(service1))
					Expect(servieStatuses[0].Type).To(Equal(v1.ServiceTypeClusterIP))
					Expect(servieStatuses[0].Ports).To(HaveLen(1))
					Expect(servieStatuses[0].Ports[0].Name).To(Equal(httpPortName))
					Expect(servieStatuses[0].Ports[0].Port).To(Equal(int32(httpPort)))
					Expect(servieStatuses[0].Ports[0].Protocol).To(Equal(v1.ProtocolTCP))
					Expect(servieStatuses[0].Ports[0].TargetPort).To(Equal(intstr.FromInt(httpPort)))
					Expect(servieStatuses[1].Name).To(Equal(service2))
					Expect(servieStatuses[1].Type).To(Equal(v1.ServiceTypeNodePort))
					Expect(servieStatuses[1].Ports).To(HaveLen(1))
					Expect(servieStatuses[1].Ports[0].Name).To(Equal(pycharmApp))
					Expect(servieStatuses[1].Ports[0].Port).To(Equal(int32(32131)))
					Expect(servieStatuses[1].Ports[0].Protocol).To(Equal(v1.ProtocolTCP))

				})
			})
			Context("when the job does not exist", func() {
				It("should reset the status to Creating but preserve the cumulative RestartCount", func() {
					workload.Status.RestartCount = 1
					workload.Status.WorkloadCommonStatus.Phase = common.PhaseRestarting.String()

					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)

					Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseCreating.String()))
					Expect(workload.Status.RestartCount).To(Equal(int32(1)),
						"restart count must survive the transient job-absent status reset during a restart")
				})
			})
			Context("when services status are not nil", func() {
				BeforeEach(func() {
					workload = &job.JobWorkload{
						Status: job.JobStatus{
							NetworkStatus: common.NetworkStatus{
								ServiceStatuses: []common.ServiceStatus{
									{
										Name: service1,
										Type: v1.ServiceTypeClusterIP,
										Ports: []common.ExposePort{
											{
												Name:       httpPortName,
												Port:       9999,
												TargetPort: intstr.FromInt32(9999),
											},
										},
									},
								},
							},
						},
					}
					aggr = resource.NewAggregator(workload)
				})
				It("should update exiting status with the current services", func() {

					statusCalculator.WithServices(services)
					statusCalculator.WithSecrets(secrets)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
					Expect(workload.Status.ServiceStatuses).To(HaveLen(2))
					servieStatuses := workload.Status.ServiceStatuses
					// sort
					sort.Slice(servieStatuses, func(i, j int) bool {
						return strings.Compare(servieStatuses[i].Name, servieStatuses[j].Name) < 0
					})
					Expect(servieStatuses[0].Name).To(Equal(service1))
					Expect(servieStatuses[0].Type).To(Equal(v1.ServiceTypeClusterIP))
					Expect(servieStatuses[0].Ports).To(HaveLen(1))
					Expect(servieStatuses[0].Ports[0].Name).To(Equal(httpPortName))
					Expect(servieStatuses[0].Ports[0].Port).To(Equal(int32(httpPort)))
					Expect(servieStatuses[0].Ports[0].Protocol).To(Equal(v1.ProtocolTCP))
					Expect(servieStatuses[0].Ports[0].TargetPort).To(Equal(intstr.FromInt(httpPort)))
					Expect(servieStatuses[1].Name).To(Equal(service2))
					Expect(servieStatuses[1].Type).To(Equal(v1.ServiceTypeNodePort))
					Expect(servieStatuses[1].Ports).To(HaveLen(1))
					Expect(servieStatuses[1].Ports[0].Name).To(Equal(pycharmApp))
					Expect(servieStatuses[1].Ports[0].Port).To(Equal(int32(32131)))
					Expect(servieStatuses[1].Ports[0].Protocol).To(Equal(v1.ProtocolTCP))

				})
				It("should be empty service list stauts when services are nil", func() {
					Expect(workload.Status.ServiceStatuses).To(HaveLen(1))
					statusCalculator.WithServices(nil)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
					Expect(workload.Status.ServiceStatuses).To(BeEmpty())
				})
			})
		})
	})

	Context("No common resource", func() {
		Context("Job Finished Successfully", func() {

			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				jobAdapter.On("GetStatus").Return(status)
				setJobMock(status, jobAdapter, false, false, false, false, true, false, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				aggr = resource.NewAggregator(workload)
				createDegradedWorkload(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)

				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Ready condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Ready, JobComplete)
			})
			It("job urls is empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Condition Succeeded is True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Succeeded)
			})
			It("Condition Failed is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
			})
			It("Condition Suspended is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Suspended, "")
			})
			It("Condition Initialized is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")
			})
			It("Condition Pending is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
			})
			It("Degraded condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Degraded, "")
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
			It("Number of running pods should be 0", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of succeeded pods should be 1", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(1)))
			})
			It("Number of failed pods should be 0", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(0)))
			})
			It("Number of pending pods should be 0", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(0)))
			})
			It("Number of allocated gpu memory should be 0", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
			})
			It("Number of allocated gpu should be 0", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("0"))
			})
			It("Time completed should be set", func() {
				Expect(workload.Status.CompletionTime).To(Equal(&timeHourBack))
			})
			It("Job start time should be nil", func() {
				Expect(workload.Status.StartTime).To(BeNil())
			})
		})
		Context("Job Failed", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, false, true, false, false, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				aggr = resource.NewAggregator(workload)
				createDegradedWorkload(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Succeeded condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
			})
			It("Failed condition should be True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Failed)
			})
			It("Suspended condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Suspended, "")
			})
			It("Degraded condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Degraded, "")
			})
			It("Pending condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
			})
			It("Initialized  condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")
			})
			It("Ready condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Ready, JobFailed)
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
			It("Job urls should be empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Number of running pods should be zero ", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of running pods should be 0", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of succeeded pods should be 0", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
			})
			It("Number of failed pods should be 1", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(1)))
			})
			It("Number of pending pods should be 0", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(0)))
			})
			It("Number of allocated gpu memory should be 0", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
			})
			It("Number of allocated gpu should be 0", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("0"))
			})
			It("Time completed should be set", func() {
				Expect(workload.Status.CompletionTime).To(Equal(&timeHourBack))
			})
			It("Job start time should be nil", func() {
				Expect(workload.Status.StartTime).To(BeNil())
			})
			It("phase message should be set", func() {
				Expect(workload.Status.WorkloadCommonStatus.PhaseMessage).To(ContainSubstring("Failed"))
			})
			Context("Job failed no failed message", func() {
				BeforeEach(func() {
					jobAdapter = new(mocks.JobAdapter)
					status := new(mocks.JobStatus)
					setJobWithFailedMsg(status, jobAdapter, false, false, false, true, false, false, false, false, "", nil)
					statusCalculator = NewStatusCalculator(mnnvlNotSupported)
					workload = &job.JobWorkload{}
					aggr = resource.NewAggregator(workload)
					createDegradedWorkload(workload)
					statusCalculator.WithJob(jobAdapter)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
				})
				It("phase message should be set", func() {
					Expect(workload.Status.WorkloadCommonStatus.PhaseMessage).To(Equal(GenericJobFailedMessage))
				})
			})
		})
		Context("Job Pending", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, true, false,
					false, false, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				aggr = resource.NewAggregator(workload)
				createDegradedWorkload(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Condition Succeeded is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
			})
			It("Condition Failed is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
			})
			It("Condition Suspended is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Suspended, "")
			})
			It("Condition Pending is True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Pending)
			})
			It("Condition Initialized is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")
			})
			It("Job start time should be nil", func() {
				Expect(workload.Status.StartTime).To(BeNil())
			})
			It("Ready condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Ready, JobNotRunning)
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
			It("Job urls should be empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Number of running pods should be 0 ", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of pending pods should be 1", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(1)))
			})
			It("Number of succeeded pods should be 0", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
			})
			It("Number of failed pods should be 0", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(0)))
			})
			It("Number of allocated gpu memory should be 0", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
			})
			It("Number of allocated gpu should be 0", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("0"))
			})
			It("Time completed should be nil", func() {
				Expect(workload.Status.CompletionTime.IsZero()).To(BeTrue())
			})
			It("Degraded condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Degraded, "")
			})

		})
		Context("Job Running", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false, false,
					false, false, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				aggr = resource.NewAggregator(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Condition Succeeded is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
			})
			It("Condition Failed is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
			})
			It("Condition Suspended is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Suspended, "")
			})
			It("Condition Pending is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
			})
			It("Condition Initialized is True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Initialized)
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
			It("Job should have 1 gpu", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("1"))
			})
			It("Job should have 1200 gpu memory", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(1200)))
			})
			It("Job start time should not nil", func() {
				Expect(workload.Status.StartTime.IsZero()).To(BeFalse())
			})
			It("Job start time should not change if the job is running", func() {
				startTime := workload.Status.StartTime.DeepCopy()
				statusCalculator.CalculateAll(context.TODO(), aggr)
				Expect(workload.Status.StartTime).To(Equal(startTime))

			})
			It("Ready condition should be True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Ready)
			})
			It("Job urls should be empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Number of running pods should be 1 ", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(1)))
			})
			It("Number of pending pods should be 0", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(0)))
			})
			It("Number of succeeded pods should be 0", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
			})
			It("Number of failed pods should be 0", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(0)))
			})
			It("Time completed should be nil", func() {
				Expect(workload.Status.CompletionTime.IsZero()).To(BeTrue())
			})

		})
		Context("Job Running with external urls", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false, false,
					false, false, false, false, map[string]string{constants.AnnotationJobUrl: "http://test.com, http://test2.com"})
				workload = &job.JobWorkload{}
				workload.Annotations = map[string]string{constants.AnnotationJobUrl: "http://test.com"}
				aggr = resource.NewAggregator(workload)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Job urls should be set", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(HaveLen(2))
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(ContainElement("http://test.com"))
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(ContainElement("http://test2.com"))
			})
		})
		Context("Job Initialized", func() {
			var status *mocks.JobStatus
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status = new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, false, false,
					false, false, true, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			})
			Context("Job is interactive", func() {
				BeforeEach(func() {
					workload = &job.JobWorkload{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindInteractiveWorkload,
						},
					}
					aggr = resource.NewAggregator(workload)
					statusCalculator = NewStatusCalculator(mnnvlNotSupported)
					createDegradedWorkload(workload)
					statusCalculator.WithJob(jobAdapter)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
				})
				It("Condition Initialized is false", func() {
					statusUtils.CheckConditionIsFalseOrUnsetWithMessageReason(workload.Status.Conditions, common.Initialized, InitializingSinglePodMessage, common.ReplicaNotReadyReason.String())
				})
			})
			Context("Job is training with 1 parallelism", func() {
				BeforeEach(func() {
					workload = &job.JobWorkload{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindTrainingWorkload,
						},
						Spec: job.JobSpec{
							LifeCycleFields: field_groups.LifeCycleFields{
								Parallelism: field_types.NewOptionalIntField(ptr.To(1)),
							},
						},
					}
					aggr = resource.NewAggregator(workload)
					statusCalculator = NewStatusCalculator(mnnvlNotSupported)
					createDegradedWorkload(workload)
					statusCalculator.WithJob(jobAdapter)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
				})
				It("Condition Initialized is fase", func() {
					statusUtils.CheckConditionIsFalseOrUnsetWithMessageReason(workload.Status.Conditions, common.Initialized, InitializingSinglePodMessage, common.ReplicaNotReadyReason.String())
				})
			})
			Context("Job is training with 2 parallelism", func() {
				BeforeEach(func() {
					workload = &job.JobWorkload{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindTrainingWorkload,
						},
						Spec: job.JobSpec{
							LifeCycleFields: field_groups.LifeCycleFields{
								Parallelism: field_types.NewOptionalIntField(ptr.To(2)),
							},
						},
					}
					aggr = resource.NewAggregator(workload)
					statusCalculator = NewStatusCalculator(mnnvlNotSupported)
					createDegradedWorkload(workload)
					statusCalculator.WithJob(jobAdapter)
					statusCalculator.WithWorkload(workload)
					statusCalculator.CalculateAll(context.TODO(), aggr)
				})
				It("Condition Initialized is fase", func() {
					statusUtils.CheckConditionIsFalseOrUnsetWithMessageReason(workload.Status.Conditions,
						common.Initialized, fmt.Sprintf(InitializingMultiPodsMessage, 1),
						common.MinimumReplicasNotReadyReason.String())
				})
			})
		})
		Context("Job Suspended", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, false, false,
					false, true, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				aggr = resource.NewAggregator(workload)
				createDegradedWorkload(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Condition Succeeded is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
			})
			It("Condition Initialized is False ", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")

			})
			It("Condition Failed is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
			})
			It("Condition Suspended is True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Suspended)
			})
			It("Condition Pending is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
			})
			It("Job should have 0 gpu", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("0"))
			})
			It("Job should have 0 gpu memory", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
			})
			It("Ready condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Ready, JobSuspended)
			})
			It("Job start time should be nil", func() {
				Expect(workload.Status.StartTime.IsZero()).To(BeTrue())
			})
			It("Job urls should be empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Number of running pods should be 0 ", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of pending pods should be 0", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(0)))
			})
			It("Number of succeeded pods should be 0", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
			})
			It("Number of failed pods should be 0", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(0)))
			})
			It("Time completed should be nil", func() {
				Expect(workload.Status.CompletionTime.IsZero()).To(BeTrue())
			})
			It("Degraded condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Degraded, "")
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
		})
		Context("Job Suspending", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, false, false,
					false, false, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindTrainingWorkload,
					},
					Spec: job.JobSpec{
						StateFields: field_groups.StateFields{
							Active: field_types.NewOptionalBoolField(ptr.To(false)),
						},
					},
				}
				aggr = resource.NewAggregator(workload)
				createDegradedWorkload(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Condition Succeeded is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
			})
			It("Condition Initialized is False ", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")

			})
			It("Condition Failed is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
			})
			It("Condition Suspended is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Suspended, "")
			})
			It("Condition Pending is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
			})
			It("Job should have 0 gpu", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("0"))
			})
			It("Job should have 0 gpu memory", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
			})
			It("Ready condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Ready, "workload is suspending")
			})
			It("Job start time should be nil", func() {
				Expect(workload.Status.StartTime.IsZero()).To(BeTrue())
			})
			It("Job urls should be empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Number of running pods should be 0 ", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of pending pods should be 0", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(0)))
			})
			It("Number of succeeded pods should be 0", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
			})
			It("Number of failed pods should be 0", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(0)))
			})
			It("Time completed should be nil", func() {
				Expect(workload.Status.CompletionTime.IsZero()).To(BeTrue())
			})
			It("Degraded condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Degraded, "")
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
			It("workload should be suspending", func() {
				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseSuspending.String()))
			})
		})
		Context("Job Resuming", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, false, false,
					false, true, false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindTrainingWorkload,
					},
					Spec: job.JobSpec{
						StateFields: field_groups.StateFields{
							Active: field_types.NewOptionalBoolField(ptr.To(true)),
						},
					},
				}
				aggr = resource.NewAggregator(workload)
				createDegradedWorkload(workload)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
				statusCalculator.CalculateAll(context.TODO(), aggr)
			})
			It("Condition Succeeded is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
			})
			It("Condition Initialized is False ", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")

			})
			It("Condition Failed is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
			})
			It("Condition Suspended is True", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.Conditions, common.Suspended)
			})
			It("Condition Pending is False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
			})
			It("Job should have 0 gpu", func() {
				Expect(workload.Status.TotalAllocatedGPU).To(Equal("0"))
			})
			It("Job should have 0 gpu memory", func() {
				Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
			})
			It("Ready condition should be False", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Ready, "job suspended")
			})
			It("Job start time should be nil", func() {
				Expect(workload.Status.StartTime.IsZero()).To(BeTrue())
			})
			It("Job urls should be empty", func() {
				Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
			})
			It("Number of running pods should be 0 ", func() {
				Expect(workload.Status.RunningPods).To(Equal(int32(0)))
			})
			It("Number of pending pods should be 0", func() {
				Expect(workload.Status.PendingPods).To(Equal(int32(0)))
			})
			It("Number of succeeded pods should be 0", func() {
				Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
			})
			It("Number of failed pods should be 0", func() {
				Expect(workload.Status.FailedPods).To(Equal(int32(0)))
			})
			It("Time completed should be nil", func() {
				Expect(workload.Status.CompletionTime.IsZero()).To(BeTrue())
			})
			It("Degraded condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Degraded, "")
			})
			It("MNNVL condition should be unset", func() {
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
			})
			It("workload should be resuming", func() {
				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseResuming.String()))
			})
		})
	})
	Context("Workload Types", func() {
		Context("Min Required Ready Replica", func() {

			// Table of test cases
			workloadTypes := []struct {
				description string
				kind        string
				expected    int32
			}{
				{
					description: "Interactive Workload",
					kind:        constants.KindInteractiveWorkload,
					expected:    1,
				},
				{
					description: "Training Workload",
					kind:        constants.KindTrainingWorkload,
					expected:    1,
				},
			}

			for _, tt := range workloadTypes {
				tt := tt // Capture range variable

				Context(tt.description, func() {
					BeforeEach(func() {
						jobAdapter = new(mocks.JobAdapter)
						status := new(mocks.JobStatus)
						statusCalculator = NewStatusCalculator(mnnvlNotSupported)
						workload = &job.JobWorkload{
							TypeMeta: metav1.TypeMeta{
								Kind: tt.kind,
							},
						}
						aggr = resource.NewAggregator(workload)

						setJobMock(status, jobAdapter, false, false, false, false,
							false, true, false, false, nil)
						statusCalculator.WithJob(jobAdapter)
						statusCalculator.WithWorkload(workload)
						statusCalculator.CalculateAll(context.TODO(), aggr)
					})

					It(fmt.Sprintf("should be %d", tt.expected), func() {
						Expect(workload.Status.MinRequiredReadyPods).To(Equal(tt.expected))
					})
				})
			}
		})
	})
	Context("Job is nil", func() {
		BeforeEach(func() {
			statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			workload = &job.JobWorkload{}
			aggr = resource.NewAggregator(workload)
			statusCalculator.WithJob(nil)
			statusCalculator.WithWorkload(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
		})

		It("Condition Succeeded is False", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Succeeded, "")
		})
		It("Condition Failed is False", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Failed, "")
		})
		It("Condition Suspended is False", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Suspended, "")
		})
		It("Condition Pending is False", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Pending, "")
		})
		It("Condition Initialized is False", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.Initialized, "")
		})
		It("Job should have empty gpu", func() {
			Expect(workload.Status.TotalAllocatedGPU).To(Equal(""))
		})
		It("Job should have 0 gpu memory", func() {
			Expect(workload.Status.TotalAllocatedGPUMemory).To(Equal(int64(0)))
		})
		It("Ready condition should be False", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, JobNotExistsMessage)
		})
		It("Job start time should be nil", func() {
			Expect(workload.Status.StartTime.IsZero()).To(BeTrue())
		})
		It("Job urls should be empty", func() {
			Expect(workload.Status.WorkloadCommonStatus.ExternalUrls).To(BeEmpty())
		})
		It("Number of running pods should be 0 ", func() {
			Expect(workload.Status.RunningPods).To(Equal(int32(0)))
		})
		It("Number of pending pods should be 0", func() {
			Expect(workload.Status.PendingPods).To(Equal(int32(0)))
		})
		It("Number of succeeded pods should be 0", func() {
			Expect(workload.Status.SucceededPods).To(Equal(int32(0)))
		})
		It("Number of failed pods should be 0", func() {
			Expect(workload.Status.FailedPods).To(Equal(int32(0)))
		})
		It("degraded condition should be unset", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
		})
		It("MNNVL condition should be unset", func() {
			statusUtils.CheckConditionIsFalseOrUnset(workload.Status.Conditions, common.GPUNetworkAccelerated, "")
		})
	})
	Context("Handle ready Condition", func() {
		var (
			ingressObjects    []k8sclient.Object
			routeObjects      []k8sclient.Object
			ExpectedResources []common.Resource
		)

		const defaultNamespace = "runai-test"
		ExpectedResources = []common.Resource{
			{
				Kind: metav1.GroupVersionKind{
					Group:   "",
					Version: "",
					Kind:    constants.KindService,
				},
				Name: service1,
			},
			{
				Kind: metav1.GroupVersionKind{
					Group:   "",
					Version: "",
					Kind:    constants.KindService,
				},
				Name: service2,
			},
			{Kind: metav1.GroupVersionKind{
				Kind: constants.KindIngress,
			},
				Name: "test-ingress-1",
			},
			{Kind: metav1.GroupVersionKind{
				Kind: constants.KindIngress,
			},
				Name: "test-ingress-2",
			},
			{Kind: metav1.GroupVersionKind{
				Kind: constants.KindRoute,
			},
				Name: "test-route-1",
			},
			{Kind: metav1.GroupVersionKind{
				Kind: constants.KindRoute,
			},
				Name: "test-route-2",
			},
		}
		ingress := []*network.Ingress{{
			TypeMeta: metav1.TypeMeta{
				Kind: "Ingress",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-ingress-1",
				Namespace: defaultNamespace,
			},
		},
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Ingress",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress-2",
					Namespace: defaultNamespace,
				},
			},
		}
		for _, ing := range ingress {
			ingressObjects = append(ingressObjects, ing)
		}
		route := []*routev1.Route{
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Route",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-route-1",
					Namespace: defaultNamespace,
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind: "Route",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-route-2",
					Namespace: defaultNamespace,
				},
			},
		}
		for _, r := range route {
			routeObjects = append(routeObjects, r)
		}

		Context("Common Resources match expected", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false,
					false, false, false,
					false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				statusCalculator.WithExpectedResources(ExpectedResources)
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithServices(services)
				statusCalculator.WithResources(ingressObjects)
				statusCalculator.WithResources(routeObjects)
				statusCalculator.WithWorkload(workload)
				statusCalculator.handleReadyCondition(context.Background())
			})
			It("Should set ready condition to true", func() {
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Ready)
			})

		})
		Context("Common Resources do not match expected, ready condition false ", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false,
					false, false, false,
					false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				statusCalculator.WithServices(services)
				statusCalculator.WithSecrets(secrets)
				statusCalculator.WithResources(ingressObjects)
				statusCalculator.WithResources(routeObjects)
				statusCalculator.WithWorkload(workload)
				statusCalculator.WithJob(jobAdapter)

			})
			It("Service 3 does not exists", func() {
				newExpectedResources := append(ExpectedResources, common.Resource{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindService,
					},
					Name: "test-service-3",
				})
				statusCalculator.WithExpectedResources(newExpectedResources)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "service test-service-3 does not exists")
			})
			It("Ingress 3 does not exists", func() {
				newExpectedResources := append(ExpectedResources, common.Resource{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindIngress,
					},
					Name: "test-ingress-3",
				})
				statusCalculator.WithExpectedResources(newExpectedResources)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "ingress test-ingress-3 does not exists")

			})
			It("Route 3 does not exists", func() {
				newExpectedResources := append(ExpectedResources, common.Resource{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindRoute,
					},
					Name: "test-route-3",
				})
				statusCalculator.WithExpectedResources(newExpectedResources)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "route test-route-3 does not exists")
			})
			It("Secret 2 does not exists", func() {
				newExpectedResources := append(ExpectedResources, common.Resource{
					Kind: metav1.GroupVersionKind{
						Kind: constants.KindSecret,
					},
					Name: "test-secret-2",
				})
				statusCalculator.WithExpectedResources(newExpectedResources)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "secret test-secret-2 does not exists")
			})
		})
	})
	Context("Handle ready Condition with HTTPRoute", func() {
		Context("HTTPRoute exists in expected resources", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false,
					false, false, false,
					false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
			})

			It("Should set ready condition to true when HTTPRoute exists", func() {
				httpRouteObjects := []k8sclient.Object{
					&gatewayv1.HTTPRoute{
						TypeMeta:   metav1.TypeMeta{Kind: constants.KindHTTPRoute},
						ObjectMeta: metav1.ObjectMeta{Name: "test-httproute-1", Namespace: defaultNamespace},
					},
				}
				expectedRes := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: "test-httproute-1"},
				}
				statusCalculator.WithExpectedResources(expectedRes)
				statusCalculator.WithResources(httpRouteObjects)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Ready)
			})

			It("Should set ready condition to false when HTTPRoute does not exist", func() {
				expectedRes := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: "missing-httproute"},
				}
				statusCalculator.WithExpectedResources(expectedRes)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "httproute missing-httproute does not exists")
			})
		})

		Context("Dual-mode: Ingress and HTTPRoute expected", func() {
			BeforeEach(func() {
				jobAdapter = new(mocks.JobAdapter)
				status := new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false,
					false, false, false,
					false, false, nil)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				statusCalculator.WithJob(jobAdapter)
				statusCalculator.WithWorkload(workload)
			})

			It("Should set ready condition to true when both Ingress and HTTPRoute exist", func() {
				ingressObj := []k8sclient.Object{
					&network.Ingress{
						TypeMeta:   metav1.TypeMeta{Kind: constants.KindIngress},
						ObjectMeta: metav1.ObjectMeta{Name: "dual-ingress", Namespace: defaultNamespace},
					},
				}
				httpRouteObj := []k8sclient.Object{
					&gatewayv1.HTTPRoute{
						TypeMeta:   metav1.TypeMeta{Kind: constants.KindHTTPRoute},
						ObjectMeta: metav1.ObjectMeta{Name: "dual-httproute", Namespace: defaultNamespace},
					},
				}
				expectedRes := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}, Name: "dual-ingress"},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: "dual-httproute"},
				}
				statusCalculator.WithExpectedResources(expectedRes)
				statusCalculator.WithResources(ingressObj)
				statusCalculator.WithResources(httpRouteObj)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Ready)
			})

			It("Should set ready condition to false when HTTPRoute is missing in dual-mode", func() {
				ingressObj := []k8sclient.Object{
					&network.Ingress{
						TypeMeta:   metav1.TypeMeta{Kind: constants.KindIngress},
						ObjectMeta: metav1.ObjectMeta{Name: "dual-ingress", Namespace: defaultNamespace},
					},
				}
				expectedRes := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}, Name: "dual-ingress"},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: "dual-httproute"},
				}
				statusCalculator.WithExpectedResources(expectedRes)
				statusCalculator.WithResources(ingressObj)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "httproute dual-httproute does not exists")
			})

			It("Should set ready condition to false when Ingress is missing in dual-mode", func() {
				httpRouteObj := []k8sclient.Object{
					&gatewayv1.HTTPRoute{
						TypeMeta:   metav1.TypeMeta{Kind: constants.KindHTTPRoute},
						ObjectMeta: metav1.ObjectMeta{Name: "dual-httproute", Namespace: defaultNamespace},
					},
				}
				expectedRes := []common.Resource{
					{Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}, Name: "dual-ingress"},
					{Kind: metav1.GroupVersionKind{Kind: constants.KindHTTPRoute}, Name: "dual-httproute"},
				}
				statusCalculator.WithExpectedResources(expectedRes)
				statusCalculator.WithResources(httpRouteObj)
				statusCalculator.handleReadyCondition(context.Background())
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, "ingress dual-ingress does not exists")
			})
		})

		Context("WithResources dispatches HTTPRoute correctly", func() {
			BeforeEach(func() {
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			})

			It("Should add HTTPRoute to internal map via WithResources", func() {
				hr := &gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{Name: "hr-dispatch"},
				}
				statusCalculator.WithResources([]k8sclient.Object{hr})
				Expect(statusCalculator.httpRoutes).To(HaveKey("hr-dispatch"))
			})

			It("Should clear HTTPRoute on Clear", func() {
				hr := &gatewayv1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{Name: "hr-clear"},
				}
				statusCalculator.WithResources([]k8sclient.Object{hr})
				Expect(statusCalculator.httpRoutes).To(HaveLen(1))
				statusCalculator.Clear()
				Expect(statusCalculator.httpRoutes).To(BeEmpty())
			})
		})
	})
	Context("When has issues", func() {
		var (
			pvc      *v1.PersistentVolumeClaim
			ingress  *network.Ingress
			runaijob *runaiJobV1.RunaiJob

			resourceErr error

			issuesNonCritical        []*issue.Issue
			issuesCriticalNonCompute []*issue.Issue
			issuesCriticalCompute    []*issue.Issue

			statusCalcFn func(ctx context.Context, aggr *resource.Aggregator) job.JobStatus
		)

		BeforeEach(func() {
			jobAdapter = new(mocks.JobAdapter)
			statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			workload = &job.JobWorkload{}
			statusCalculator.WithWorkload(workload)
			statusCalculator.WithJob(jobAdapter)

			aggr = resource.NewAggregator(workload)

			pvc = &v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "pvc01"}}
			ingress = &network.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ingress01"}}
			runaijob = &runaiJobV1.RunaiJob{ObjectMeta: metav1.ObjectMeta{Name: "job01"}}

			resourceErr = errors.New("default err")

			issuesNonCritical = []*issue.Issue{
				issue2.NewIssue(pvc, resourceErr, false, ""),
				issue2.NewIssue(ingress, resourceErr, false, ""),
				issue2.NewIssue(runaijob, resourceErr, false, ""),
			}

			issuesCriticalNonCompute = []*issue.Issue{
				issue2.NewIssue(pvc, resourceErr, true, ""),
				issue2.NewIssue(ingress, resourceErr, true, ""),
				issue2.NewIssue(runaijob, resourceErr, false, ""),
			}

			issuesCriticalCompute = []*issue.Issue{
				issue2.NewIssue(pvc, resourceErr, false, ""),
				issue2.NewIssue(ingress, resourceErr, true, ""),
				issue2.NewIssue(runaijob, resourceErr, true, ""),
			}
		})

		DescribeTableSubtree("Calculate status with issues",
			func(calculateAll bool) {
				BeforeEach(func() {
					if calculateAll {
						statusCalcFn = statusCalculator.CalculateAll
					} else {
						statusCalcFn = statusCalculator.CalculatePhaseAndConditions
					}
				})

				Context("Pre Running phase", func() {
					BeforeEach(func() {
						status := new(mocks.JobStatus)
						setJobMock(status, jobAdapter, false, false, false,
							false, false, false,
							true, false, nil)

						calculated := statusCalculator.CalculateAll(context.Background(), aggr)
						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseInitializing.String()))
					})

					It("non-critical issue", func() {
						aggr.AddIssues(issuesNonCritical...)
						calculated := statusCalcFn(context.Background(), aggr)

						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseInitializing.String())) // non-critical when not running should not affect

						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, WorkloadHasIssuesMessage)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, "")
					})
					It("critical non-compute issue", func() {
						aggr.AddIssues(issuesCriticalNonCompute...)
						calculated := statusCalcFn(context.Background(), aggr)

						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String())) // critical when not running should fail workload
						Expect(calculated.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to Network, Storage issues"))

						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Failed)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, WorkloadHasIssuesMessage)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, "")
					})
					It("critical compute issue", func() {
						aggr.AddIssues(issuesCriticalCompute...)
						calculated := statusCalcFn(context.Background(), aggr)

						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String())) // critical when not running should fail workload
						Expect(calculated.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to Compute, Network issues"))

						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Failed)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, WorkloadHasIssuesMessage)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
					})
				})
				Context("In Running phase", func() {
					BeforeEach(func() {
						status := new(mocks.JobStatus)
						setJobMock(status, jobAdapter, false, true, false,
							false, false, false,
							false, false, nil)

						calculated := statusCalculator.CalculateAll(context.Background(), aggr)
						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
					})

					It("non-critical issue", func() {
						aggr.AddIssues(issuesNonCritical...)
						calculated := statusCalcFn(context.Background(), aggr)

						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String())) // non-critical when running should not affect

						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, WorkloadHasIssuesMessage)
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
					})
					It("critical non-compute issue", func() {
						aggr.AddIssues(issuesCriticalNonCompute...)
						calculated := statusCalcFn(context.Background(), aggr)

						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseDegraded.String())) // critical when running should make workload degraded
						Expect(calculated.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload is in degraded state due to Network, Storage issues"))

						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, WorkloadHasIssuesMessage)
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
					})
					It("critical compute issue", func() {
						aggr.AddIssues(issuesCriticalCompute...)
						calculated := statusCalcFn(context.Background(), aggr)

						Expect(calculated).NotTo(BeNil())
						Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String())) // critical compute when running should fail workload
						Expect(calculated.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to Compute, Network issues"))

						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Failed)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, WorkloadHasIssuesMessage)
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
						statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded)
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, "")
						statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
					})
				})
			},
			Entry("CalculateAll", true),
			Entry("CalculatePhaseAndConditions", false),
		)

		Context("issue groups in phase message are sorted", MustPassRepeatedly(5), func() {
			var (
				issues []*issue.Issue
			)

			BeforeEach(func() {
				pvc2 := &v1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "pvc02"}}
				ingress2 := &network.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ingress02"}}
				service := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "pvc02"}}

				issues = []*issue.Issue{
					issue2.NewIssue(pvc, resourceErr, true, ""),
					issue2.NewIssue(ingress, resourceErr, true, ""),
					issue2.NewIssue(pvc2, resourceErr, true, ""),
					issue2.NewIssue(ingress2, resourceErr, true, ""),
					issue2.NewIssue(service, resourceErr, true, ""),
					issue2.NewIssue(workload, resourceErr, true, ""),
				}

				rand.New(rand.NewSource(time.Now().UnixNano()))
				rand.Shuffle(len(issues), func(i, j int) {
					issues[i], issues[j] = issues[j], issues[i]
				})

				aggr.AddIssues(issues...)
			})

			It("failed", func() {
				setJobMock(new(mocks.JobStatus), jobAdapter, false, false, false,
					false, false, false,
					true, false, nil)

				calculated := statusCalculator.CalculatePhaseAndConditions(context.Background(), aggr)
				Expect(calculated).NotTo(BeNil())

				Expect(calculated.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload failed due to Internal, Network, Storage issues"))
			})
			It("degraded", func() {
				setJobMock(new(mocks.JobStatus), jobAdapter, false, true, false,
					false, false, false,
					false, false, nil)

				calculated := statusCalculator.CalculatePhaseAndConditions(context.Background(), aggr)
				Expect(calculated).NotTo(BeNil())

				Expect(calculated.WorkloadCommonStatus.PhaseMessage).To(Equal("Workload is in degraded state due to Internal, Network, Storage issues"))
			})
		})

		Context("CompletionTime", func() {
			var (
				status *mocks.JobStatus
			)
			BeforeEach(func() {
				status = new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, true, false,
					false, false, false,
					false, false, nil)

				Expect(status.GetCompletionTime()).To(BeNil())
				Expect(workload.Status.CompletionTime).To(BeNil())

				calculated := statusCalculator.CalculateAll(context.Background(), aggr)
				Expect(calculated).NotTo(BeNil())
				Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))

				Expect(status.GetCompletionTime()).To(BeNil())
				Expect(workload.Status.CompletionTime).To(BeNil())
			})

			// this test checks that if the workload is failed by the workload (a.k.a issues)
			// the completion time is not overridden by the job's completion time
			It("workload CompletionTime is not set by the job's if job not in finite state", func() {
				aggr.AddIssues(issuesCriticalCompute...)
				workload.Status.CompletionTime = &timeHourBack // in other circumstances, we would have set it in the reconcile

				// run twice - make sure idempotent
				for i := 0; i < 2; i++ {
					calculated := statusCalculator.CalculateAll(context.Background(), aggr)
					Expect(calculated).NotTo(BeNil())
					Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))

					Expect(status.GetCompletionTime()).To(BeNil())
					Expect(workload.Status.CompletionTime).NotTo(BeNil())
				}
			})

			It("workload CompletionTime is set by the job's if job is in finite state", func() {
				jobAdapter = new(mocks.JobAdapter)
				statusCalculator = NewStatusCalculator(mnnvlNotSupported)
				workload = &job.JobWorkload{}
				statusCalculator.WithWorkload(workload)
				statusCalculator.WithJob(jobAdapter)

				aggr = resource.NewAggregator(workload)

				status = new(mocks.JobStatus)
				setJobMock(status, jobAdapter, false, false, false,
					false, true, false,
					false, false, nil)

				aggr.AddIssues(issuesCriticalCompute...)

				// run twice - make sure idempotent
				for i := 0; i < 2; i++ {
					calculated := statusCalculator.CalculateAll(context.Background(), aggr)
					Expect(calculated).NotTo(BeNil())
					Expect(calculated.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))

					Expect(status.GetCompletionTime()).NotTo(BeNil())
					Expect(workload.Status.CompletionTime).NotTo(BeNil())
				}
			})
		})
	})
	Context("Failed by workload", func() {
		BeforeEach(func() {
			statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			workload = &job.JobWorkload{}
			makeWorkloadFailedByWorkload(workload)
			aggr = resource.NewAggregator(workload)

			statusCalculator.WithJob(nil)
			statusCalculator.WithWorkload(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
		})

		It("phase is not calculated again", func() {
			Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseFailed.String()))
			Expect(workload.Status.WorkloadCommonStatus.Conditions).To(HaveLen(1)) //indicates we suspended early
		})
	})
	Context("MNNVL enabled in cluster", func() {
		var (
			computeDomain *nvapi.ComputeDomain
		)
		ctx := context.Background()

		BeforeEach(func() {
			computeDomain = &nvapi.ComputeDomain{
				TypeMeta: metav1.TypeMeta{
					Kind: nvapi.ComputeDomainKind,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cd",
					Namespace: defaultNamespace,
				},
				Status: nvapi.ComputeDomainStatus{},
			}

			jobAdapter = new(mocks.JobAdapter)
			status := new(mocks.JobStatus)
			setJobMock(status, jobAdapter, false, true, false, false,
				false, false, false, false, nil)
			statusCalculator = NewStatusCalculator(mnnvlSupported)
			workload = &job.JobWorkload{}
			aggr = resource.NewAggregator(workload)
			statusCalculator.WithJob(jobAdapter)
			statusCalculator.WithWorkload(workload)

			Expect(computeDomain).NotTo(BeNil())
		})

		Context("workload is not accelerated", func() {
			It("workload is Running", func() {
				statusCalculator.CalculateAll(ctx, aggr)

				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, MnnvlNotRequestedMessage)

				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Ready)
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
			})
		})

		Context("workload is accelerated", func() {
			BeforeEach(func() {
				enabledStatus := mnnvlutils.Status{
					MasterMode: mnnvl.ModeNone,
					WorkerMode: mnnvl.ModeRequired,
				}
				ctx = mnnvlutils.ToContext(ctx, enabledStatus)
			})

			It("no compute domain - workload is Running, not Ready", func() {
				statusCalculator.CalculateAll(ctx, aggr)

				//until runtime status management of compute domain is in place, the phase will not be affected, only the conditions
				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, MnnvlNotReadyMessage)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, MnnvlNotReadyMessage)

				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
			})
			It("with compute domain not Ready - workload is Running, not Ready", func() {
				computeDomain.Status.Status = constants.ComputeDomainStatusNotReady
				statusCalculator.WithComputeDomain(computeDomain)
				statusCalculator.CalculateAll(ctx, aggr)

				//until runtime status management of compute domain is in place, the phase will not be affected, only the conditions
				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, MnnvlNotReadyMessage)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, MnnvlNotReadyMessage)

				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
			})
			It("with compute domain Ready but deleted - workload is Running, not Ready", func() {
				computeDomain.DeletionTimestamp = &metav1.Time{Time: time.Now()}
				computeDomain.Status.Status = constants.ComputeDomainStatusReady
				statusCalculator.WithComputeDomain(computeDomain)
				statusCalculator.CalculateAll(ctx, aggr)

				//until runtime status management of compute domain is in place, the phase will not be affected, only the conditions
				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated, MnnvlNotReadyMessage)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Ready, MnnvlNotReadyMessage)

				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
			})
			It("with compute domain Ready - workload is Running, Ready", func() {
				computeDomain.Status.Status = constants.ComputeDomainStatusReady
				statusCalculator.WithComputeDomain(computeDomain)
				statusCalculator.CalculateAll(ctx, aggr)

				Expect(workload.Status.WorkloadCommonStatus.Phase).To(Equal(common.PhaseRunning.String()))
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.GPUNetworkAccelerated)
				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Ready)

				statusUtils.CheckConditionIsTrue(workload.Status.WorkloadCommonStatus.Conditions, common.Initialized)
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Suspended, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Succeeded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Failed, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, "")
				statusUtils.CheckConditionIsFalseOrUnset(workload.Status.WorkloadCommonStatus.Conditions, common.Pending, "")
			})
		})
	})
	Context("LoadBalancer Service Status", func() {
		var (
			loadBalancerService *v1.Service
		)
		BeforeEach(func() {
			statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			workload = &job.JobWorkload{}
			loadBalancerService = &v1.Service{
				TypeMeta: metav1.TypeMeta{
					Kind: "Service",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "lb-service",
					Namespace: defaultNamespace,
				},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeLoadBalancer,
					Ports: []v1.ServicePort{
						{Name: "a", Port: 80, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
						{Name: "b", Port: 443, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8443)},
					},
				},
				Status: v1.ServiceStatus{
					LoadBalancer: v1.LoadBalancerStatus{
						Ingress: []v1.LoadBalancerIngress{
							{IP: "2.2.2.2"},
							{Hostname: "host-b"},
							{IP: "1.1.1.1", Hostname: "host-a"},
						},
					},
				},
			}
		})
		It("should expose sorted ports and sorted ingress in ServiceStatus", func() {
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Sort ports by Port (still a map list)
			sortedPorts := append([]common.ExposePort(nil), ss.Ports...)
			sort.Slice(sortedPorts, func(i, j int) bool { return sortedPorts[i].Port < sortedPorts[j].Port })
			Expect(sortedPorts).To(HaveLen(2))
			Expect(sortedPorts[0].Port).To(Equal(int32(80)))
			Expect(sortedPorts[1].Port).To(Equal(int32(443)))
			// Ingress: order does not matter, use ConsistOf
			Expect(ss.LoadBalancer.Ingress).To(ConsistOf(
				And(HaveField("IP", "1.1.1.1"), HaveField("Hostname", "host-a")),
				And(HaveField("IP", "2.2.2.2"), HaveField("Hostname", "")),
				And(HaveField("IP", ""), HaveField("Hostname", "host-b")),
			))
		})
		It("should handle nil and empty ingress and ports", func() {
			// No ingress, no ports
			loadBalancerService.Spec.Ports = nil
			loadBalancerService.Status.LoadBalancer.Ingress = nil
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			Expect(ss.Ports).To(BeEmpty())
			Expect(ss.LoadBalancer.Ingress).To(BeEmpty())
		})
		It("should produce consistent output regardless of input order", func() {
			// Shuffle ports and ingress
			loadBalancerService.Spec.Ports = []v1.ServicePort{
				{Name: "b", Port: 443, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8443)},
				{Name: "a", Port: 80, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
			}
			loadBalancerService.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
				{Hostname: "host-b"},
				{IP: "2.2.2.2"},
				{IP: "1.1.1.1", Hostname: "host-a"},
			}
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Sort ports by Port
			sortedPorts := append([]common.ExposePort(nil), ss.Ports...)
			sort.Slice(sortedPorts, func(i, j int) bool { return sortedPorts[i].Port < sortedPorts[j].Port })
			Expect(sortedPorts).To(HaveLen(2))
			Expect(sortedPorts[0].Port).To(Equal(int32(80)))
			Expect(sortedPorts[1].Port).To(Equal(int32(443)))
			// Ingress: order does not matter, use ConsistOf
			Expect(ss.LoadBalancer.Ingress).To(ConsistOf(
				And(HaveField("IP", "1.1.1.1"), HaveField("Hostname", "host-a")),
				And(HaveField("IP", "2.2.2.2"), HaveField("Hostname", "")),
				And(HaveField("IP", ""), HaveField("Hostname", "host-b")),
			))
		})
		It("should handle nil Service pointer gracefully in buildServiceStatus", func() {
			ss := statusCalculator.buildServiceStatus(context.Background(), nil)
			Expect(ss.Name).To(Equal(""))
			Expect(ss.Ports).To(BeEmpty())
			Expect(ss.LoadBalancer).To(BeNil())
		})
		It("should handle nil and empty in buildServicePorts and buildLoadBalancerStatus", func() {
			Expect(buildServicePorts(nil)).To(BeEmpty())
			lb := buildLoadBalancerStatus(context.Background(), nil)
			Expect(lb).To(BeNil())
			svc := &v1.Service{}
			Expect(buildServicePorts(svc)).To(BeEmpty())
			// For a Service with no type, LoadBalancerStatus should be nil
			lb2 := buildLoadBalancerStatus(context.Background(), svc)
			Expect(lb2).To(BeNil())

			// For a LoadBalancer service with nil or empty ingress, should return non-nil with empty Ingress
			lbSvc := &v1.Service{Spec: v1.ServiceSpec{Type: v1.ServiceTypeLoadBalancer}}
			lb3 := buildLoadBalancerStatus(context.Background(), lbSvc)
			Expect(lb3).NotTo(BeNil())
			Expect(lb3.Ingress).To(BeEmpty())

			lbSvc2 := &v1.Service{Spec: v1.ServiceSpec{Type: v1.ServiceTypeLoadBalancer}, Status: v1.ServiceStatus{LoadBalancer: v1.LoadBalancerStatus{Ingress: nil}}}
			lb4 := buildLoadBalancerStatus(context.Background(), lbSvc2)
			Expect(lb4).NotTo(BeNil())
			Expect(lb4.Ingress).To(BeEmpty())
		})
		It("should ignore ingress with empty IP and Hostname", func() {
			loadBalancerService.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
				{IP: "1.1.1.1"},
				{}, // Both IP and Hostname empty
				{Hostname: "host-b"},
			}
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Only two ingress entries should be present
			Expect(ss.LoadBalancer.Ingress).To(HaveLen(2))
			Expect(ss.LoadBalancer.Ingress[0].IP).To(Equal("1.1.1.1"))
			Expect(ss.LoadBalancer.Ingress[1].Hostname).To(Equal("host-b"))
		})
		It("should set Name field to IP if present, otherwise Hostname", func() {
			loadBalancerService.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
				{IP: "1.1.1.1"},
				{Hostname: "host-b"},
				{IP: "2.2.2.2", Hostname: "host-c"},
				{}, // Both empty, should be ignored
			}
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Ingress: order does not matter, use ConsistOf
			Expect(ss.LoadBalancer.Ingress).To(ConsistOf(
				HaveField("IP", "1.1.1.1"),
				HaveField("IP", "2.2.2.2"),
				HaveField("Hostname", "host-b"),
			))
		})
		It("should expose correct ingress entries regardless of order", func() {
			loadBalancerService.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
				{IP: "2.2.2.2"},
				{Hostname: "host-b"},
				{IP: "1.1.1.1", Hostname: "host-a"},
			}
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Ingress entries should be present, order does not matter
			Expect(ss.LoadBalancer.Ingress).To(ConsistOf(
				HaveField("IP", "1.1.1.1"),
				HaveField("IP", "2.2.2.2"),
				HaveField("Hostname", "host-b"),
			))
		})
		It("should expose sorted ports and sorted ingress in ServiceStatus (order agnostic)", func() {
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Ports: order does not matter
			Expect(ss.Ports).To(ConsistOf(
				HaveField("Port", int32(80)),
				HaveField("Port", int32(443)),
			))
			// Ingress: order does not matter
			Expect(ss.LoadBalancer.Ingress).To(ConsistOf(
				And(HaveField("IP", "1.1.1.1"), HaveField("Hostname", "host-a")),
				And(HaveField("IP", "2.2.2.2"), HaveField("Hostname", "")),
				And(HaveField("IP", ""), HaveField("Hostname", "host-b")),
			))
		})
		It("should produce consistent output regardless of input order (order agnostic)", func() {
			// Shuffle ports and ingress
			loadBalancerService.Spec.Ports = []v1.ServicePort{
				{Name: "b", Port: 443, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8443)},
				{Name: "a", Port: 80, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
			}
			loadBalancerService.Status.LoadBalancer.Ingress = []v1.LoadBalancerIngress{
				{Hostname: "host-b"},
				{IP: "2.2.2.2"},
				{IP: "1.1.1.1", Hostname: "host-a"},
			}
			statusCalculator.WithServices([]*v1.Service{loadBalancerService})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses[0]
			// Ports: order does not matter
			Expect(ss.Ports).To(ConsistOf(
				HaveField("Port", int32(80)),
				HaveField("Port", int32(443)),
			))
			// Ingress: order does not matter
			Expect(ss.LoadBalancer.Ingress).To(ConsistOf(
				And(HaveField("IP", "1.1.1.1"), HaveField("Hostname", "host-a")),
				And(HaveField("IP", "2.2.2.2"), HaveField("Hostname", "")),
				And(HaveField("IP", ""), HaveField("Hostname", "host-b")),
			))
		})
	})
	Context("Service Status (General)", func() {
		var (
			serviceA *v1.Service
			serviceB *v1.Service
		)
		BeforeEach(func() {
			statusCalculator = NewStatusCalculator(mnnvlNotSupported)
			workload = &job.JobWorkload{}
			serviceA = &v1.Service{
				TypeMeta:   metav1.TypeMeta{Kind: "Service"},
				ObjectMeta: metav1.ObjectMeta{Name: "svc-a", Namespace: defaultNamespace},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeClusterIP,
					Ports: []v1.ServicePort{
						{Name: "z", Port: 8080, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(80)},
						{Name: "a", Port: 80, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8080)},
					},
				},
			}
			serviceB = &v1.Service{
				TypeMeta:   metav1.TypeMeta{Kind: "Service"},
				ObjectMeta: metav1.ObjectMeta{Name: "svc-b", Namespace: defaultNamespace},
				Spec: v1.ServiceSpec{
					Type: v1.ServiceTypeNodePort,
					Ports: []v1.ServicePort{
						{Name: "b", Port: 443, Protocol: v1.ProtocolTCP, TargetPort: intstr.FromInt(8443), NodePort: 32443},
					},
				},
			}
		})
		It("should expose sorted ports and correct fields for all service types", func() {
			statusCalculator.WithServices([]*v1.Service{serviceB, serviceA}) // out of order
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses
			Expect(ss).To(HaveLen(2))
			// Should contain both service names, order does not matter
			Expect(ss).To(ConsistOf(
				HaveField("Name", "svc-a"),
				HaveField("Name", "svc-b"),
			))
			// Ports for svc-a: should contain both ports, order does not matter
			var portsA, portsB []common.ExposePort
			for _, s := range ss {
				if s.Name == "svc-a" {
					portsA = s.Ports
				}
				if s.Name == "svc-b" {
					portsB = s.Ports
				}
			}
			Expect(portsA).To(ConsistOf(
				HaveField("Port", int32(80)),
				HaveField("Port", int32(8080)),
			))
			Expect(portsB).To(ConsistOf(
				HaveField("Port", int32(32443)),
			))
		})
		It("should handle nil and empty ports", func() {
			serviceA.Spec.Ports = nil
			serviceB.Spec.Ports = nil
			statusCalculator.WithServices([]*v1.Service{serviceA, serviceB})
			statusCalculator.WithWorkload(workload)
			aggr = resource.NewAggregator(workload)
			statusCalculator.CalculateAll(context.TODO(), aggr)
			ss := workload.Status.ServiceStatuses
			Expect(ss[0].Ports).To(BeEmpty())
			Expect(ss[1].Ports).To(BeEmpty())
		})
	})
})

func makeWorkloadFailedByWorkload(workload *job.JobWorkload) {
	workload.Status.WorkloadCommonStatus.Phase = common.PhaseFailed.String()

	workload.Status.WorkloadCommonStatus.Conditions = []common.WorkloadCondition{}
	condition.UpsertCondition(&workload.Status.WorkloadCommonStatus.Conditions, common.Failed, v1.ConditionTrue, common.WorkloadBuildFailedReason, "")
}

func createDegradedWorkload(workload *job.JobWorkload) {
	workload.Status.WorkloadCommonStatus.Conditions = []common.WorkloadCondition{}
	condition.UpsertCondition(&workload.Status.WorkloadCommonStatus.Conditions, common.Degraded, v1.ConditionTrue, "degraded", "degraded")
}

func setJobMock(status *mocks.JobStatus, jobAdapter *mocks.JobAdapter, isCreating, isRunning, isPending, isJobFailed, isSucceeded, isSuspended, isInitializing, isTerminating bool, annotations map[string]string) {
	setJobWithFailedMsg(status, jobAdapter, isCreating, isRunning, isPending, isJobFailed, isSucceeded, isSuspended, isInitializing, isTerminating, "Failed", annotations)
}

func setJobWithFailedMsg(status *mocks.JobStatus, jobAdapter *mocks.JobAdapter, isCreating bool, isRunning bool, isPending bool, isJobFailed bool, isSucceeded bool, isSuspended bool, isInitializing bool, isTerminating bool, faildMessage string, annotations map[string]string) {
	jobAdapter.On("GetStatus").Return(status)
	status.On("IsCreating").Return(isCreating)
	status.On("IsJobSucceeded").Return(isSucceeded)
	status.On("IsRunning").Return(isRunning)
	status.On("IsPending").Return(isPending)
	status.On("IsJobFailed").Return(isJobFailed)
	status.On("IsTerminating").Return(isTerminating)

	var finished bool
	if isJobFailed || isSucceeded {
		finished = true
		status.On("GetCompletionTime").Return(&timeHourBack)
	} else {
		status.On("GetCompletionTime").Return(nil)
	}

	status.On("IsJobFinished").Return(finished)
	status.On("IsSuspended").Return(isSuspended)
	status.On("GetSuspendInfo").Return("")
	status.On("IsInitializing").Return(isInitializing)
	if isRunning {
		status.On("GetAllocatedGPUMemory").Return(int64(1200))
		status.On("GetAllocatedGPU").Return("1")
		status.On("GetReadyPods").Return(int32(1))
		status.On("GetRunningPods").Return(int32(1))
	} else if isInitializing {
		status.On("GetAllocatedGPUMemory").Return(int64(0))
		status.On("GetAllocatedGPU").Return("0")
		status.On("GetReadyPods").Return(int32(0))
		status.On("GetRunningPods").Return(int32(1))

	} else {
		status.On("GetAllocatedGPUMemory").Return(int64(0))
		status.On("GetAllocatedGPU").Return("0")
		status.On("GetReadyPods").Return(int32(0))
		status.On("GetRunningPods").Return(int32(0))
	}

	if isPending {
		status.On("GetPendingPods").Return(int32(1))
	} else {
		status.On("GetPendingPods").Return(int32(0))
	}
	if isSucceeded {
		status.On("GetSucceededPods").Return(int32(1))
	} else {
		status.On("GetSucceededPods").Return(int32(0))
	}
	if isJobFailed {
		status.On("GetFailedPods").Return(int32(1))
		status.On("GetFailedInfo").Return(faildMessage)
	} else {
		status.On("GetFailedPods").Return(int32(0))
		status.On("GetFailedInfo").Return("")
	}
	jobAdapter.On("GetAnnotations").Return(annotations)
}

func setCondition(readyCondition bool, degradedCondition bool, time time.Time) []common.WorkloadCondition {
	var readyStatus = v1.ConditionTrue
	var degradedStatus = v1.ConditionTrue

	if !readyCondition {
		readyStatus = v1.ConditionFalse
	}

	if !degradedCondition {
		degradedStatus = v1.ConditionFalse
	}

	conditions := []common.WorkloadCondition{
		{
			Type:   common.Ready,
			Status: readyStatus,
		},
	}

	if degradedCondition {
		conditions = append(conditions, common.WorkloadCondition{
			Type:               common.Degraded,
			Status:             degradedStatus,
			LastTransitionTime: metav1.Time{Time: time},
		})
	}

	return conditions
}
