/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package managers

import (
	"context"
	"errors"
	"fmt"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	issue2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"
)

const (
	fakeClientErrorMsgFormat = "I was told to fail for %s"
)

type FakeClientWithErrors struct {
	k8sclient.Client
	namePrefixToFail string
}

func NewFakeClientWithErrors(scheme *runtime.Scheme, objects []k8sclient.Object, namePrefixToFail string) *FakeClientWithErrors {
	innerClient := client.InitFakeClientWithObjectNoStatus(scheme, objects...)

	return &FakeClientWithErrors{Client: innerClient, namePrefixToFail: namePrefixToFail}
}

func (c *FakeClientWithErrors) Delete(ctx context.Context, obj k8sclient.Object, opts ...k8sclient.DeleteOption) error {
	if strings.HasPrefix(obj.GetName(), c.namePrefixToFail) {
		return fmt.Errorf(fakeClientErrorMsgFormat, obj.GetName())
	}

	return c.Client.Delete(ctx, obj, opts...)
}

var _ = Describe("Cleanup Manager Tests", func() {
	const (
		Namespace  = "namespace"
		FailPrefix = "aaa-"
		JobName    = "job01"
	)
	var (
		workload                *jobs.JobWorkload
		allResources            []k8sclient.Object
		failedToDeleteResources []common.ObservedResource
		resourceFetchingFn      FetchWorkloadSubResourcesFn
		aggr                    *resource.Aggregator
		manager                 *CleanupManager
		recorder                *record.FakeRecorder
	)
	ctx := context.Background()
	scheme := runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		ingress1 := &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindIngress,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1",
				Namespace: Namespace,
			},
		}
		ingress2 := &networkingv1.Ingress{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindIngress,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      FailPrefix + "ingress2",
				Namespace: Namespace,
			},
		}
		service1 := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindService,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      FailPrefix + "service1",
				Namespace: Namespace,
			},
		}
		service2 := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindService,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      FailPrefix + "service2",
				Namespace: Namespace,
			},
		}
		service3 := &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindService,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service3",
				Namespace: Namespace,
			},
		}
		pvc1 := &corev1.PersistentVolumeClaim{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindPvc,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pvc1",
				Namespace: Namespace,
			},
		}
		job := &runaiJobV1.RunaiJob{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindRunaiJob,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      JobName,
				Namespace: Namespace,
			},
		}
		secret1 := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindSecret,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret1",
				Namespace: Namespace,
			},
		}

		allResources = []k8sclient.Object{ingress1, ingress2, service1, service2, service3, pvc1, job}

		failedToDeleteResources = make([]common.ObservedResource, 0)
		ownedResources := make([]common.ObservedResource, len(allResources))
		for i, obj := range allResources {
			r := common.ObservedResource{
				Resource: common.Resource{
					Name:      obj.GetName(),
					Namespace: obj.GetNamespace(),
				},
			}
			ownedResources[i] = r

			if strings.HasPrefix(obj.GetName(), FailPrefix) {
				failedToDeleteResources = append(failedToDeleteResources, r)
			}
		}

		workload = &jobs.JobWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind: constants.KindTrainingWorkload,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "w",
				Namespace: Namespace,
			},
			Status: jobs.JobStatus{
				WorkloadDebugStatus: common.WorkloadDebugStatus{
					OwnedResources: ownedResources,
				},
			},
		}
		aggr = resource.NewAggregator(workload)
		aggr.AddServices(service1, service2, service3)
		aggr.AddIngresses(ingress1, ingress2)
		aggr.AddPvcs(pvc1)
		aggr.AddSecrets(secret1)
		aggr.SetJob(job)

		resourceFetchingFn = func(ctx context.Context, aggr *resource.Aggregator) ([]k8sclient.Object, error) {
			return nil, nil
		}

		recorder = record.NewFakeRecorder(200)
	})

	Context("ForceFailureCleanup", func() {
		It("successful cleanup", func() {
			c := client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
			manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)
			err := manager.ForceFailureCleanup(ctx, aggr, resourceFetchingFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(workload.Status.OwnedResources)).To(BeZero())
		})

		It("partial failures in cleanup", func() {
			c := NewFakeClientWithErrors(scheme, allResources, FailPrefix)
			manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)
			err := manager.ForceFailureCleanup(ctx, aggr, resourceFetchingFn)

			Expect(err).To(HaveOccurred())
			uw, isJoined := err.(interface{ Unwrap() []error })
			Expect(isJoined).To(BeTrue())
			errs := uw.Unwrap()
			Expect(len(errs)).To(Equal(len(failedToDeleteResources)))

			Expect(len(workload.Status.OwnedResources)).To(Equal(len(failedToDeleteResources)))
			for i, failed := range failedToDeleteResources {
				Expect(workload.Status.OwnedResources[i].Name).To(Equal(failed.Name))
			}
		})

		Context("FailureCleanupPolicy", func() {
			DescribeTable("Parsing",
				func(inputStr string, expectedParsed FailureCleanupPolicy) {
					actualParsed := ParseFailureCleanupPolicy(inputStr)
					Expect(actualParsed).To(Equal(expectedParsed))
				},
				Entry("All", "All", All),
				Entry("None", "None", None),
				Entry("KeepFailing", "KeepFailing", KeepFailing),
				Entry("all", "all", Undefined),
				Entry("blabla", "blabla", Undefined),
				Entry(" ", "", Undefined),
			)

			Context("Cleanup", func() {
				var (
					k8sClient                 k8sclient.Client
					failingIngressNotCritical *networkingv1.Ingress
					failingIngressCritical    *networkingv1.Ingress
				)
				BeforeEach(func() {
					failingIngressNotCritical = &networkingv1.Ingress{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindIngress,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "failing-ingress-not-critical",
							Namespace: Namespace,
						},
					}
					failingIngressCritical = &networkingv1.Ingress{
						TypeMeta: metav1.TypeMeta{
							Kind: constants.KindIngress,
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "failing-ingress-critical",
							Namespace: Namespace,
						},
					}

					for _, ingress := range []*networkingv1.Ingress{failingIngressNotCritical, failingIngressCritical} {
						workload.Status.OwnedResources = append(workload.Status.OwnedResources, common.ObservedResource{
							Resource: common.Resource{
								Name:      ingress.GetName(),
								Namespace: ingress.GetNamespace(),
							},
						})
						allResources = append(allResources, ingress)
						aggr.AddIngresses(ingress)
					}

					aggr.AddIssues(issue2.NewIssue(failingIngressNotCritical, errors.New("this is a not critical error"), false, ""))
					aggr.AddIssues(issue2.NewIssue(failingIngressCritical, errors.New("this is a critical error"), true, ""))

					k8sClient = client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
				})
				It("KeepFailing", func() {
					manager = NewCleanupManager(k8sClient, KeepFailing, recorder)

					err := manager.ForceFailureCleanup(ctx, aggr, resourceFetchingFn)

					Expect(err).NotTo(HaveOccurred())
					Expect(len(workload.Status.OwnedResources)).To(Equal(2))

					ExpectFound(ctx, k8sClient, failingIngressNotCritical)
					ExpectFound(ctx, k8sClient, failingIngressCritical)
				})
				It("All", func() {
					manager = NewCleanupManager(k8sClient, All, recorder)

					err := manager.ForceFailureCleanup(ctx, aggr, resourceFetchingFn)

					Expect(err).NotTo(HaveOccurred())
					Expect(len(workload.Status.OwnedResources)).To(Equal(0))

					ExpectNotFound(ctx, k8sClient, failingIngressNotCritical)
					ExpectNotFound(ctx, k8sClient, failingIngressCritical)
				})
				It("None", func() {
					manager = NewCleanupManager(k8sClient, None, recorder)

					err := manager.ForceFailureCleanup(ctx, aggr, resourceFetchingFn)

					Expect(err).NotTo(HaveOccurred())
					Expect(len(workload.Status.OwnedResources)).To(Equal(len(allResources)))

					ExpectFound(ctx, k8sClient, failingIngressNotCritical)
					ExpectFound(ctx, k8sClient, failingIngressCritical)
				})
			})
		})
	})

	Context("DefaultCleanup", func() {
		It("does not cleanup job", func() {
			c := client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
			manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)
			err := manager.DefaultCleanup(ctx, aggr, resourceFetchingFn)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(workload.Status.OwnedResources)).To(Equal(1))

			foundJob := &runaiJobV1.RunaiJob{}
			Expect(c.Get(ctx, types.NamespacedName{Name: JobName, Namespace: Namespace}, foundJob)).To(Succeed())
			Expect(foundJob).ToNot(BeNil())
		})
	})

	Context("SuspendedCleanup", func() {
		It("unknown workload kind", func() {
			workload.TypeMeta.Kind = ""
			c := client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
			manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)
			err := manager.SuspendedCleanup(ctx, aggr, resourceFetchingFn)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown workload kind"))
		})

		Context("no compute domain", func() {
			It("does nothing (no err, delete no resources", func() {
				initialResourceCount := len(workload.Status.OwnedResources)
				c := client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
				manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)
				err := manager.SuspendedCleanup(ctx, aggr, resourceFetchingFn)

				Expect(err).NotTo(HaveOccurred())
				Expect(len(workload.Status.OwnedResources)).To(Equal(initialResourceCount))
			})
		})

		Context("with compute domain", func() {
			var (
				computeDomain *nvapi.ComputeDomain
			)

			BeforeEach(func() {
				computeDomain = &nvapi.ComputeDomain{
					TypeMeta: metav1.TypeMeta{
						Kind: nvapi.ComputeDomainKind,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cd",
						Namespace: Namespace,
					},
				}

				allResources = append(allResources, computeDomain)
				r := common.ObservedResource{
					Resource: common.Resource{
						Name:      computeDomain.GetName(),
						Namespace: computeDomain.GetNamespace(),
					},
				}
				workload.Status.OwnedResources = append(workload.Status.OwnedResources, r)

				aggr.SetComputeDomain(computeDomain)
			})

			It("already deleted", func() {
				initialResourceCount := len(workload.Status.OwnedResources)
				c := client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
				manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)

				computeDomain.DeletionTimestamp = &metav1.Time{Time: time.Now()}

				err := manager.SuspendedCleanup(ctx, aggr, resourceFetchingFn)

				Expect(err).NotTo(HaveOccurred())
				Expect(len(workload.Status.OwnedResources)).To(Equal(initialResourceCount))
			})

			It("successful deletion", func() {
				initialResourceCount := len(workload.Status.OwnedResources)
				c := client.InitFakeClientWithObjectNoStatus(scheme, allResources...)
				manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)

				err := manager.SuspendedCleanup(ctx, aggr, resourceFetchingFn)

				Expect(err).NotTo(HaveOccurred())
				Expect(len(workload.Status.OwnedResources)).To(Equal(initialResourceCount - 1))
				Eventually(recorder.Events).WithTimeout(3 * time.Second).
					Should(Receive(ContainSubstring(fmt.Sprintf("Successfully deleted %s %s", computeDomain.Kind, computeDomain.Name))))
			})

			It("failed deletion", func() {
				computeDomain.Name = FailPrefix + computeDomain.Name
				initialResourceCount := len(workload.Status.OwnedResources)
				c := NewFakeClientWithErrors(scheme, allResources, FailPrefix)
				manager = NewCleanupManager(c, GetDefaultFailureCleanupPolicy(), recorder)

				err := manager.SuspendedCleanup(ctx, aggr, resourceFetchingFn)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(fakeClientErrorMsgFormat, computeDomain.GetName())))
				Expect(len(workload.Status.OwnedResources)).To(Equal(initialResourceCount))
				Eventually(recorder.Events).WithTimeout(3 * time.Second).
					Should(Receive(ContainSubstring(fmt.Sprintf("Failed to delete %s %s", computeDomain.Kind, computeDomain.Name))))
			})
		})
	})
})

func ExpectNotFound(ctx context.Context, c k8sclient.Client, obj k8sclient.Object) {
	tryGetObject(ctx, c, obj, false)
}

func ExpectFound(ctx context.Context, c k8sclient.Client, obj k8sclient.Object) {
	tryGetObject(ctx, c, obj, true)
}

func tryGetObject(ctx context.Context, c k8sclient.Client, obj k8sclient.Object, shouldFind bool) {
	i := &networkingv1.Ingress{}
	err := c.Get(ctx, k8sclient.ObjectKeyFromObject(obj), i)

	if shouldFind {
		Expect(err).NotTo(HaveOccurred())
		Expect(i.GetName()).To(Equal(obj.GetName()))
	} else {
		Expect(err).To(HaveOccurred())
		Expect(k8serrors.IsNotFound(err)).To(BeTrue())
	}
}
