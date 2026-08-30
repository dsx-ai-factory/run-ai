package pod_utils

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestWatcherUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Watcher Utils Suite")
}

var _ = Describe("PodShouldUpdate", func() {
	var (
		oldPod *corev1.Pod
		newPod *corev1.Pod
	)

	BeforeEach(func() {
		oldPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Annotations: map[string]string{
					constants.PodStatusAnnotation: "Running",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "container1",
						Ready: true,
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{},
						},
					},
				},
				Conditions: []corev1.PodCondition{
					{
						Type:    corev1.PodReady,
						Status:  corev1.ConditionTrue,
						Reason:  "Ready",
						Message: "Pod is ready",
					},
				},
			},
		}
		newPod = oldPod.DeepCopy()
	})

	Context("when pod status annotation changes", func() {
		It("should return true", func() {
			newPod.Annotations[constants.PodStatusAnnotation] = "Pending"
			e := event.UpdateEvent{
				ObjectOld: oldPod,
				ObjectNew: newPod,
			}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})
	})

	Context("when pod phase changes", func() {
		It("should return true", func() {
			newPod.Status.Phase = corev1.PodPending
			e := event.UpdateEvent{
				ObjectOld: oldPod,
				ObjectNew: newPod,
			}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})
	})

	Context("when container status changes", func() {
		It("should return true", func() {
			newPod.Status.ContainerStatuses[0].Ready = false
			e := event.UpdateEvent{
				ObjectOld: oldPod,
				ObjectNew: newPod,
			}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})
	})

	Context("when pod conditions change", func() {
		It("should return true", func() {
			newPod.Status.Conditions[0].Status = corev1.ConditionFalse
			e := event.UpdateEvent{
				ObjectOld: oldPod,
				ObjectNew: newPod,
			}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})
	})

	Context("when nothing changes", func() {
		It("should return false", func() {
			e := event.UpdateEvent{
				ObjectOld: oldPod,
				ObjectNew: newPod,
			}
			Expect(PodShouldUpdate(e)).To(BeFalse())
		})
	})
})

var _ = Describe("IsPodStatusAnnotationChanged", func() {
	var (
		oldPod *corev1.Pod
		newPod *corev1.Pod
	)

	BeforeEach(func() {
		oldPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{
					constants.PodStatusAnnotation: "Running",
				},
			},
		}
		newPod = oldPod.DeepCopy()
	})

	Context("when pod status annotation changes", func() {
		It("should return true", func() {
			newPod.Annotations[constants.PodStatusAnnotation] = "Pending"
			Expect(IsPodStatusAnnotationChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when pod status annotation is the same", func() {
		It("should return false", func() {
			Expect(IsPodStatusAnnotationChanged(oldPod, newPod)).To(BeFalse())
		})
	})

	Context("when old pod has no annotations", func() {
		It("should return true if new pod has annotation", func() {
			oldPod.Annotations = nil
			newPod.Annotations = map[string]string{
				constants.PodStatusAnnotation: "Running",
			}
			Expect(IsPodStatusAnnotationChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when new pod has no annotations", func() {
		It("should return true if old pod had annotation", func() {
			newPod.Annotations = nil
			Expect(IsPodStatusAnnotationChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when both pods have no annotations", func() {
		It("should return false", func() {
			oldPod.Annotations = nil
			newPod.Annotations = nil
			Expect(IsPodStatusAnnotationChanged(oldPod, newPod)).To(BeFalse())
		})
	})
})

var _ = Describe("IsPodPhaseChanged", func() {
	var (
		oldPod *corev1.Pod
		newPod *corev1.Pod
	)

	BeforeEach(func() {
		oldPod = &corev1.Pod{
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		}
		newPod = oldPod.DeepCopy()
	})

	Context("when pod phase changes from Running to Pending", func() {
		It("should return true", func() {
			newPod.Status.Phase = corev1.PodPending
			Expect(IsPodPhaseChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when pod phase changes from Pending to Failed", func() {
		It("should return true", func() {
			oldPod.Status.Phase = corev1.PodPending
			newPod.Status.Phase = corev1.PodFailed
			Expect(IsPodPhaseChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when pod phase is the same", func() {
		It("should return false", func() {
			Expect(IsPodPhaseChanged(oldPod, newPod)).To(BeFalse())
		})
	})
})

var _ = Describe("IsContainerStatusChanged", func() {
	var (
		oldPod *corev1.Pod
		newPod *corev1.Pod
	)

	BeforeEach(func() {
		oldPod = &corev1.Pod{
			Status: corev1.PodStatus{
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "container1",
						Ready: true,
						State: corev1.ContainerState{
							Running: &corev1.ContainerStateRunning{},
						},
					},
				},
			},
		}
		newPod = oldPod.DeepCopy()
	})

	Context("when container ready status changes", func() {
		It("should return true", func() {
			newPod.Status.ContainerStatuses[0].Ready = false
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when container state changes", func() {
		It("should return true when state changes from Running to Waiting", func() {
			newPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Waiting: &corev1.ContainerStateWaiting{
					Reason: "ImagePullBackOff",
				},
			}
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should return true when state changes from Running to Terminated", func() {
			newPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					ExitCode: 1,
				},
			}
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when number of containers changes", func() {
		It("should return true when a new container is added", func() {
			newPod.Status.ContainerStatuses = append(newPod.Status.ContainerStatuses, corev1.ContainerStatus{
				Name:  "container2",
				Ready: false,
			})
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should return true when a container is removed", func() {
			oldPod.Status.ContainerStatuses = append(oldPod.Status.ContainerStatuses, corev1.ContainerStatus{
				Name:  "container2",
				Ready: true,
			})
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when init container status changes", func() {
		BeforeEach(func() {
			oldPod.Status.InitContainerStatuses = []corev1.ContainerStatus{
				{
					Name:  "init-container",
					Ready: true,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			}
			newPod = oldPod.DeepCopy()
		})

		It("should return true when init container ready status changes", func() {
			newPod.Status.InitContainerStatuses[0].Ready = false
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should return true when init container state changes", func() {
			newPod.Status.InitContainerStatuses[0].State = corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					ExitCode: 0,
				},
			}
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when a new container with different name appears", func() {
		It("should return true", func() {
			newPod.Status.ContainerStatuses[0].Name = "different-container"
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when nothing changes", func() {
		It("should return false", func() {
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeFalse())
		})
	})

	Context("when both pods have no containers", func() {
		It("should return false", func() {
			oldPod.Status.ContainerStatuses = []corev1.ContainerStatus{}
			newPod.Status.ContainerStatuses = []corev1.ContainerStatus{}
			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeFalse())
		})
	})
})

var _ = Describe("IsPodConditionsChanged", func() {
	var (
		oldPod *corev1.Pod
		newPod *corev1.Pod
	)

	BeforeEach(func() {
		oldPod = &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:    corev1.PodReady,
						Status:  corev1.ConditionTrue,
						Reason:  "Ready",
						Message: "Pod is ready",
					},
				},
			},
		}
		newPod = oldPod.DeepCopy()
	})

	Context("when condition status changes", func() {
		It("should return true", func() {
			newPod.Status.Conditions[0].Status = corev1.ConditionFalse
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when condition reason changes", func() {
		It("should return true", func() {
			newPod.Status.Conditions[0].Reason = "NotReady"
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when condition message changes", func() {
		It("should return true", func() {
			newPod.Status.Conditions[0].Message = "Pod is not ready"
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when number of conditions changes", func() {
		It("should return true when a new condition is added", func() {
			newPod.Status.Conditions = append(newPod.Status.Conditions, corev1.PodCondition{
				Type:    corev1.PodScheduled,
				Status:  corev1.ConditionTrue,
				Reason:  "Scheduled",
				Message: "Pod has been scheduled",
			})
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should return true when a condition is removed", func() {
			oldPod.Status.Conditions = append(oldPod.Status.Conditions, corev1.PodCondition{
				Type:   corev1.PodScheduled,
				Status: corev1.ConditionTrue,
			})
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when a new condition type appears", func() {
		It("should return true", func() {
			newPod.Status.Conditions[0].Type = corev1.PodScheduled
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when multiple conditions exist and one changes", func() {
		BeforeEach(func() {
			oldPod.Status.Conditions = []corev1.PodCondition{
				{
					Type:    corev1.PodReady,
					Status:  corev1.ConditionTrue,
					Reason:  "Ready",
					Message: "Pod is ready",
				},
				{
					Type:    corev1.PodScheduled,
					Status:  corev1.ConditionTrue,
					Reason:  "Scheduled",
					Message: "Pod is scheduled",
				},
				{
					Type:    corev1.PodInitialized,
					Status:  corev1.ConditionTrue,
					Reason:  "Initialized",
					Message: "All init containers completed",
				},
			}
			newPod = oldPod.DeepCopy()
		})

		It("should return true when one condition changes", func() {
			newPod.Status.Conditions[1].Status = corev1.ConditionFalse
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when nothing changes", func() {
		It("should return false", func() {
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeFalse())
		})
	})

	Context("when both pods have no conditions", func() {
		It("should return false", func() {
			oldPod.Status.Conditions = []corev1.PodCondition{}
			newPod.Status.Conditions = []corev1.PodCondition{}
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeFalse())
		})
	})

	Context("when LastTransitionTime changes but other fields stay the same", func() {
		It("should return false", func() {
			// Note: The function doesn't check LastTransitionTime, only Status, Reason, and Message
			newPod.Status.Conditions[0].LastTransitionTime = metav1.Now()
			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeFalse())
		})
	})
})

var _ = Describe("PodShouldUpdate - Complex scenarios with non-equal pods", func() {
	var (
		basePod *corev1.Pod
	)

	BeforeEach(func() {
		basePod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Annotations: map[string]string{
					constants.PodStatusAnnotation: "Pending",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "app",
						Ready: false,
						State: corev1.ContainerState{
							Waiting: &corev1.ContainerStateWaiting{
								Reason: "ContainerCreating",
							},
						},
					},
				},
				InitContainerStatuses: []corev1.ContainerStatus{
					{
						Name:  "init",
						Ready: true,
						State: corev1.ContainerState{
							Terminated: &corev1.ContainerStateTerminated{
								ExitCode: 0,
							},
						},
					},
				},
				Conditions: []corev1.PodCondition{
					{
						Type:    corev1.PodScheduled,
						Status:  corev1.ConditionTrue,
						Reason:  "Scheduled",
						Message: "Pod scheduled",
					},
					{
						Type:    corev1.PodReady,
						Status:  corev1.ConditionFalse,
						Reason:  "ContainersNotReady",
						Message: "Containers not ready",
					},
				},
			},
		}
	})

	Context("when multiple fields change simultaneously", func() {
		It("should return true when both phase and annotation change", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Annotations[constants.PodStatusAnnotation] = "Running"
			newPod.Status.Phase = corev1.PodRunning

			e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})

		It("should return true when annotation, phase, and container status all change", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Annotations[constants.PodStatusAnnotation] = "Running"
			newPod.Status.Phase = corev1.PodRunning
			newPod.Status.ContainerStatuses[0].Ready = true
			newPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			}

			e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})

		It("should return true when all monitored fields change", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Annotations[constants.PodStatusAnnotation] = "Running"
			newPod.Status.Phase = corev1.PodRunning
			newPod.Status.ContainerStatuses[0].Ready = true
			newPod.Status.Conditions[1].Status = corev1.ConditionTrue

			e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})
	})

	Context("when container counts change", func() {
		It("should detect when containers are added", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Status.ContainerStatuses = append(newPod.Status.ContainerStatuses,
				corev1.ContainerStatus{Name: "sidecar", Ready: false},
				corev1.ContainerStatus{Name: "proxy", Ready: true})

			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should detect when init containers change state", func() {
			oldPod := basePod.DeepCopy()
			oldPod.Status.InitContainerStatuses = append(oldPod.Status.InitContainerStatuses,
				corev1.ContainerStatus{
					Name:  "init-2",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "PodInitializing"},
					},
				})
			newPod := oldPod.DeepCopy()
			newPod.Status.InitContainerStatuses[1].Ready = true
			newPod.Status.InitContainerStatuses[1].State = corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{ExitCode: 0},
			}

			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should detect when regular and init containers both change", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Status.ContainerStatuses[0].Ready = true
			newPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			}

			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("when conditions change", func() {
		It("should detect when one condition out of many changes", func() {
			oldPod := basePod.DeepCopy()
			oldPod.Status.Conditions = append(oldPod.Status.Conditions,
				corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionTrue},
				corev1.PodCondition{Type: corev1.ContainersReady, Status: corev1.ConditionFalse})
			newPod := oldPod.DeepCopy()
			newPod.Status.Conditions[3].Status = corev1.ConditionTrue

			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should detect when multiple conditions change", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Status.Conditions[1].Status = corev1.ConditionTrue
			newPod.Status.Conditions[1].Reason = "Ready"
			newPod.Status.Conditions[1].Message = "All containers ready"

			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})

		It("should detect when conditions are added", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Status.Conditions = append(newPod.Status.Conditions,
				corev1.PodCondition{Type: corev1.ContainersReady, Status: corev1.ConditionTrue})

			Expect(IsPodConditionsChanged(oldPod, newPod)).To(BeTrue())
		})
	})

	Context("real-world pod lifecycle transitions", func() {
		It("should detect pending to running transition", func() {
			oldPod := basePod.DeepCopy()
			newPod := basePod.DeepCopy()
			newPod.Annotations[constants.PodStatusAnnotation] = "Running"
			newPod.Status.Phase = corev1.PodRunning
			newPod.Status.ContainerStatuses[0].Ready = true
			newPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			}
			newPod.Status.Conditions = append(newPod.Status.Conditions,
				corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue})

			e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})

		It("should detect running to failed transition", func() {
			oldPod := basePod.DeepCopy()
			oldPod.Status.Phase = corev1.PodRunning
			oldPod.Status.ContainerStatuses[0].Ready = true
			oldPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			}

			newPod := oldPod.DeepCopy()
			newPod.Annotations[constants.PodStatusAnnotation] = "Failed"
			newPod.Status.Phase = corev1.PodFailed
			newPod.Status.ContainerStatuses[0].Ready = false
			newPod.Status.ContainerStatuses[0].State = corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{ExitCode: 1, Reason: "Error"},
			}

			e := event.UpdateEvent{ObjectOld: oldPod, ObjectNew: newPod}
			Expect(PodShouldUpdate(e)).To(BeTrue())
		})

		It("should detect sidecar container recovery", func() {
			oldPod := basePod.DeepCopy()
			oldPod.Status.ContainerStatuses = append(oldPod.Status.ContainerStatuses,
				corev1.ContainerStatus{
					Name:  "sidecar",
					Ready: false,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					},
				})

			newPod := oldPod.DeepCopy()
			newPod.Status.ContainerStatuses[1].Ready = true
			newPod.Status.ContainerStatuses[1].State = corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			}

			Expect(IsContainerStatusChanged(oldPod, newPod)).To(BeTrue())
		})
	})
})
