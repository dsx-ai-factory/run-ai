/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"fmt"
	"reflect"

	anyworkload_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/anyworkload-controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	runai_scheduler "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-scheduler"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

var _ = Describe("RunaiConfig", func() {
	Context("Defaulting", func() {
		It("inherits workload group resources when inference endpoint settings are omitted", func() {
			configured := &global.Resources{
				Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("250m"), v1.ResourceMemory: resource.MustParse("64Mi")},
				Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("1"), v1.ResourceMemory: resource.MustParse("128Mi")},
			}
			rc := &RunaiConfig{Spec: RunaiConfigSpec{Global: &global.Global{WorkloadServices: &global.WorkloadServices{Resources: configured}}}}
			rc.SetDefaultsWhereNeeded()
			Expect(rc.Spec.InferenceEndpointController.Resources.Requests).To(Equal(configured.Requests))
			Expect(rc.Spec.InferenceEndpointController.Resources.Limits).To(Equal(configured.Limits))
		})

		It("should set default values", func() {
			resource := &RunaiConfig{}
			resource.SetDefaultsWhereNeeded()
			spec := &resource.Spec
			deprecatedCompatibilityFields := map[string]struct{}{
				"RunAIContainerToolkit":            {},
				"RunAIContainerToolkitReservation": {},
				"DevicePlugin":                     {},
				"Engine":                           {},
				"NodeExporter":                     {},
			}

			fieldCount := reflect.TypeOf(spec).Elem().NumField()

			for i := 0; i < fieldCount; i++ {
				field := reflect.ValueOf(spec).Elem().FieldByIndex([]int{i})
				fieldName := reflect.TypeOf(spec).Elem().FieldByIndex([]int{i}).Name
				if _, deprecated := deprecatedCompatibilityFields[fieldName]; deprecated {
					By(fmt.Sprintf("verifying deprecated compatibility field [%s] remains unset", fieldName))
					Expect(field.IsNil()).To(BeTrue())
					continue
				}
				fieldValue := field.Interface()
				zeroValueForField := reflect.Zero(reflect.TypeOf(fieldValue).Elem()).Interface()

				By(fmt.Sprintf("comparing value of field [%s] to its zero value [%v]",
					fieldName, zeroValueForField))
				Expect(reflect.DeepEqual(fieldValue, zeroValueForField)).To(BeFalse())
			}
		})

		It("should leave workload-injected container configuration unset by default", func() {
			resource := &RunaiConfig{}
			resource.SetDefaultsWhereNeeded()

			Expect(resource.Spec.Global.WorkloadInjectedContainers).To(BeNil())
		})

		It("should preserve independent workload-injected container configuration", func() {
			resource := &RunaiConfig{
				Spec: RunaiConfigSpec{
					Global: &global.Global{
						WorkloadInjectedContainers: &global.WorkloadInjectedContainersConfig{
							OAuth2Proxy:    &global.OAuth2ProxyContainerConfig{},
							UsernameLoader: &global.UsernameLoaderContainerConfig{},
							GitSync:        &global.GitSyncContainerConfig{},
							S3FileSystem:   &global.S3FileSystemContainerConfig{},
						},
					},
				},
			}

			resource.SetDefaultsWhereNeeded()

			Expect(resource.Spec.Global.WorkloadInjectedContainers.OAuth2Proxy).NotTo(BeNil())
			Expect(resource.Spec.Global.WorkloadInjectedContainers.UsernameLoader).NotTo(BeNil())
			Expect(resource.Spec.Global.WorkloadInjectedContainers.GitSync).NotTo(BeNil())
			Expect(resource.Spec.Global.WorkloadInjectedContainers.S3FileSystem).NotTo(BeNil())
		})

		It("should use chart-provided images for anyworkload-controller", func() {
			GinkgoT().Setenv("IMG_DEFAULT_REGISTRY", "internal.example.com")
			GinkgoT().Setenv("IMG_DEFAULT_VERSION", "runai-version")
			GinkgoT().Setenv("IMG_THIRD_PARTY_GIT_SYNC_REGISTRY", "registry.k8s.io")
			GinkgoT().Setenv("IMG_THIRD_PARTY_GIT_SYNC_REPOSITORY", "git-sync/git-sync")
			GinkgoT().Setenv("IMG_THIRD_PARTY_GIT_SYNC_TAG", "v4.7.0")
			GinkgoT().Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_REGISTRY", "quay.io")
			GinkgoT().Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_REPOSITORY", "oauth2-proxy/oauth2-proxy")
			GinkgoT().Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_TAG", "v7.15.4")
			GinkgoT().Setenv("IMG_THIRD_PARTY_WAIT_UBI_MINIMAL_REGISTRY", "registry.access.redhat.com")
			GinkgoT().Setenv("IMG_THIRD_PARTY_WAIT_UBI_MINIMAL_REPOSITORY", "ubi9/ubi-minimal")
			GinkgoT().Setenv("IMG_THIRD_PARTY_WAIT_UBI_MINIMAL_TAG", "9.8")
			resource := &RunaiConfig{}
			resource.SetDefaultsWhereNeeded()

			gitSync := resource.Spec.AnyworkloadController.GitSyncImage
			Expect(*gitSync.Registry).To(Equal("registry.k8s.io"))
			Expect(*gitSync.Name).To(Equal("git-sync/git-sync"))
			Expect(*gitSync.Tag).To(Equal("v4.7.0"))

			oauth2Proxy := resource.Spec.AnyworkloadController.RunaiOauth2ProxyImage
			Expect(*oauth2Proxy.Registry).To(Equal("quay.io"))
			Expect(*oauth2Proxy.Name).To(Equal("oauth2-proxy/oauth2-proxy"))
			Expect(*oauth2Proxy.Tag).To(Equal("v7.15.4"))

			usernameLoader := resource.Spec.AnyworkloadController.RunaiOauth2ProxyUsernameLoaderImage
			Expect(*usernameLoader.Registry).To(Equal("registry.access.redhat.com"))
			Expect(*usernameLoader.Name).To(Equal("ubi9/ubi-minimal"))
			Expect(*usernameLoader.Tag).To(Equal("9.8"))
		})

		It("should preserve explicit oauth2-proxy image overrides", func() {
			GinkgoT().Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_REGISTRY", "quay.io")
			GinkgoT().Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_REPOSITORY", "oauth2-proxy/oauth2-proxy")
			GinkgoT().Setenv("IMG_THIRD_PARTY_OAUTH2_PROXY_TAG", "v7.15.4")
			resource := &RunaiConfig{
				Spec: RunaiConfigSpec{
					AnyworkloadController: &anyworkload_controller.AnyworkloadController{
						RunaiOauth2ProxyImage: &common.Image{
							Registry: ptr.To("registry.example.com"),
							Name:     ptr.To("custom/oauth2-proxy"),
							Tag:      ptr.To("custom-tag"),
						},
					},
				},
			}
			resource.SetDefaultsWhereNeeded()

			image := resource.Spec.AnyworkloadController.RunaiOauth2ProxyImage
			Expect(*image.Registry).To(Equal("registry.example.com"))
			Expect(*image.Name).To(Equal("custom/oauth2-proxy"))
			Expect(*image.Tag).To(Equal("custom-tag"))
		})

		It("should preserve explicit username-loader image overrides", func() {
			customImage := func() *common.Image {
				return &common.Image{
					Registry: ptr.To("registry.example.com"),
					Name:     ptr.To("custom/username-loader"),
					Tag:      ptr.To("1.2.3"),
				}
			}
			resource := &RunaiConfig{
				Spec: RunaiConfigSpec{
					AnyworkloadController: &anyworkload_controller.AnyworkloadController{
						RunaiOauth2ProxyUsernameLoaderImage: customImage(),
					},
				},
			}
			resource.SetDefaultsWhereNeeded()

			images := []*common.Image{resource.Spec.AnyworkloadController.RunaiOauth2ProxyUsernameLoaderImage}
			for _, image := range images {
				Expect(image.Registry).NotTo(BeNil())
				Expect(*image.Registry).To(Equal("registry.example.com"))
				Expect(image.Name).NotTo(BeNil())
				Expect(*image.Name).To(Equal("custom/username-loader"))
				Expect(image.Tag).NotTo(BeNil())
				Expect(*image.Tag).To(Equal("1.2.3"))
			}
		})

		When("common resources are set", func() {
			var rc *RunaiConfig

			const (
				schedulingServicesCpu    = "1"
				schedulingServicesMemory = "1Gi"
				syncServicesCpu          = "2"
				syncServicesMemory       = "2Gi"
				workloadServicesCpu      = "3"
				workloadServicesMemory   = "3Gi"
			)

			/*
				Scheduling services:
				- scheduler
				- status-updater
				- metrics-exporter
				- pod-grouper
				- pod-group-assigner
				- pod-group-controller
				- numa-placement-exporter
				- queue-controller
				- nodepool-controller
				- binder
				- device-plugin

				Sync services:
				- cluster-sync
				- agent

				Workload services:
				- workload-controller
				- job-controller
				- runtime-rules
				- external-workload-integrator
				- cluster-redis
				- cluster-api
				- inference-workload-controller
				- researcher-service
				- shared-objects-controller
				- workload-exporter
			*/
			When("common resources are set for all service-sets", func() {
				When("both requests and limits are set", func() {
					BeforeEach(func() {
						rc = &RunaiConfig{
							Spec: RunaiConfigSpec{
								Global: &global.Global{
									SchedulingServices: &global.SchedulingServices{
										Resources: &global.Resources{
											Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(schedulingServicesCpu), v1.ResourceMemory: resource.MustParse(schedulingServicesMemory)},
											Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(schedulingServicesCpu), v1.ResourceMemory: resource.MustParse(schedulingServicesMemory)},
										},
									},
									SyncServices: &global.SyncServices{
										Resources: &global.Resources{
											Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(syncServicesCpu), v1.ResourceMemory: resource.MustParse(syncServicesMemory)},
											Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(syncServicesCpu), v1.ResourceMemory: resource.MustParse(syncServicesMemory)},
										},
									},
									WorkloadServices: &global.WorkloadServices{
										Resources: &global.Resources{
											Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)},
											Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)},
										},
									},
								},
							},
						}
					})
					It("should set common requests and limits correctly", func() {
						rc.SetDefaultsWhereNeeded()

						By("for scheduling services")
						schedulingServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(schedulingServicesCpu), v1.ResourceMemory: resource.MustParse(schedulingServicesMemory)}

						Expect(rc.Spec.RunAIScheduler.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.RunAIScheduler.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.StatusUpdater.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.StatusUpdater.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.MetricsExporter.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.MetricsExporter.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.PodGrouper.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.PodGrouper.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.PodGroupAssigner.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.PodGroupAssigner.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.PodGroupController.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.PodGroupController.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.QueueController.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.QueueController.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.NodePoolController.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.NodePoolController.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						Expect(rc.Spec.Binder.Resources.Requests).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.Binder.Resources.Limits).To(Equal(schedulingServicesExpectedResources))

						By("for sync services")
						syncServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(syncServicesCpu), v1.ResourceMemory: resource.MustParse(syncServicesMemory)}

						Expect(rc.Spec.ClusterSync.Resources.Requests).To(Equal(syncServicesExpectedResources))
						Expect(rc.Spec.ClusterSync.Resources.Limits).To(Equal(syncServicesExpectedResources))

						Expect(rc.Spec.Agent.Resources.Requests).To(Equal(syncServicesExpectedResources))
						Expect(rc.Spec.Agent.Resources.Limits).To(Equal(syncServicesExpectedResources))

						By("for workload services")
						workloadServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)}

						Expect(rc.Spec.AnyworkloadController.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.AnyworkloadController.Resources.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.RuntimeRules.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.RuntimeRules.Resources.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.ExternalWorkloadIntegrator.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ExternalWorkloadIntegrator.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.InferenceEndpointController.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.InferenceEndpointController.Resources.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.ClusterRedis.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterRedis.Resources.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.ClusterApi.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterApi.Resources.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.ResearcherService.Resources.ResearcherService.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ResearcherService.Resources.ResearcherService.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.SharedObjectsController.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.SharedObjectsController.Resources.Limits).To(Equal(workloadServicesExpectedResources))

						Expect(rc.Spec.WorkloadExporter.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.WorkloadExporter.Resources.Limits).To(Equal(workloadServicesExpectedResources))
					})

				})

				When("only limits are set", func() {
					BeforeEach(func() {
						rc = &RunaiConfig{
							Spec: RunaiConfigSpec{
								Global: &global.Global{
									SchedulingServices: &global.SchedulingServices{
										Resources: &global.Resources{
											Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse(schedulingServicesCpu), v1.ResourceMemory: resource.MustParse(schedulingServicesMemory)},
										},
									},
									SyncServices: &global.SyncServices{
										Resources: &global.Resources{
											Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse(syncServicesCpu), v1.ResourceMemory: resource.MustParse(syncServicesMemory)},
										},
									},
									WorkloadServices: &global.WorkloadServices{
										Resources: &global.Resources{
											Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)},
										},
									},
								},
							},
						}
					})

					It("should set common limits correctly", func() {
						rc.SetDefaultsWhereNeeded()

						By("for scheduling services")
						schedulingServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(schedulingServicesCpu), v1.ResourceMemory: resource.MustParse(schedulingServicesMemory)}

						Expect(rc.Spec.RunAIScheduler.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.StatusUpdater.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.MetricsExporter.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.PodGrouper.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.PodGroupAssigner.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.PodGroupController.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.QueueController.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.NodePoolController.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						Expect(rc.Spec.Binder.Resources.Limits).To(Equal(schedulingServicesExpectedResources))
						By("for sync services")
						syncServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(syncServicesCpu), v1.ResourceMemory: resource.MustParse(syncServicesMemory)}
						Expect(rc.Spec.ClusterSync.Resources.Limits).To(Equal(syncServicesExpectedResources))
						Expect(rc.Spec.Agent.Resources.Limits).To(Equal(syncServicesExpectedResources))

						By("for workload services")
						workloadServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)}
						Expect(rc.Spec.AnyworkloadController.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.RuntimeRules.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ExternalWorkloadIntegrator.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterRedis.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterApi.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ResearcherService.Resources.ResearcherService.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.SharedObjectsController.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.WorkloadExporter.Resources.Limits).To(Equal(workloadServicesExpectedResources))
					})
				})
			})

			When("common resources are set for a specific service-set", func() {
				When("only workload services resources are set", func() {
					BeforeEach(func() {
						rc = &RunaiConfig{
							Spec: RunaiConfigSpec{
								Global: &global.Global{
									WorkloadServices: &global.WorkloadServices{
										Resources: &global.Resources{
											Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)},
											Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)},
										},
									},
								},
							},
						}
					})

					It("should set workload services resources correctly", func() {
						rc.SetDefaultsWhereNeeded()

						workloadServicesExpectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(workloadServicesCpu), v1.ResourceMemory: resource.MustParse(workloadServicesMemory)}
						Expect(rc.Spec.AnyworkloadController.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.AnyworkloadController.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.RuntimeRules.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.RuntimeRules.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ExternalWorkloadIntegrator.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ExternalWorkloadIntegrator.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterRedis.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterRedis.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterApi.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ClusterApi.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ResearcherService.Resources.ResearcherService.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.ResearcherService.Resources.ResearcherService.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.SharedObjectsController.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.SharedObjectsController.Resources.Limits).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.WorkloadExporter.Resources.Requests).To(Equal(workloadServicesExpectedResources))
						Expect(rc.Spec.WorkloadExporter.Resources.Limits).To(Equal(workloadServicesExpectedResources))
					})
				})
			})

		})

		When("group resources are set", func() {
			var rc *RunaiConfig

			const (
				groupCpu    = "2"
				groupMemory = "2Gi"
			)

			BeforeEach(func() {
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							WorkloadServices: &global.WorkloadServices{
								Resources: &global.Resources{
									Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(groupCpu), v1.ResourceMemory: resource.MustParse(groupMemory)},
									Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(groupCpu), v1.ResourceMemory: resource.MustParse(groupMemory)},
								},
							},
							SchedulingServices: &global.SchedulingServices{
								Resources: &global.Resources{
									Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(groupCpu), v1.ResourceMemory: resource.MustParse(groupMemory)},
									Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(groupCpu), v1.ResourceMemory: resource.MustParse(groupMemory)},
								},
							},
						},
					},
				}
			})

			It("should override individually set resources with group resources", func() {
				rc.Spec.AnyworkloadController = &anyworkload_controller.AnyworkloadController{
					Resources: &anyworkload_controller.Resources{
						Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("100m"), v1.ResourceMemory: resource.MustParse("128Mi")},
						Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("100m"), v1.ResourceMemory: resource.MustParse("128Mi")},
					},
				}
				rc.Spec.RunAIScheduler = &runai_scheduler.RunAIScheduler{
					Resources: &runai_scheduler.Resources{
						Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("100m"), v1.ResourceMemory: resource.MustParse("128Mi")},
						Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse("100m"), v1.ResourceMemory: resource.MustParse("128Mi")},
					},
				}

				rc.SetDefaultsWhereNeeded()

				groupExpected := v1.ResourceList{v1.ResourceCPU: resource.MustParse(groupCpu), v1.ResourceMemory: resource.MustParse(groupMemory)}
				Expect(rc.Spec.AnyworkloadController.Resources.Requests).To(Equal(groupExpected))
				Expect(rc.Spec.AnyworkloadController.Resources.Limits).To(Equal(groupExpected))
				Expect(rc.Spec.RunAIScheduler.Resources.Requests).To(Equal(groupExpected))
				Expect(rc.Spec.RunAIScheduler.Resources.Limits).To(Equal(groupExpected))
			})
		})

		When("individual service resources are set without group resources", func() {
			var rc *RunaiConfig

			const (
				individualCpu    = "500m"
				individualMemory = "512Mi"
			)

			BeforeEach(func() {
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						AnyworkloadController: &anyworkload_controller.AnyworkloadController{
							Resources: &anyworkload_controller.Resources{
								Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(individualCpu), v1.ResourceMemory: resource.MustParse(individualMemory)},
								Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(individualCpu), v1.ResourceMemory: resource.MustParse(individualMemory)},
							},
						},
						RunAIScheduler: &runai_scheduler.RunAIScheduler{
							Resources: &runai_scheduler.Resources{
								Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse(individualCpu), v1.ResourceMemory: resource.MustParse(individualMemory)},
								Limits:   v1.ResourceList{v1.ResourceCPU: resource.MustParse(individualCpu), v1.ResourceMemory: resource.MustParse(individualMemory)},
							},
						},
					},
				}
			})

			It("should preserve individual service resources", func() {
				rc.SetDefaultsWhereNeeded()

				expectedResources := v1.ResourceList{v1.ResourceCPU: resource.MustParse(individualCpu), v1.ResourceMemory: resource.MustParse(individualMemory)}
				Expect(rc.Spec.AnyworkloadController.Resources.Requests).To(Equal(expectedResources))
				Expect(rc.Spec.AnyworkloadController.Resources.Limits).To(Equal(expectedResources))
				Expect(rc.Spec.RunAIScheduler.Resources.Requests).To(Equal(expectedResources))
				Expect(rc.Spec.RunAIScheduler.Resources.Limits).To(Equal(expectedResources))
			})
		})

		Context("ServiceGroupMapping", func() {
			// ServiceGroupMapping is consumed by reflection (specValue.FieldByName(serviceName)),
			// so every key MUST be an exact RunaiConfigSpec field name. A stale key (e.g. after a
			// field rename) silently no-ops with no compile error, so assert resolution here.
			It("should map every key to a real RunaiConfigSpec field", func() {
				specType := reflect.TypeOf(RunaiConfigSpec{})
				for serviceName := range ServiceGroupMapping {
					_, ok := specType.FieldByName(serviceName)
					Expect(ok).To(BeTrue(), fmt.Sprintf("ServiceGroupMapping key %q has no matching RunaiConfigSpec field", serviceName))
				}
			})
		})
	})
})
