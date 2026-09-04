/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	client2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	workloadutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Service builder", func() {
	var (
		scheme    *runtime.Scheme
		k8sClient client.Client
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
	})
	Context("Workload with default service type", func() {
		var (
			serviceBuilder  *ServiceBuilder
			workload        *jobs.JobWorkload
			workloadBuilder *workloadutils.JobWorkloadBuilder
			ctx             context.Context
		)
		Context("All the services have the same type", func() {
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.Background(), "default")
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workloadBuilder.WithPorts([]string{"30080:8080", "30060:6060"})
				workloadBuilder.WithDefaultServiceType(constants.ServiceTypeNodePort)
				workloadBuilder.WithTraining()
				workload = workloadBuilder.Build()
				serviceBuilder = NewServiceBuilder(k8sClient)
			})
			It("should build 1 service", func() {
				services, err := serviceBuilder.Build(ctx, workload, ResourceBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(services).To(HaveLen(1))
				service, ok := services[0].(*corev1.Service)
				Expect(ok).To(BeTrue())
				Expect(service.Name).To(Equal(fmt.Sprintf("tw-%s-0-%s", workload.Name, "nodeport")))
			})
			It("should build 1 service with 2 ports", func() {
				services, err := serviceBuilder.Build(ctx, workload, ResourceBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(services).To(HaveLen(1))
				service, ok := services[0].(*corev1.Service)
				Expect(ok).To(BeTrue())
				Expect(service.Spec.Ports).To(HaveLen(2))
			})
			It("build service should match the preview", func() {
				previews, err := serviceBuilder.GetPreview(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
				Expect(previews).To(HaveLen(1))
				services, err := serviceBuilder.Build(ctx, workload, ResourceBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(services).To(HaveLen(1))
				service, ok := services[0].(*corev1.Service)
				Expect(ok).To(BeTrue())
				Expect(service.Name).To(Equal(previews[0].Name))
				Expect(service.Namespace).To(Equal(previews[0].Namespace))
			})

		})

	})
	Context("Build preview", func() {
		var (
			workloadFields *field_groups.WorkloadFields
			serviceBuilder *ServiceBuilder
		)
		BeforeEach(func() {
			serviceBuilder = NewServiceBuilder(k8sClient)
		})
		Context("default service type is nil", func() {
			It("Build preview with nil service type should not fail ", func() {
				workloadFields = &field_groups.WorkloadFields{
					NetworkingFields: field_groups.NetworkingFields{
						ServiceType: &flds.StringField{
							Value: nil,
						},
					},
				}
				workload := &jobs.JobWorkload{
					TypeMeta: metav1.TypeMeta{
						Kind: constants.KindTrainingWorkload,
					},
					Spec: jobs.JobSpec{
						WorkloadFields: *workloadFields,
					},
				}
				_, err := serviceBuilder.GetPreview(context.Background(), workload)
				Expect(err).ToNot(HaveOccurred())
			})
		})
		It("Build preview with nil port should have 0 services ", func() {
			workloadFields = &field_groups.WorkloadFields{
				NetworkingFields: field_groups.NetworkingFields{
					Ports: &flds.PortsField{
						ItemRules: &flds.PortItemRules{},
						Value: &map[string]flds.PortItemAndRules{
							"test": {
								Value: nil,
							},
						},
					},
				},
			}
			workload := &jobs.JobWorkload{
				TypeMeta: metav1.TypeMeta{
					Kind: constants.KindTrainingWorkload,
				},

				Spec: jobs.JobSpec{
					WorkloadFields: *workloadFields,
				},
			}
			preview, err := serviceBuilder.GetPreview(context.Background(), workload)
			Expect(err).ToNot(HaveOccurred())
			Expect(preview).To(HaveLen(0))
		})
		It("Build preview with nil port and two port with same service type should have one service ", func() {
			workloadFields = &field_groups.WorkloadFields{
				NetworkingFields: field_groups.NetworkingFields{
					Ports: &flds.PortsField{
						ItemRules: &flds.PortItemRules{},
						Value: &map[string]flds.PortItemAndRules{
							"test": {
								Value: nil,
							},
							"notEmpty": {
								Value: &flds.PortItemValue{
									ContainerPort: ptr.To(8080),
									HostPort:      ptr.To(8080),
									ServiceType:   ptr.To(constants.ServiceTypeNodePort),
								},
							},
							"notEmpty2": {
								Value: &flds.PortItemValue{
									ContainerPort: ptr.To(9090),
									HostPort:      ptr.To(9090),
									ServiceType:   ptr.To(constants.ServiceTypeNodePort),
								},
							},
						},
					},
				},
			}
			workload := &jobs.JobWorkload{
				TypeMeta: metav1.TypeMeta{
					Kind: constants.KindTrainingWorkload,
				},

				Spec: jobs.JobSpec{
					WorkloadFields: *workloadFields,
				},
			}
			preview, err := serviceBuilder.GetPreview(context.Background(), workload)
			Expect(err).ToNot(HaveOccurred())
			Expect(preview).To(HaveLen(1))

		})
		It("Build preview with nil port and two port with different service type should have two service ", func() {
			workloadFields = &field_groups.WorkloadFields{
				NetworkingFields: field_groups.NetworkingFields{
					Ports: &flds.PortsField{
						ItemRules: &flds.PortItemRules{},
						Value: &map[string]flds.PortItemAndRules{
							"test": {
								Value: nil,
							},
							"notEmpty": {
								Value: &flds.PortItemValue{
									ContainerPort: ptr.To(8080),
									HostPort:      ptr.To(8080),
									ServiceType:   ptr.To(constants.ServiceTypeNodePort),
								},
							},
							"notEmpty2": {
								Value: &flds.PortItemValue{
									ContainerPort: ptr.To(9090),
									HostPort:      ptr.To(9090),
									ServiceType:   ptr.To(constants.ServiceTypeLoadBalancer),
								},
							},
						},
					},
				},
			}
			workload := &jobs.JobWorkload{
				TypeMeta: metav1.TypeMeta{
					Kind: constants.KindTrainingWorkload,
				},

				Spec: jobs.JobSpec{
					WorkloadFields: *workloadFields,
				},
			}
			preview, err := serviceBuilder.GetPreview(context.Background(), workload)
			Expect(err).ToNot(HaveOccurred())
			Expect(preview).To(HaveLen(2))
		})
	})

})
