/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managers

import (
	"context"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	openshiftV1 "github.com/openshift/api/route/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Runtime Manager Tests", func() {
	const (
		workloadName = "w1"
		workloadNs   = "runai-test"

		routeName = "r1"

		jobName       = "j1"
		jobFailureMsg = "I'm a failure"
	)
	var (
		srm       *SubResourcesRuntimeManager
		aggr      *resource.Aggregator
		ctx       = context.Background()
		scheme    = runtime.NewScheme()
		k8sClient k8sClient.Client
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		k8sClient = client.InitFakeClientWithObjectNoStatus(scheme)

		srm = NewSubResourcesRuntimeManager(k8sClient)

		training := &twv2.TrainingWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: workloadNs,
			},
		}
		jw := &jobs.JobWorkload{}
		training.CopyTo(jw)

		aggr = resource.NewAggregator(jw)
	})

	Context("general", func() {
		It("no resources - no issues", func() {
			issues := srm.CheckSubResources(ctx, aggr)
			Expect(issues).To(HaveLen(0))
		})
		It("no workload", func() {
			noWorkloadAggr := resource.NewAggregator(nil)
			issues := srm.CheckSubResources(ctx, noWorkloadAggr)
			Expect(issues).To(HaveLen(0))
		})
	})

	Context("Network Issues", func() {
		Context("Route", func() {
			It("url already exist", func() {
				r := getRouteWithError(routeName, RouteUrlAlreadyExistConditionReason)
				aggr.AddRoutes(r)

				issues := srm.CheckSubResources(ctx, aggr)
				Expect(issues).To(HaveLen(1))
				Expect(issues[0].GetGroup()).To(Equal(issue.Network))
				Expect(issues[0].Kind()).To(Equal(constants.KindRoute))
				Expect(issues[0].IsCritical()).To(BeTrue())
				Expect(issues[0].Message()).To(ContainSubstring(routeName))
				Expect(issues[0].Message()).To(ContainSubstring("A route with the same url already exist"))
			})
			It("non-url already exist", func() {
				r := getRouteWithError(routeName, "blabla")
				aggr.AddRoutes(r)

				issues := srm.CheckSubResources(ctx, aggr)
				Expect(issues).To(HaveLen(0))
			})
			It("ignores ingress", func() {
				i := &networkingv1.Ingress{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ingress1",
					},
				}
				aggr.AddIngresses(i)

				issues := srm.CheckSubResources(ctx, aggr)
				Expect(issues).To(HaveLen(0))
			})
		})

		Context("Load Balancer issues", func() {
			It("ingress for service, no events - should be ok", func() {
				r := getLoadBalancerServiceWithNoIngress("my-service", workloadNs, uuid.NewString())
				aggr.AddServices(&r)

				// there are currently no events, so we are good
				issues := srm.CheckSubResources(ctx, aggr)
				Expect(k8sClient.(*client.FakeClientWithEventUIDIndex).EventListCalls).To(Equal(1))
				Expect(issues).To(HaveLen(0))
			})

			It("ingress for service, there is event - should report issue", func() {
				serviceId := uuid.NewString()

				err := k8sClient.Create(context.Background(), &corev1.Event{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-event",
						Namespace: workloadNs,
					},
					InvolvedObject: corev1.ObjectReference{
						UID: types.UID(serviceId),
					},
					Reason: "AllocationFailed",
				})
				Expect(err).To(BeNil())

				r := getLoadBalancerServiceWithNoIngress("my-service", workloadNs, serviceId)
				aggr.AddServices(&r)

				issues := srm.CheckSubResources(ctx, aggr)

				Expect(k8sClient.(*client.FakeClientWithEventUIDIndex).EventListCalls).To(Equal(1))
				Expect(issues).To(HaveLen(1))
				Expect(issues[0].GetGroup()).To(Equal(issue.Network))
				Expect(issues[0].Kind()).To(Equal(constants.KindService))
				Expect(issues[0].IsCritical()).To(BeTrue())
				Expect(issues[0].Message()).To(ContainSubstring("[Service:my-service] No IP address is assigned to the service"))
			})

			It("ingress for service, there is an event but not allocation failed - no issue", func() {
				serviceId := uuid.NewString()

				err := k8sClient.Create(context.Background(), &corev1.Event{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-event",
						Namespace: workloadNs,
					},
					InvolvedObject: corev1.ObjectReference{
						UID: types.UID(serviceId),
					},
					Reason: "AllocationSucceeded",
				})
				Expect(err).To(BeNil())

				r := getLoadBalancerServiceWithNoIngress("my-service", workloadNs, serviceId)
				aggr.AddServices(&r)

				issues := srm.CheckSubResources(ctx, aggr)

				// there are currently no events, so we are good
				Expect(k8sClient.(*client.FakeClientWithEventUIDIndex).EventListCalls).To(Equal(1))
				Expect(issues).To(HaveLen(0))
			})

			It("there is ingress for the service - no issue and no event list called", func() {
				r := getLoadBalancerServiceWithNoIngress("my-service", workloadNs, uuid.NewString())
				r.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{
					{
						IP: "1.2.3.4",
					},
				}
				aggr.AddServices(&r)

				issues := srm.CheckSubResources(ctx, aggr)

				// no call to event list!
				Expect(k8sClient.(*client.FakeClientWithEventUIDIndex).EventListCalls).To(Equal(0))

				// and naturally, no issues
				Expect(issues).To(HaveLen(0))
			})

			It("this is not load balancer - no issue and no event list called", func() {
				r := getLoadBalancerServiceWithNoIngress("my-service", workloadNs, uuid.NewString())
				r.Spec.Type = constants.ServiceTypeNodePort
				aggr.AddServices(&r)

				issues := srm.CheckSubResources(ctx, aggr)

				// no call to event list!
				Expect(k8sClient.(*client.FakeClientWithEventUIDIndex).EventListCalls).To(Equal(0))

				// and naturally, no issues
				Expect(issues).To(HaveLen(0))
			})
		})
	})

	Context("Compute Issues", func() {
		It("RunaiJob failed job", func() {
			job := getRunaiJob(jobName, true, jobFailureMsg)
			aggr.SetJob(job)

			issues := srm.CheckSubResources(ctx, aggr)
			Expect(issues).To(HaveLen(1))
			Expect(issues[0].GetGroup()).To(Equal(issue.Compute))
			Expect(issues[0].Kind()).To(Equal(constants.KindRunaiJob))
			Expect(issues[0].IsCritical()).To(BeTrue())
			Expect(issues[0].Message()).To(ContainSubstring(jobName))
			Expect(issues[0].Message()).To(ContainSubstring(jobFailureMsg))
		})
		It("RunaiJob non-failed job", func() {
			job := getRunaiJob(jobName, false, "")
			aggr.SetJob(job)

			issues := srm.CheckSubResources(ctx, aggr)
			Expect(issues).To(HaveLen(0))
		})
		It("Distributed failed job", func() {
			job := getPytorchJob(jobName, true, jobFailureMsg)
			aggr.SetJob(job)

			issues := srm.CheckSubResources(ctx, aggr)
			Expect(issues).To(HaveLen(1))
			Expect(issues[0].GetGroup()).To(Equal(issue.Compute))
			Expect(issues[0].Kind()).To(Equal(constants.KindPyTorchJob))
			Expect(issues[0].IsCritical()).To(BeTrue())
			Expect(issues[0].Message()).To(ContainSubstring(jobName))
			Expect(issues[0].Message()).To(ContainSubstring(jobFailureMsg))
		})
		It("Distributed non-failed job", func() {
			job := getPytorchJob(jobName, false, "")
			aggr.SetJob(job)

			issues := srm.CheckSubResources(ctx, aggr)
			Expect(issues).To(HaveLen(0))
		})
	})

	Context("Multiple Groups", func() {
		It("RunaiJob job & route", func() {
			job := getRunaiJob(jobName, true, jobFailureMsg)
			aggr.SetJob(job)

			r := getRouteWithError(routeName, RouteUrlAlreadyExistConditionReason)
			aggr.AddRoutes(r)

			issues := srm.CheckSubResources(ctx, aggr)
			Expect(issues).To(HaveLen(2))
		})
	})
})

func getRouteWithError(name, errorReason string) *openshiftV1.Route {
	return &openshiftV1.Route{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindRoute,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: openshiftV1.RouteStatus{
			Ingress: []openshiftV1.RouteIngress{
				{
					Conditions: []openshiftV1.RouteIngressCondition{
						{
							Type:    openshiftV1.RouteAdmitted,
							Status:  corev1.ConditionFalse,
							Reason:  errorReason,
							Message: "",
						},
					},
				},
			},
		},
	}
}

func getRunaiJob(name string, failed bool, errorMsg string) *runaiJobV1.RunaiJob {
	job := &runaiJobV1.RunaiJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindRunaiJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if failed {
		job.Status.Conditions = []runaiJobV1.JobCondition{{Type: runaiJobV1.JobFailed, Status: "True", Message: errorMsg}}
	}

	return job
}

func getPytorchJob(name string, failed bool, errorMsg string) *kubeflowv1.PyTorchJob {
	job := &kubeflowv1.PyTorchJob{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindPyTorchJob,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if failed {
		job.Status.Conditions = []kubeflowv1.JobCondition{{Type: kubeflowv1.JobFailed, Status: "True", Message: errorMsg}}
	}

	return job
}

// a service needs to have the following to raise error:
// - type load balancer
// - have load balancer in the status, but no ingress with it
// note that these conditions are not sufficient to raise error, error depends on finding an event
func getLoadBalancerServiceWithNoIngress(name string, ns string, uid string) corev1.Service {
	return corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind: constants.KindService,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			UID:       types.UID(uid),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{},
		},
	}
}
