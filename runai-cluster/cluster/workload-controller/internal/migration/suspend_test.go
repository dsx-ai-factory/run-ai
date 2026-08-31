/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterCommon "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	v1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	workloadName = "test-workload"
	namespace    = "test-namespace"
	jobName      = "test-job"
	job2Name     = "test-job-2"
	job3Name     = "test-job-3"
	job4Name     = "test-job-4"
)

var _ = Describe("SuspendHandler", func() {
	var (
		handler             *SuspendHandler
		fakeClient          client.Client
		trainWorkload       *twv2.TrainingWorkload
		interactiveWorkload *iwv2.InteractiveWorkload
		scheme              *runtime.Scheme
	)
	scheme = runtime.NewScheme()

	BeforeEach(func() {
		trainWorkload = &twv2.TrainingWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: namespace,
			},
		}
		interactiveWorkload = &iwv2.InteractiveWorkload{
			ObjectMeta: metav1.ObjectMeta{
				Name:      workloadName,
				Namespace: namespace,
			},
		}

		Expect(v1.AddToScheme(scheme)).To(Succeed())
		Expect(twv2.AddToScheme(scheme)).To(Succeed())
		Expect(iwv2.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().WithScheme(scheme).
			WithObjects(trainWorkload, interactiveWorkload).Build()
		handler = NewSuspendHandler(fakeClient)
	})

	Context("when handling a training workload", func() {
		It("should set the workload to be inactive", func() {
			runaiJob := v1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterCommon.LabelWorkloadName: workloadName,
						clusterCommon.LabelWorkloadKind: clusterCommon.KindTrainingWorkload,
					},
					Namespace: namespace,
				},
			}
			err := handler.Handle(context.Background(), runaiJob)
			Expect(err).NotTo(HaveOccurred())
			err = fakeClient.Get(context.Background(),
				client.ObjectKey{Name: workloadName, Namespace: namespace}, trainWorkload)
			Expect(err).NotTo(HaveOccurred())
			Expect(trainWorkload.Spec.Active.Deref()).To(BeFalse())

		})

		It("should not return an error if the workload does not exist", func() {
			runaiJob := v1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterCommon.LabelWorkloadName: "non-existent-workload",
						clusterCommon.LabelWorkloadKind: clusterCommon.KindTrainingWorkload,
					},
					Namespace: namespace,
				},
			}
			err := handler.Handle(context.Background(), runaiJob)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("when the workload is already inactive", func() {
			BeforeEach(func() {
				trainWorkload.Spec.Active = field_types.NewOptionalBoolField(ptr.To(false))
				Expect(fakeClient.Update(context.Background(), trainWorkload)).To(Succeed())
			})
			It("should not return an error", func() {
				runaiJob := v1.RunaiJob{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							clusterCommon.LabelWorkloadName: workloadName,
							clusterCommon.LabelWorkloadKind: clusterCommon.KindTrainingWorkload,
						},
						Namespace: namespace,
					},
				}
				err := handler.Handle(context.Background(), runaiJob)
				Expect(err).NotTo(HaveOccurred())
				actualWorkload := twv2.TrainingWorkload{}
				err = fakeClient.Get(context.Background(),
					client.ObjectKey{Name: workloadName, Namespace: namespace}, &actualWorkload)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualWorkload.Spec.Active.Deref()).To(BeFalse())

			})
		})

	})

	Context("when handling an interactive workload", func() {
		It("should set the workload to be inactive", func() {
			runaiJob := v1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterCommon.LabelWorkloadName: workloadName,
						clusterCommon.LabelWorkloadKind: clusterCommon.KindInteractiveWorkload,
					},
					Namespace: namespace,
				},
			}
			err := handler.Handle(context.Background(), runaiJob)
			Expect(err).NotTo(HaveOccurred())
			actualWorkload := iwv2.InteractiveWorkload{}
			err = fakeClient.Get(context.Background(),
				client.ObjectKey{Name: workloadName, Namespace: namespace}, &actualWorkload)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualWorkload.Spec.Active.Deref()).To(BeFalse())

		})

		It("should not return an error if the workload does not exist", func() {
			runaiJob := v1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterCommon.LabelWorkloadName: "non-existent-workload",
						clusterCommon.LabelWorkloadKind: clusterCommon.KindInteractiveWorkload,
					},
					Namespace: namespace,
				},
			}
			err := handler.Handle(context.Background(), runaiJob)
			Expect(err).NotTo(HaveOccurred())
		})
		Context("when the workload is already inactive", func() {
			BeforeEach(func() {
				interactiveWorkload.Spec.Active = field_types.NewOptionalBoolField(ptr.To(false))
				Expect(fakeClient.Update(context.Background(), interactiveWorkload)).To(Succeed())
			})
			It("should not return an error", func() {
				runaiJob := v1.RunaiJob{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							clusterCommon.LabelWorkloadName: workloadName,
							clusterCommon.LabelWorkloadKind: clusterCommon.KindInteractiveWorkload,
						},
						Namespace: namespace,
					},
				}
				err := handler.Handle(context.Background(), runaiJob)
				Expect(err).NotTo(HaveOccurred())
				actualWorkload := iwv2.InteractiveWorkload{}
				err = fakeClient.Get(context.Background(),
					client.ObjectKey{Name: workloadName, Namespace: namespace}, &actualWorkload)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualWorkload.Spec.Active.Deref()).To(BeFalse())

			})

		})
	})

	Context("when handling an unsupported workload kind", func() {
		It("should not return an error", func() {
			runaiJob := v1.RunaiJob{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						clusterCommon.LabelWorkloadName: workloadName,
						clusterCommon.LabelWorkloadKind: "unsupported-kind",
					},
					Namespace: namespace,
				},
			}
			err := handler.Handle(context.Background(), runaiJob)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("SuspendWorkloads", func() {
	var (
		suspendWorkloads *SuspendWorkloads
		fakeClient       client.Client
		mockExecutor     *mocks.Executor
		scheme           *runtime.Scheme
		job1, job2       *v1.RunaiJob
	)
	scheme = runtime.NewScheme()

	BeforeEach(func() {
		job1 = &v1.RunaiJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					clusterCommon.LabelWorkloadName: workloadName,
					clusterCommon.LabelWorkloadKind: clusterCommon.KindTrainingWorkload,
				},
				Name:      jobName,
				Namespace: namespace,
			},
		}
		job2 = &v1.RunaiJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					clusterCommon.LabelWorkloadName: workloadName,
					clusterCommon.LabelWorkloadKind: clusterCommon.KindInteractiveWorkload,
				},
				Name:      job2Name,
				Namespace: namespace,
			},
			Spec: v1.JobSpec{
				Suspend: ptr.To(true),
			},
		}

		Expect(v1.AddToScheme(scheme)).To(Succeed())
		Expect(twv2.AddToScheme(scheme)).To(Succeed())
		baseClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(job1, job2).Build()
		mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1.GroupVersion})
		mapper.Add(runaijobGVK(), meta.RESTScopeNamespace)
		fakeClient = &clientWithRESTMapper{Client: baseClient, mapper: mapper}
		mockExecutor = &mocks.Executor{}
		suspendWorkloads = NewSuspendWorkloads(fakeClient, MaxWorkers)
		suspendWorkloads.MigrateExecuteCreateFn = func(int, []v1.RunaiJob, func(context.Context, v1.RunaiJob) error) Executor {
			return mockExecutor
		}
	})

	Context("when migrating workloads", func() {
		It("skips before constructing the migrator when the RunaiJob API is not served", func() {
			mapper := &suspendTestRESTMapper{RESTMapper: fakeClient.RESTMapper(), err: &meta.NoKindMatchError{
				GroupKind: runaijobGVK().GroupKind(), SearchedVersions: []string{runaijobGVK().Version},
			}}
			called := false
			suspendWorkloads = NewSuspendWorkloads(&clientWithRESTMapper{Client: fakeClient, mapper: mapper}, MaxWorkers)
			suspendWorkloads.MigrateExecuteCreateFn = func(int, []v1.RunaiJob, func(context.Context, v1.RunaiJob) error) Executor {
				called = true
				return mockExecutor
			}

			Expect(suspendWorkloads.Migrate(context.Background())).To(Succeed())
			Expect(called).To(BeFalse())
		})

		It("returns non-NoMatch RunaiJob mapper errors", func() {
			mapperErr := errors.New("discovery unavailable")
			suspendWorkloads = NewSuspendWorkloads(&clientWithRESTMapper{Client: fakeClient, mapper: &suspendTestRESTMapper{RESTMapper: fakeClient.RESTMapper(), err: mapperErr}}, MaxWorkers)
			Expect(suspendWorkloads.Migrate(context.Background())).To(MatchError(mapperErr))
		})

		It("should return nil if there are no workloads to suspend", func() {
			// turn job2 to not require migration
			job2.Spec.Suspend = ptr.To(false)
			err := fakeClient.Update(context.Background(), job2)
			Expect(err).NotTo(HaveOccurred())

			err = suspendWorkloads.Migrate(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return an error if the workload migration fails", func() {
			mockExecutor.On("Execute", mock.Anything).Return(errors.New("migration error"))
			err := suspendWorkloads.Migrate(context.Background())
			Expect(err).To(HaveOccurred())
		})
		Context("when context is cancelled", func() {
			It("should return an error", func() {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				err := suspendWorkloads.Migrate(ctx)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("when filtering RunaiJobs", func() {
		It("should return only jobs that are not owned by any workload", func() {
			runaiJobs := &v1.RunaiJobList{
				Items: []v1.RunaiJob{
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								clusterCommon.LabelWorkloadName: "owned-workload",
								clusterCommon.LabelWorkloadKind: clusterCommon.KindTrainingWorkload,
							},
							OwnerReferences: []metav1.OwnerReference{
								{
									Kind: clusterCommon.KindTrainingWorkload,
									Name: "owned-workload",
								},
							},
							Name: jobName,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								clusterCommon.LabelWorkloadName: workloadName,
								clusterCommon.LabelWorkloadKind: clusterCommon.KindTrainingWorkload,
							},
							Name: job2Name,
						},
						Spec: v1.JobSpec{
							Suspend: ptr.To(false),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: job3Name,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								clusterCommon.LabelWorkloadName: "owned-workload",
								clusterCommon.LabelWorkloadKind: clusterCommon.KindInteractiveWorkload,
							},
							Name: job4Name,
						},
						Spec: v1.JobSpec{
							Suspend: ptr.To(true),
						},
					},
				},
			}

			filterJobs := suspendWorkloads.filterRunaiJobs(runaiJobs)

			Expect(len(filterJobs)).To(Equal(1))
			Expect(filterJobs[0].Name).To(Equal(job4Name))
		})
	})
})

type clientWithRESTMapper struct {
	client.Client
	mapper meta.RESTMapper
}

func (c *clientWithRESTMapper) RESTMapper() meta.RESTMapper { return c.mapper }

type suspendTestRESTMapper struct {
	meta.RESTMapper
	err error
}

func (m *suspendTestRESTMapper) RESTMapping(schema.GroupKind, ...string) (*meta.RESTMapping, error) {
	return nil, m.err
}

func runaijobGVK() schema.GroupVersionKind {
	return v1.GroupVersion.WithKind("RunaiJob")
}
