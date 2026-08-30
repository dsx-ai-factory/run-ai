package jobs

import (
	"context"
	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	podPhase "github.com/run-ai/runai/runai-cluster/cluster/common/utils/pod"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks/external"
	fakeClient "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	testJobName      = "test-job"
	testUid          = "test-uid"
	testJobNamespace = "test-namespace"
	masterRole       = "master"
	workerRole       = "worker"
	masterPodName    = "master-pod"
	workerPodName    = "worker-pod"
	launcherRole     = "launcher"
	launcherPodName  = "launcher-pod"
)

var _ = Describe("DistributedStatus", func() {
	var (
		k8sClient client.Client
		ctx       context.Context
	)
	BeforeEach(func() {
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		k8sClient = fakeClient.InitFakeClientWithObjectNoStatus(scheme)
		ctx = context.Background()
	})
	Context("IsJobFinished", func() {
		It("should return false when job is not finished", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					CompletionTime: &metav1.Time{},
				},
			}
			Expect(status.IsJobFinished()).To(BeFalse())
		})

		It("should return true when job is finished", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					CompletionTime: &metav1.Time{Time: time.Now()},
				},
			}
			Expect(status.IsJobFinished()).To(BeTrue())
		})
	})

	Context("IsTerminating", func() {
		It("should return false when job is not terminating", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
			}
			Expect(status.IsTerminating()).To(BeFalse())
		})

		It("should return true when job is terminating", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
			}
			status.originalMetaObj.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
			Expect(status.IsTerminating()).To(BeTrue())
		})
	})

	Context("IsJobFailed", func() {
		It("should return false when job is not failed", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					Conditions: []kubeflowcommon.JobCondition{
						{
							Type:   kubeflowcommon.JobFailed,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}
			Expect(status.IsJobFailed()).To(BeFalse())
		})

		It("should return true when job is failed", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					Conditions: []kubeflowcommon.JobCondition{
						{
							Type:   kubeflowcommon.JobFailed,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			Expect(status.IsJobFailed()).To(BeTrue())
		})
	})

	Context("IsJobSucceeded", func() {
		It("should return false when job is not succeeded", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					Conditions: []kubeflowcommon.JobCondition{
						{
							Type:   kubeflowcommon.JobSucceeded,
							Status: corev1.ConditionFalse,
						},
					},
				},
			}
			Expect(status.IsJobSucceeded()).To(BeFalse())
		})

		It("should return true when job is succeeded", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					Conditions: []kubeflowcommon.JobCondition{
						{
							Type:   kubeflowcommon.JobSucceeded,
							Status: corev1.ConditionTrue,
						},
					},
				},
			}
			Expect(status.IsJobSucceeded()).To(BeTrue())
		})
	})
	Context("IsSuspended", func() {
		const msg = "suspend message"

		It("should return false when job is not suspended", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					Conditions: []kubeflowcommon.JobCondition{
						{
							Type:    kubeflowcommon.JobSuspended,
							Status:  corev1.ConditionFalse,
							Message: msg,
						},
					},
				},
			}
			Expect(status.IsSuspended()).To(BeFalse())
			Expect(status.GetSuspendInfo()).To(BeEmpty())
		})

		It("should return true when job is suspended", func() {
			status := &DistributedStatus{
				originalMetaObj: &kubeflowcommon.MPIJob{},
				originalStatusObj: kubeflowcommon.JobStatus{
					Conditions: []kubeflowcommon.JobCondition{
						{
							Type:    kubeflowcommon.JobSuspended,
							Status:  corev1.ConditionTrue,
							Message: msg,
						},
					},
				},
			}
			Expect(status.IsSuspended()).To(BeTrue())
			Expect(status.GetSuspendInfo()).To(Equal(msg))
		})
	})
	Context("Pods base status", func() {
		var (
			status             *DistributedStatus
			runaiPhaseCalcMock external.IRunaiPhaseCalculator
		)
		BeforeEach(func() {
			runaiPhaseCalcMock = external.IRunaiPhaseCalculator{}
			masterLabelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"test": "test", kubeflowcommon.JobRoleLabel: masterRole}}
			workerLabelSelector := metav1.LabelSelector{MatchLabels: map[string]string{"test": "test", kubeflowcommon.JobRoleLabel: workerRole}}
			originalStatusObj := kubeflowcommon.JobStatus{
				ReplicaStatuses: map[kubeflowcommon.ReplicaType]*kubeflowcommon.ReplicaStatus{
					masterRole: {
						Active:        1,
						Succeeded:     0,
						Failed:        0,
						LabelSelector: &masterLabelSelector,
						Selector:      "",
					},
					workerPodName: {
						Active:        3,
						Succeeded:     0,
						Failed:        0,
						LabelSelector: &workerLabelSelector,
					},
				},
			}
			originalMetaObj := &metav1.ObjectMeta{
				Name:      testJobName,
				Namespace: testJobNamespace,

				UID: testUid,
			}
			status = NewDistributedStatusFromJobStatus(originalStatusObj, originalMetaObj, k8sClient, &runaiPhaseCalcMock, getMpiManualPodLabelSelector)
		})
		Context("IsRunning", func() {
			var (
				mockPod *external.IRunaiPodPhaseExtended
			)
			BeforeEach(func() {
				mockPod = &external.IRunaiPodPhaseExtended{}
				runaiPhaseCalcMock.On("Calculate", mock.Anything).Return(mockPod).Once()
				status.pods = []podPhase.IRunaiPodPhaseExtended{mockPod}
			})
			Context("job is not running", func() {
				BeforeEach(func() {
					mockPod.On("GetPhase").Return(podPhase.RunaiPhaseInitializing).Once()
				})
				It("should return false", func() {
					Expect(status.IsRunning()).To(BeFalse())
				})
			})
			Context("job is running", func() {
				BeforeEach(func() {
					mockPod.On("GetPhase").Return(podPhase.RunaiPhaseReady).Once()
				})
				It("should return true", func() {
					Expect(status.IsRunning()).To(BeTrue())
				})
			})
		})
		Context("IsInitializing", func() {
			var (
				mockPod *external.IRunaiPodPhaseExtended
			)
			BeforeEach(func() {
				mockPod = &external.IRunaiPodPhaseExtended{}
				runaiPhaseCalcMock.On("Calculate", mock.Anything).Return(mockPod).Once()
				status.pods = []podPhase.IRunaiPodPhaseExtended{mockPod}
			})
			Context("job is not initializing", func() {
				BeforeEach(func() {
					mockPod.On("GetPhase").Return(podPhase.RunaiPhaseReady).Once()
				})
				It("should return false", func() {
					Expect(status.IsInitializing()).To(BeFalse())
				})
			})
			Context("job is initializing", func() {
				BeforeEach(func() {
					mockPod.On("GetPhase").Return(podPhase.RunaiPhaseInitializing).Twice()
				})
				It("should return true", func() {
					Expect(status.IsInitializing()).To(BeTrue())
				})
			})
		})
		Context("IsPending", func() {
			var (
				mockPod *external.IRunaiPodPhaseExtended
			)
			BeforeEach(func() {
				mockPod = &external.IRunaiPodPhaseExtended{}
				runaiPhaseCalcMock.On("Calculate", mock.Anything).Return(mockPod).Once()
				status.pods = []podPhase.IRunaiPodPhaseExtended{mockPod}
			})
			Context("job is not pending", func() {
				BeforeEach(func() {
					mockPod.On("GetPhase").Return(podPhase.RunaiPhaseReady).Once()
				})
				It("should return false", func() {
					Expect(status.IsPending()).To(BeFalse())
				})
			})
			Context("job is pending", func() {
				BeforeEach(func() {
					mockPod.On("GetPhase").Return(podPhase.RunaiPhasePending).Once()
				})
				It("should return true", func() {
					Expect(status.IsPending()).To(BeTrue())
				})
			})
		})
		Context("getDistributedPods", func() {
			Context("selector status is not nil", func() {
				Context("pod is not owned by job", func() {
					It("should return empty pods", func() {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:            "test-pod",
								Namespace:       testJobNamespace,
								Labels:          map[string]string{"test": "test", kubeflowcommon.JobRoleLabel: "master"},
								OwnerReferences: []metav1.OwnerReference{},
							},
						}
						Expect(k8sClient.Create(ctx, pod)).To(Succeed())
						pods, err := status.getDistributedPods(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(pods).To(HaveLen(0))
					})
				})
				Context("pod is owned by job", func() {
					BeforeEach(func() {
						runaiPhaseCalcMock.On("Calculate", mock.Anything).Return(&podPhase.Extended{}).Once()
					})
					It("should return pods with status", func() {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-pod",
								Namespace: testJobNamespace,
								Labels:    map[string]string{"test": "test", kubeflowcommon.JobRoleLabel: "master"},
								OwnerReferences: []metav1.OwnerReference{
									{
										APIVersion: "v1",
										Kind:       "Job",
										Name:       testJobName,
										UID:        testUid,
									},
								},
							},
							Status: corev1.PodStatus{
								Phase: corev1.PodRunning,
							},
						}

						Expect(k8sClient.Create(ctx, pod)).To(Succeed())
						pods, err := status.getDistributedPods(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(pods).To(HaveLen(1))
					})
				})
			})
			Context("selector is nil use manual labels selector ", func() {
				BeforeEach(func() {
					status.originalStatusObj.ReplicaStatuses["master"].LabelSelector = nil
				})
				Context("pod is not owned by job", func() {
					It("should return empty pods", func() {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:            "test-pod",
								Namespace:       testJobNamespace,
								Labels:          map[string]string{kubeflowcommon.JobNameLabel: testJobName, kubeflowcommon.JobRoleLabel: "master"},
								OwnerReferences: []metav1.OwnerReference{},
							},
						}
						Expect(k8sClient.Create(ctx, pod)).To(Succeed())
						pods, err := status.getDistributedPods(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(pods).To(HaveLen(0))
					})
				})
				Context("pod is owned by job", func() {
					BeforeEach(func() {
						runaiPhaseCalcMock.On("Calculate", mock.Anything).Return(&podPhase.Extended{}).Once()
					})
					It("should return pods with status", func() {
						pod := &corev1.Pod{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-pod",
								Namespace: testJobNamespace,
								Labels:    map[string]string{kubeflowcommon.JobNameLabel: testJobName, kubeflowcommon.JobRoleLabel: "master"},
								OwnerReferences: []metav1.OwnerReference{
									{

										Name: testJobName,
										UID:  testUid,
									},
								},
							},
							Status: corev1.PodStatus{
								Phase: corev1.PodRunning,
							},
						}

						Expect(k8sClient.Create(ctx, pod)).To(Succeed())
						pods, err := status.getDistributedPods(ctx)
						Expect(err).ToNot(HaveOccurred())
						Expect(pods).To(HaveLen(1))
					})
				})
			})
		})
		Context("getPodsLabelSelector", func() {
			Context("selector is not nil", func() {
				BeforeEach(func() {
					status.originalStatusObj.ReplicaStatuses["master"].LabelSelector = nil
				})
				It("should return manual label selector", func() {
					labelSelector, err := status.getLabelSelectorForReplicaPods(masterRole, status.originalStatusObj.ReplicaStatuses[masterRole])
					Expect(err).ToNot(HaveOccurred())
					Expect(labelSelector).ToNot(BeNil())
				})
			})
			Context("selector is  not nil", func() {
				It("should return manual label selector", func() {
					labelSelector, err := status.getLabelSelectorForReplicaPods(masterRole, status.originalStatusObj.ReplicaStatuses[masterRole])
					Expect(err).ToNot(HaveOccurred())
					Expect(labelSelector).ToNot(BeNil())
				})
			})
		})
		Context("getMPiManualPodLabelSelector", func() {
			It("should return master manual label selector", func() {
				labelSelector := getMpiManualPodLabelSelector(testJobName, masterRole)
				Expect(labelSelector).ToNot(BeNil())
				Expect(labelSelector.MatchLabels).To(HaveLen(2))
				Expect(labelSelector.MatchLabels[kubeflowcommon.JobNameLabel]).To(Equal(testJobName))
				Expect(labelSelector.MatchLabels[kubeflowcommon.JobRoleLabel]).To(Equal(masterRole))
			})
			It("should return worker manual label selector", func() {
				labelSelector := getMpiManualPodLabelSelector(testJobName, workerRole)
				Expect(labelSelector).ToNot(BeNil())
				Expect(labelSelector.MatchLabels).To(HaveLen(2))
				Expect(labelSelector.MatchLabels[kubeflowcommon.JobNameLabel]).To(Equal(testJobName))
				Expect(labelSelector.MatchLabels[kubeflowcommon.JobRoleLabel]).To(Equal(workerRole))
			})
		})
		Context("getTrainingJobManualPodLabelSelector", func() {
			It("should return manual label selector", func() {
				labelSelector := getTrainingJobManualPodLabelSelector(testJobName, masterRole)
				Expect(labelSelector).ToNot(BeNil())
				Expect(labelSelector.MatchLabels).To(HaveLen(2))
				Expect(labelSelector.MatchLabels[kubeflowcommon.JobNameLabel]).To(Equal(testJobName))
				Expect(labelSelector.MatchLabels[kubeflowcommon.ReplicaTypeLabel]).To(Equal(masterRole))
			})
		})
	})

	Context("Job is nil", func() {
		var distributedStatus *DistributedStatus

		BeforeEach(func() {
			distributedStatus = &DistributedStatus{originalMetaObj: nil}
		})
		It("Job is not Suspended", func() {
			Expect(distributedStatus.IsSuspended()).To(BeFalse())
		})
		It("Job is not Pending", func() {
			Expect(distributedStatus.IsPending()).To(BeFalse())
		})
		It("Job is not Finished", func() {
			Expect(distributedStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Succeeded", func() {
			Expect(distributedStatus.IsJobSucceeded()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(distributedStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job is not Running", func() {
			Expect(distributedStatus.IsRunning()).To(BeFalse())
		})
		It("Job is not Initializing", func() {
			Expect(distributedStatus.IsInitializing()).To(BeFalse())
		})
		It("Job is Creating", func() {
			Expect(distributedStatus.IsCreating()).To(BeTrue())
		})
		It("Job is not Terminating", func() {
			Expect(distributedStatus.IsTerminating()).To(BeFalse())
		})
	})

	Context("GetPendingPods with ImagePullBackOff pod", func() {
		var (
			status *DistributedStatus
		)

		BeforeEach(func() {
			// Create a pod in ImagePullBackOff state (Phase=Pending, PodScheduled=True)
			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      launcherPodName,
					Namespace: testJobNamespace,
					Labels: map[string]string{
						kubeflowcommon.JobNameLabel: testJobName,
						kubeflowcommon.JobRoleLabel: launcherRole, // MPI jobs use JobRoleLabel, not ReplicaTypeLabel
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "kubeflow.org/v1",
							Kind:       "MPIJob",
							Name:       testJobName,
							UID:        testUid,
							Controller: func() *bool { b := true; return &b }(),
						},
					},
				},
				Spec: corev1.PodSpec{
					NodeName: "test-node", // Pod is scheduled
					Containers: []corev1.Container{
						{
							Name:  "main",
							Image: "invalid-image",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending, // Phase is Pending
					Conditions: []corev1.PodCondition{
						{
							Type:   corev1.PodScheduled,
							Status: corev1.ConditionTrue, // Scheduled
						},
						{
							Type:   corev1.PodReady,
							Status: corev1.ConditionFalse,
							Reason: "ContainersNotReady",
						},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name:  "main",
							Ready: false,
							State: corev1.ContainerState{
								Waiting: &corev1.ContainerStateWaiting{
									Reason:  "ImagePullBackOff",
									Message: "Back-off pulling image \"invalid-image\"",
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, pod)).To(Succeed())

			originalMetaObj := &kubeflowcommon.MPIJob{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testJobName,
					Namespace: testJobNamespace,
					UID:       testUid,
				},
			}

			originalStatusObj := kubeflowcommon.JobStatus{
				ReplicaStatuses: map[kubeflowcommon.ReplicaType]*kubeflowcommon.ReplicaStatus{
					kubeflowcommon.ReplicaType(launcherRole): {
						Active: 1, // Pod exists
					},
				},
			}

			status = NewDistributedStatusFromJobStatus(originalStatusObj, originalMetaObj, k8sClient, &podPhase.RunaiPhaseCalculator{}, getMpiManualPodLabelSelector)
		})

		It("should count 0 pending pods (pod is Initializing, not Pending)", func() {
			// Pod with Phase=Pending + PodScheduled=True should be classified as Initializing
			// Therefore GetPendingPods() should return 0
			pending := status.GetPendingPods()
			initializing := status.GetInitializingPods()
			GinkgoWriter.Printf("DEBUG: PendingPods=%d, InitializingPods=%d\n", pending, initializing)
			_, err := status.getDistributedPods(ctx)
			Expect(err).ToNot(HaveOccurred())

			Expect(pending).To(Equal(int32(0)))
		})

		It("should count 1 initializing pod", func() {
			// The pod should be counted as initializing
			Expect(status.GetInitializingPods()).To(Equal(int32(1)))
		})

		It("should be in Initializing state", func() {
			// With 1 initializing pod and 0 ready pods, job should be initializing
			Expect(status.IsInitializing()).To(BeTrue())
		})
	})
})
