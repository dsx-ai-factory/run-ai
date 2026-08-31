/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"os"

	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	workloadutils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	openshiftV1 "github.com/openshift/api/route/v1"
	workloadconfig "github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	workloadproject "github.com/run-ai/runai/runai-cluster/cluster/workload-common/project"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultNamespace = "runai-default"
	defaultName      = "test-job"
	ExpectedName     = "tw-test-job-%d"
	projectName      = "default"
	ExpectedJupyter  = "tw-test-job-jupyter"
)

var _ = Describe("ExternalsUrlsBuilder", func() {
	var (
		scheme    *runtime.Scheme
		k8sClient client.Client
	)
	BeforeEach(func() {
		scheme = runtime.NewScheme()
		register.InitSchema(scheme)
		k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
	})

	var (
		externalUrlsBuilder *ExternalsUrlsBuilder
		workload            *jobs.JobWorkload
		workloadBuilder     *workloadutils.JobWorkloadBuilder
	)
	const defaultPort = 8888
	Context("Multiple external urls", func() {
		var (
			ctx context.Context
		)
		expectedResources := []common.Resource{
			{Name: fmt.Sprintf(ExpectedName, 0), Namespace: defaultNamespace, Kind: metav1.GroupVersionKind{Kind: constants.KindService}},
			{Name: fmt.Sprintf(ExpectedName, 1), Namespace: defaultNamespace, Kind: metav1.GroupVersionKind{Kind: constants.KindService}},
			{Name: fmt.Sprintf(ExpectedName, 0), Namespace: defaultNamespace, Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}},
			{Name: fmt.Sprintf(ExpectedName, 1), Namespace: defaultNamespace, Kind: metav1.GroupVersionKind{Kind: constants.KindIngress}},
		}
		const secondedTargetPort = 9999
		BeforeEach(func() {
			ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
			workloadBuilder = workloadutils.NewJobWorkloadBuilder()
			workloadBuilder.WithUid("test-uid")
			workloadBuilder.WithTraining()
			workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
			workloadBuilder.WithExternalUrl([]string{"https://testtos.com", "https://anothertest.com"}, []int{defaultPort, secondedTargetPort})
			labels := map[string]string{}
			workload = workloadBuilder.WithLabels(labels).Build()
			externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, false)

		})
		It("preview function should return service and ingress ", func() {

			externalUrlsResources, err := externalUrlsBuilder.GetPreview(ctx, workload)
			Expect(err).ToNot(HaveOccurred())
			Expect(externalUrlsResources).To(HaveLen(len(expectedResources)))
			for i, r := range externalUrlsResources {
				Expect(r.Name).To(Equal(externalUrlsResources[i].Name))
				Expect(r.Namespace).To(Equal(externalUrlsResources[i].Namespace))
				Expect(r.Kind.Kind).To(Equal(externalUrlsResources[i].Kind.Kind))
			}
		})
		It("Build should return service and ingress with the same namespace as the Workload", func() {
			externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
			Expect(err).ToNot(HaveOccurred())
			Expect(externalUrlsResources).To(HaveLen(len(expectedResources)))
			for i, r := range externalUrlsResources {
				Expect(r.GetName()).To(Equal(expectedResources[i].Name))
				Expect(r.GetNamespace()).To(Equal(expectedResources[i].Namespace))
				Expect(r.GetObjectKind().GroupVersionKind().Kind).To(Equal(expectedResources[i].Kind.Kind))
			}
		})
		Context("Build", func() {
			Context("with auth enabled", func() {
				BeforeEach(func() {
					Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "true"))
				})
				AfterEach(func() {
					Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
				})
				Context("with auth support", func() {
					It("Build should return service with the auth port ", func() {
						secName := fmt.Sprintf(ExpectedName, 1)
						validateServiceBuilderTwoEndpoint(ctx, workload, true, externalUrlsBuilder,
							constants.ExternalUrlAuthContainerDefaultPort,
							constants.ExternalUrlAuthContainerDefaultPort+1, fmt.Sprintf(ExpectedName, 0), secName, constants.KindIngress)
					})
				})
				Context("without auth support", func() {
					It("Build should return service with the external port ", func() {
						secName := fmt.Sprintf(ExpectedName, 1)
						validateServiceBuilderTwoEndpoint(ctx, workload, false, externalUrlsBuilder,
							defaultPort, secondedTargetPort, fmt.Sprintf(ExpectedName, 0), secName, constants.KindIngress)
					})
				})
			})
			Context("with auth disabled", func() {
				BeforeEach(func() {
					Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "false"))
				})
				AfterEach(func() {
					Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
				})
				Context("with auth support", func() {
					It("Build should return service with the external port ", func() {
						secName := fmt.Sprintf(ExpectedName, 1)
						validateServiceBuilderTwoEndpoint(ctx, workload, true, externalUrlsBuilder,
							defaultPort, secondedTargetPort, fmt.Sprintf(ExpectedName, 0), secName, constants.KindIngress)
					})
				})
				Context("without auth support", func() {
					It("Build should return service with the external port ", func() {
						secName := fmt.Sprintf(ExpectedName, 1)
						validateServiceBuilderTwoEndpoint(ctx, workload, false, externalUrlsBuilder,
							defaultPort, secondedTargetPort, fmt.Sprintf(ExpectedName, 0), secName, constants.KindIngress)
					})
				})
			})
		})

	})

	Context("Workload K8s cluster", func() {
		Context("Non Jupyter/TensorFlow Job", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workloadBuilder.WithTraining()
				workloadBuilder.WithExternalUrl([]string{"https://testtos.com"}, []int{defaultPort})
				labels := map[string]string{}
				workload = workloadBuilder.WithLabels(labels).Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, false)

			})
			It("preview function should return service and ingress ", func() {
				externalUrlsResources, _ := externalUrlsBuilder.GetPreview(ctx, workload)
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].Name).To(Equal(fmt.Sprintf(ExpectedName, 0)))
				Expect(externalUrlsResources[0].Kind.Kind).To(Equal(constants.KindService))
				Expect(externalUrlsResources[1].Name).To(Equal(fmt.Sprintf(ExpectedName, 0)))
				Expect(externalUrlsResources[1].Kind.Kind).To(Equal(constants.KindIngress))
			})
			It("preview resources should have the same namespace as the Workload", func() {
				externalUrlsResources, _ := externalUrlsBuilder.GetPreview(ctx, workload)
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].Namespace).To(Equal(defaultNamespace))
				Expect(externalUrlsResources[1].Namespace).To(Equal(defaultNamespace))
			})
			It("preview resources should have same Name, Namespace and Kind as Build", func() {
				externalUrlsPreview, _ := externalUrlsBuilder.GetPreview(ctx, workload)
				acualResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				service, ok := acualResources[0].(*corev1.Service)
				Expect(ok).To(BeTrue())
				Expect(service.Name).To(Equal(externalUrlsPreview[0].Name))
				Expect(service.Namespace).To(Equal(externalUrlsPreview[0].Namespace))
				Expect(service.Kind).To(Equal(externalUrlsPreview[0].Kind.Kind))
				ingress, ok := acualResources[1].(*networkingv1.Ingress)
				Expect(ok).To(BeTrue())
				Expect(ingress.Name).To(Equal(externalUrlsPreview[1].Name))
				Expect(ingress.Namespace).To(Equal(externalUrlsPreview[1].Namespace))
				Expect(ingress.Kind).To(Equal(externalUrlsPreview[1].Kind.Kind))
			})
			Context("Build", func() {
				Context("with auth enabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "true"))
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the auth port ", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, constants.ExternalUrlAuthContainerDefaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindIngress)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, defaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindIngress)
						})
					})
				})
				Context("with auth disabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "false"))
						Expect(workloadconfig.ExternalAuthUrlEnabled(), false)
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, defaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindIngress)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, defaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindIngress)
						})
						It("Build should return service and ingress with the same namespace as the Workload", func() {
							externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(externalUrlsResources).To(HaveLen(2))
							Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
							Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
						})
						It("job url annotation should be added to the Workload", func() {
							_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Equal("https://testtos.com"))

						})
					})
				})
			})

		})
		Context("Old Jupyter Jobs", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
				Expect(err).ToNot(HaveOccurred())
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.SetOldJupyter()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithTraining()
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workload = workloadBuilder.Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, false)
			})
			AfterEach(func() {
				err := os.Unsetenv(workloadconfig.EnvVarClusterDomain)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("Build", func() {
				Context("with auth enabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "true"))
						Expect(workloadconfig.ExternalAuthUrlEnabled(), true)
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the regular port no private endpoints", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "jupyter"), constants.KindIngress)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "jupyter"), constants.KindIngress)
						})
					})
				})
				Context("with auth disabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "false"))
						Expect(workloadconfig.ExternalAuthUrlEnabled(), false)
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "jupyter"), constants.KindIngress)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "jupyter"), constants.KindIngress)
						})
						It("Build should return service and ingress with the same namespace as the Workload", func() {
							externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(externalUrlsResources).To(HaveLen(2))
							Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
							Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
						})
						It("job url annotation should be added to the Workload", func() {
							_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Not(Equal(constants.KindIngress)))

						})
					})
					It("Build should return service and ingress with the same namespace as the Workload", func() {
						externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
						Expect(err).ToNot(HaveOccurred())
						Expect(externalUrlsResources).To(HaveLen(2))
						Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
						Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
					})
					It("job url annotation should be added to the Workload", func() {
						_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
						Expect(err).ToNot(HaveOccurred())
						Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Not(Equal(constants.KindIngress)))

					})
				})
			})
		})
		Context("Old Tensorboard Jobs", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
				Expect(err).ToNot(HaveOccurred())
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.SetOldTensorboard()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithTraining()
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workload = workloadBuilder.Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, false)
			})
			AfterEach(func() {
				err := os.Unsetenv(workloadconfig.EnvVarClusterDomain)
				Expect(err).ToNot(HaveOccurred())
			})
			Context("Build", func() {
				Context("with auth enabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "true"))
						Expect(workloadconfig.ExternalAuthUrlEnabled(), true)
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("should return service with the regular port no private endpoints", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "tensorboard"), constants.KindIngress)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "tensorboard"), constants.KindIngress)
						})
					})
				})
				Context("with auth disabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "false"))
						Expect(workloadconfig.ExternalAuthUrlEnabled(), false)
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "tensorboard"), constants.KindIngress)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, workloadbuilders.DefaultJupyterAndTensorContainerPort, fmt.Sprintf("tw-%s-%s", defaultName, "tensorboard"), constants.KindIngress)
						})
						It("Build should return service and ingress with the same namespace as the Workload", func() {
							externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(externalUrlsResources).To(HaveLen(2))
							Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
							Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
						})
						It("job url annotation should be added to the Workload", func() {
							_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Not(Equal("")))

						})
					})
				})
			})
			It("job url annotation should be added to the Workload", func() {
				_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Not(Equal("")))

			})
		})
	})

	Context("Inference", func() {
		Context("with external", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workloadBuilder.WithExternalUrl([]string{"https://testtos.com"}, []int{defaultPort})
				workloadBuilder.WithInference()
				labels := map[string]string{}
				workload = workloadBuilder.WithLabels(labels).Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, false)

			})

			// till now after adding the external url, the service should have the same port as the external url
			// but we change this after adding the auth proxy to the service, and now the service should have the same number of ports as the number of external urls

			It("for inference workload not tha same port ", func() {
				workload.Kind = constants.KindInferenceWorkload
				resources, _ := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				services := getServices(resources)

				Expect(services).To(HaveLen(1))
				Expect(services[0]).ToNot(BeNil())
				Expect(services[0].Spec.Ports).To(HaveLen(1))

			})

			It("for kind not inference workload not tha same port ", func() {
				workload.Kind = constants.KindTrainingWorkload
				resources, _ := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				services := getServices(resources)

				Expect(services).To(HaveLen(1))
				Expect(services[0]).ToNot(BeNil())
				Expect(services[0].Spec.Ports).To(HaveLen(1))
				Expect(services[0].Spec.Ports[0].TargetPort.IntVal).ToNot(Equal(int32(defaultPort)))

			})
		})
	})

	Context("OpenShift", func() {
		Context("Non Jupyter/TensorFlow Job", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workloadBuilder.WithExternalUrl([]string{"https://testtos.com"}, []int{defaultPort})
				workloadBuilder.WithTraining()
				labels := map[string]string{}
				workload = workloadBuilder.WithLabels(labels).Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, true)

			})
			It("preview function should return service and route ", func() {
				externalUrlsResources, _ := externalUrlsBuilder.GetPreview(ctx, workload)
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].Name).To(Equal(fmt.Sprintf(ExpectedName, 0)))
				Expect(externalUrlsResources[0].Kind.Kind).To(Equal(constants.KindService))
				Expect(externalUrlsResources[1].Name).To(Equal(fmt.Sprintf(ExpectedName, 0)))
				Expect(externalUrlsResources[1].Kind.Kind).To(Equal(constants.KindRoute))
			})
			It("preview resources should have the same namespace as the Workload", func() {
				externalUrlsResources, _ := externalUrlsBuilder.GetPreview(ctx, workload)
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].Namespace).To(Equal(defaultNamespace))
				Expect(externalUrlsResources[1].Namespace).To(Equal(defaultNamespace))
			})
			It("preview resources should have same Name, Namespace and Kind as Build", func() {
				externalUrlsPreview, _ := externalUrlsBuilder.GetPreview(ctx, workload)
				acualResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				service, ok := acualResources[0].(*corev1.Service)
				Expect(ok).To(BeTrue())
				Expect(service.Name).To(Equal(externalUrlsPreview[0].Name))
				Expect(service.Namespace).To(Equal(externalUrlsPreview[0].Namespace))
				Expect(service.Kind).To(Equal(externalUrlsPreview[0].Kind.Kind))
				route, ok := acualResources[1].(*openshiftV1.Route)
				Expect(ok).To(BeTrue())
				Expect(route.Name).To(Equal(externalUrlsPreview[1].Name))
				Expect(route.Namespace).To(Equal(externalUrlsPreview[1].Namespace))
				Expect(route.Kind).To(Equal(externalUrlsPreview[1].Kind.Kind))
			})
			Context("Build", func() {
				Context("with auth enabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "true"))
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the auth port ", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, constants.ExternalUrlAuthContainerDefaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindRoute)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, defaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindRoute)
						})
					})
				})
				Context("with auth disabled", func() {
					BeforeEach(func() {
						Expect(os.Setenv(workloadconfig.EnvVarExternalAuthUrlEnabled, "false"))
					})
					AfterEach(func() {
						Expect(os.Unsetenv(workloadconfig.EnvVarExternalAuthUrlEnabled))
					})
					Context("with auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, true, externalUrlsBuilder, defaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindRoute)
						})
					})
					Context("without auth support", func() {
						It("Build should return service with the external port ", func() {
							validateBuilderOneEndpoint(ctx, workload, false, externalUrlsBuilder, defaultPort, fmt.Sprintf(ExpectedName, 0), constants.KindRoute)
						})
						It("Build should return service and route with the same namespace as the Workload", func() {
							externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(externalUrlsResources).To(HaveLen(2))
							Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
							Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
							previewResources, err := externalUrlsBuilder.GetPreview(ctx, workload)
							Expect(err).ToNot(HaveOccurred())
							Expect(previewResources).To(HaveLen(2))
							Expect(externalUrlsResources[0].GetName()).To(Equal(previewResources[0].Name))
							Expect(externalUrlsResources[1].GetName()).To(Equal(previewResources[1].Name))
						})
						It("job url annotation should be added to the Workload", func() {
							_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
							Expect(err).ToNot(HaveOccurred())
							Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Equal("https://testtos.com"))

						})
					})
				})
				It("Build should return service and route with the same namespace as the Workload", func() {
					externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					Expect(externalUrlsResources).To(HaveLen(2))
					Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
					Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
				})
				It("job url annotation should be added to the Workload", func() {
					_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Equal("https://testtos.com"))

				})
			})
		})
		Context("Old Jupyter Jobs", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
				Expect(err).ToNot(HaveOccurred())
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.SetOldJupyter()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithTraining()
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workload = workloadBuilder.Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, true)
			})
			AfterEach(func() {
				err := os.Unsetenv(workloadconfig.EnvVarClusterDomain)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Build should return service and route ", func() {
				externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].GetName()).To(Equal(fmt.Sprintf("tw-%s-%s", defaultName, "jupyter")))
				Expect(externalUrlsResources[0].GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindService))
				Expect(externalUrlsResources[1].GetName()).To(Equal(fmt.Sprintf("tw-%s-%s", defaultName, "jupyter")))
				Expect(externalUrlsResources[1].GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindRoute))
				previewResources, err := externalUrlsBuilder.GetPreview(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
				Expect(previewResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].GetName()).To(Equal(previewResources[0].Name))
				Expect(externalUrlsResources[1].GetName()).To(Equal(previewResources[1].Name))
			})
			It("Build should return service and route with the same namespace as the Workload", func() {
				externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].GetNamespace()).To(Equal(defaultNamespace))
				Expect(externalUrlsResources[1].GetNamespace()).To(Equal(defaultNamespace))
			})
			It("job url annotation should be added to the Workload", func() {
				_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Not(Equal("")))

			})
		})
		Context("Old Tensorboard Jobs", func() {
			var (
				ctx context.Context
			)
			BeforeEach(func() {
				ctx = workloadproject.SetProjectInContext(context.TODO(), projectName)
				err := os.Setenv(workloadconfig.EnvVarClusterDomain, "test.com")
				Expect(err).ToNot(HaveOccurred())
				workloadBuilder = workloadutils.NewJobWorkloadBuilder()
				workloadBuilder.SetOldTensorboard()
				workloadBuilder.WithUid("test-uid")
				workloadBuilder.WithTraining()
				workloadBuilder.WithName(defaultName).WithNamespace(defaultNamespace)
				workload = workloadBuilder.Build()
				externalUrlsBuilder = NewExternalsUrlsBuilder(k8sClient, true)
			})
			AfterEach(func() {
				err := os.Unsetenv(workloadconfig.EnvVarClusterDomain)
				Expect(err).ToNot(HaveOccurred())
			})
			It("Build should return service and route ", func() {
				externalUrlsResources, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(externalUrlsResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].GetName()).To(Equal(fmt.Sprintf("tw-%s-%s", defaultName, "tensorboard")))
				Expect(externalUrlsResources[0].GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindService))
				Expect(externalUrlsResources[1].GetName()).To(Equal(fmt.Sprintf("tw-%s-%s", defaultName, "tensorboard")))
				Expect(externalUrlsResources[1].GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindRoute))
				previewResources, err := externalUrlsBuilder.GetPreview(ctx, workload)
				Expect(err).ToNot(HaveOccurred())
				Expect(previewResources).To(HaveLen(2))
				Expect(externalUrlsResources[0].GetName()).To(Equal(previewResources[0].Name))
				Expect(externalUrlsResources[1].GetName()).To(Equal(previewResources[1].Name))
			})
			It("job url annotation should be added to the Workload", func() {
				_, err := externalUrlsBuilder.Build(ctx, workload, &ExternalBuilderOptions{})
				Expect(err).ToNot(HaveOccurred())
				Expect(workload.Annotations[constants.AnnotationJobUrl]).To(Not(Equal("")))

			})
		})
	})
})

func validateBuilderOneEndpoint(ctx context.Context, workload *jobs.JobWorkload, authSupport bool, builder ExternalBuilder, expectedPort int32, name string, routeType string) {
	externalUrlsResources, err := builder.Build(ctx, workload, &ExternalBuilderOptions{
		DisableAuth: ptr.To(!authSupport),
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(externalUrlsResources).To(HaveLen(2))
	Expect(externalUrlsResources[0].GetName()).To(Equal(name))
	Expect(externalUrlsResources[0].GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindService))
	service := externalUrlsResources[0].(*corev1.Service)
	Expect(service.Spec.Ports).To(HaveLen(1))
	Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(expectedPort))
	Expect(externalUrlsResources[1].GetName()).To(Equal(name))
	Expect(externalUrlsResources[1].GetObjectKind().GroupVersionKind().Kind).To(Equal(routeType))
	previewResources, err := builder.GetPreview(ctx, workload)
	Expect(err).ToNot(HaveOccurred())
	Expect(previewResources).To(HaveLen(2))
	Expect(externalUrlsResources[0].GetName()).To(Equal(previewResources[0].Name))
	Expect(externalUrlsResources[0].GetNamespace()).To(Equal(previewResources[0].Namespace))
	Expect(externalUrlsResources[1].GetName()).To(Equal(previewResources[1].Name))
	Expect(externalUrlsResources[1].GetNamespace()).To(Equal(previewResources[1].Namespace))

}

func validateServiceBuilderTwoEndpoint(ctx context.Context, workload *jobs.JobWorkload, authSupport bool, builder ExternalBuilder, expectedPortOne, expectedPortTwo int32, name1, name2 string, routeType string) {
	externalUrlsResources, err := builder.Build(ctx, workload, &ExternalBuilderOptions{
		DisableAuth: ptr.To(!authSupport),
	})
	Expect(err).ToNot(HaveOccurred())
	Expect(externalUrlsResources).To(HaveLen(4))
	Expect(externalUrlsResources[0].GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindService))
	service, ok := externalUrlsResources[0].(*corev1.Service)
	Expect(ok).To(BeTrue())
	service2, ok := externalUrlsResources[1].(*corev1.Service)
	Expect(ok).To(BeTrue())
	Expect(service.Spec.Ports).To(HaveLen(1))
	Expect(service.Spec.Ports[0].TargetPort.IntVal).To(Equal(expectedPortOne))
	Expect(service2.Spec.Ports).To(HaveLen(1))
	Expect(service2.Spec.Ports[0].TargetPort.IntVal).To(Equal(expectedPortTwo))
	Expect(externalUrlsResources[2].GetObjectKind().GroupVersionKind().Kind).To(Equal(routeType))
	Expect(externalUrlsResources[2].GetName()).To(Equal(name1))
	Expect(externalUrlsResources[3].GetObjectKind().GroupVersionKind().Kind).To(Equal(routeType))
	Expect(externalUrlsResources[3].GetName()).To(Equal(name2))
	previewResources, err := builder.GetPreview(ctx, workload)
	Expect(err).ToNot(HaveOccurred())
	Expect(previewResources).To(HaveLen(4))
	for i, r := range externalUrlsResources {
		Expect(r.GetName()).To(Equal(previewResources[i].Name))
		Expect(r.GetNamespace()).To(Equal(previewResources[i].Namespace))
	}
}

func getServices(resources []client.Object) []*corev1.Service {
	services := make([]*corev1.Service, 0)
	for _, r := range resources {
		service, ok := r.(*corev1.Service)
		if ok {
			services = append(services, service)
		}
	}
	return services
}
