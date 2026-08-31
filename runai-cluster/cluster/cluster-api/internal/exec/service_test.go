/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec_test

import (
	"context"
	"errors"

	execmock "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/exec"
	servicemock "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/autogenerate/mocks/websocket/service"
	clustercontainer "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/container"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/exec"
	"k8s.io/client-go/rest"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/golang/mock/gomock"
	workloadAPI "github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	initContainer       = "init-container1"
	initContainer2      = "init-container2"
	ephemeralContainer  = "ephemeral-container1"
	ephemeralContainer2 = "ephemeral-container2"
	container1          = "container1"
	container2          = "container2"
)

var _ = Describe("Service", func() {
	var (
		ctrl           *gomock.Controller
		mockPodsGetter *servicemock.MockPodsGetter
		mockStreamer   *execmock.MockExecute
		service        *exec.Service
		workload       *workloadAPI.Workload
		options        stream.ExecOptions
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockPodsGetter = servicemock.NewMockPodsGetter(ctrl)
		mockStreamer = execmock.NewMockExecute(ctrl)
		service = exec.NewService(mockPodsGetter, mockStreamer, nil)
		workload = &workloadAPI.Workload{}
		options = stream.ExecOptions{
			Options: stream.Options{
				Container: container1,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("LogsReader", func() {
		Context("when the pod is found and running", func() {
			var pod *v1.Pod

			BeforeEach(func() {
				pod = &v1.Pod{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: container1},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: container1,
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{
										StartedAt: metav1.Now(),
									},
								},
							},
						},
					},
				}
				mockPodsGetter.EXPECT().GetPod(gomock.Any(), workload, options.Pod).Return(pod, nil)
				mockStreamer.EXPECT().CreateRequest(gomock.Any(), gomock.Any()).Return(&rest.Request{}, nil)
			})

			It("should return a reader without errors", func() {
				reader, err := service.GenerateRequest(context.Background(), workload, options)
				Expect(err).NotTo(HaveOccurred())
				Expect(reader).NotTo(BeNil())
			})
		})

		Context("when the pod is not found and request by the user ", func() {
			BeforeEach(func() {
				options.Pod = "pod1"
				mockPodsGetter.EXPECT().GetPod(gomock.Any(), workload, options.Pod).Return(nil, apierrors.NewNotFound(v1.Resource("pod"), "pod1"))
			})

			It("should return a NotFoundError", func() {
				_, err := service.GenerateRequest(context.Background(), workload, options)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apierrors.NewNotFound(v1.Resource("pod"), "pod1")))
			})
		})
		Context("when the pod is not found and not request by the user ", func() {
			BeforeEach(func() {
				mockPodsGetter.EXPECT().GetPod(gomock.Any(), workload, options.Pod).Return(nil, apierrors.NewNotFound(v1.Resource("pod"), ""))
			})

			It("should return a NotReadyError", func() {
				_, err := service.GenerateRequest(context.Background(), workload, options)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(apierrors.NewNotFound(v1.Resource("pod"), "")))
			})
		})

		Context("when the pod is not running", func() {
			var pod *v1.Pod

			BeforeEach(func() {
				pod = &v1.Pod{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: container1},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				}
				mockPodsGetter.EXPECT().GetPod(gomock.Any(), workload, options.Pod).Return(pod, nil)
			})

			It("should return a NotReadyError", func() {
				_, err := service.GenerateRequest(context.Background(), workload, options)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&api.NotReadyError{}))
			})
		})

		Context("when the container is not found", func() {
			var pod *v1.Pod

			BeforeEach(func() {
				pod = &v1.Pod{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: container2},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				}
				mockPodsGetter.EXPECT().GetPod(gomock.Any(), workload, options.Pod).Return(pod, nil)
			})

			It("should return a NotFoundError", func() {
				_, err := service.GenerateRequest(context.Background(), workload, options)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(&api.NotFoundError{}))
			})
		})

		Context("when there is an error getting logs", func() {
			var pod *v1.Pod

			BeforeEach(func() {
				pod = &v1.Pod{
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{Name: container1},
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						ContainerStatuses: []v1.ContainerStatus{
							{
								Name: container1,
								State: v1.ContainerState{
									Running: &v1.ContainerStateRunning{
										StartedAt: metav1.Now(),
									},
								},
							},
						},
					},
				}
				mockPodsGetter.EXPECT().GetPod(gomock.Any(), workload, options.Pod).Return(pod, nil)
				mockStreamer.EXPECT().CreateRequest(gomock.Any(), gomock.Any()).Return(nil, errors.New("streaming error"))
			})

			It("should return an error", func() {
				_, err := service.GenerateRequest(context.Background(), workload, options)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("GetContainer", func() {
		var pod *v1.Pod

		BeforeEach(func() {
			pod = &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: container1},
						{Name: container2},
					},
					InitContainers: []v1.Container{
						{Name: initContainer},
						{Name: initContainer2},
					},
					EphemeralContainers: []v1.EphemeralContainer{
						{EphemeralContainerCommon: v1.EphemeralContainerCommon{Name: ephemeralContainer}},
						{EphemeralContainerCommon: v1.EphemeralContainerCommon{Name: ephemeralContainer2}},
					},
				},
			}
			options = stream.ExecOptions{}
		})

		Context("when container is not specified in options", func() {
			It("should return the first container name if containers exist", func() {
				container, err := clustercontainer.GetContainer(pod, options.Container)
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(container1))
			})

			It("should return the first init container name if no regular containers exist", func() {
				pod.Spec.Containers = nil
				container, err := clustercontainer.GetContainer(pod, options.Container)
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(initContainer))
			})

			It("should return the first ephemeral container name if no regular or init containers exist", func() {
				pod.Spec.Containers = nil
				pod.Spec.InitContainers = nil
				container, err := clustercontainer.GetContainer(pod, options.Container)
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(ephemeralContainer))
			})

			It("should return an error if no containers are found", func() {
				pod.Spec.Containers = nil
				pod.Spec.InitContainers = nil
				pod.Spec.EphemeralContainers = nil
				_, err := clustercontainer.GetContainer(pod, options.Container)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("not found"))
			})
		})

		Context("when container is specified in options", func() {
			It("should return the specified container name", func() {
				options.Container = container2
				container, err := clustercontainer.GetContainer(pod, options.Container)
				Expect(err).NotTo(HaveOccurred())
				Expect(container).To(Equal(container2))
			})
		})
	})
	Context("ValidateContainer", func() {
		var pod *v1.Pod

		BeforeEach(func() {
			pod = &v1.Pod{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{Name: container1},
						{Name: container2},
					},
					InitContainers: []v1.Container{
						{Name: initContainer},
						{Name: initContainer2},
					},
					EphemeralContainers: []v1.EphemeralContainer{
						{EphemeralContainerCommon: v1.EphemeralContainerCommon{Name: ephemeralContainer}},
						{EphemeralContainerCommon: v1.EphemeralContainerCommon{Name: ephemeralContainer2}},
					},
				},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{Name: container1, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}},
						{Name: container2, State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "Pending"}}},
					},
					InitContainerStatuses: []v1.ContainerStatus{
						{Name: initContainer, State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 0}}},
						{Name: initContainer2, State: v1.ContainerState{Terminated: &v1.ContainerStateTerminated{ExitCode: 1}}},
					},
					EphemeralContainerStatuses: []v1.ContainerStatus{
						{Name: ephemeralContainer, State: v1.ContainerState{Running: &v1.ContainerStateRunning{}}},
						{Name: ephemeralContainer2, State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "Creating"}}},
					},
				},
			}
			options = stream.ExecOptions{}
		})

		Context("when container is not specified in options", func() {
			It("should return err", func() {
				err := exec.ValidateContainer(pod, options)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when container is specified in options", func() {
			Context("and the container exists in the pod", func() {
				It("should return nil for a running regular container", func() {
					options.Container = container1
					err := exec.ValidateContainer(pod, options)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error for a waiting regular container", func() {
					options.Container = container2
					err := exec.ValidateContainer(pod, options)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("container is waiting, reason: Pending"))
				})

				It("should return err for a terminated init container with exit code 0", func() {
					options.Container = initContainer
					err := exec.ValidateContainer(pod, options)
					Expect(err).To(HaveOccurred())
				})

				It("should return an error for a terminated init container with non-zero exit code", func() {
					options.Container = initContainer2
					err := exec.ValidateContainer(pod, options)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("container is terminated, exit code: 1"))
				})

				It("should return nil for a running ephemeral container", func() {
					options.Container = ephemeralContainer
					err := exec.ValidateContainer(pod, options)
					Expect(err).NotTo(HaveOccurred())
				})

				It("should return an error for a waiting ephemeral container", func() {
					options.Container = ephemeralContainer2
					err := exec.ValidateContainer(pod, options)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("container is waiting, reason: Creating"))
				})
			})

			Context("and the container does not exist in the pod", func() {
				It("should return an error", func() {
					options.Container = "non-existent-container"
					err := exec.ValidateContainer(pod, options)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("not found"))
				})
			})
		})
	})
})
