/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/anyworkload-controller/plugins/networking"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/controller"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/test"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

func makeEndpointService(name string, port int32, selectorApp string) endpoints.Service {
	return endpoints.Service{
		Create: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{"app": selectorApp},
				Ports:    []corev1.ServicePort{{Port: port, TargetPort: intstr.FromInt32(port)}},
			},
		},
	}
}

func makeMetricsEndpointService(name, selectorApp, engineLabel string) endpoints.Service {
	labels := map[string]string{
		clusterconstants.LabelWorkloadMetricsService: "true",
		clusterconstants.LabelWorkloadKind:           clusterconstants.KindInferenceWorkload,
		engineLabel:                                  "true",
	}
	if engineLabel == clusterconstants.LabelNIMWorkloadMetricsService {
		labels[clusterconstants.LabelNIMWorkload] = "true"
	} else {
		labels[clusterconstants.LabelVLLMWorkload] = "true"
	}
	return endpoints.Service{
		Create: &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
			Spec: corev1.ServiceSpec{
				Type:     corev1.ServiceTypeClusterIP,
				Selector: map[string]string{clusterconstants.LabelWorkloadName: selectorApp},
				Ports: []corev1.ServicePort{{
					Name:       "metrics",
					Port:       8080,
					TargetPort: intstr.FromInt32(8080),
					Protocol:   corev1.ProtocolTCP,
				}},
			},
		},
	}
}

func makeEndpointIngress(ingressName, host, svcName string, svcPort int32) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: ingressName},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: svcName,
									Port: networkingv1.ServiceBackendPort{Number: svcPort},
								},
							},
						}},
					},
				},
			}},
		},
	}
}

func createEndpointsRunner(
	ctx context.Context,
	k8sClient client.Client,
	name, namespace string,
	clientObject client.Object,
	cfg *endpoints.Config,
) *dromosv1alpha1.WorkloadRunner {
	runner := &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Annotations: map[string]string{
				constants.ControlPlaneWorkloadIdAnnotationForWorkload: uuid.New().String(),
			},
		},
		Spec: dromosv1alpha1.WorkloadRunnerSpec{
			Template: marshalRawExtension(clientObject),
		},
	}
	Expect(endpoints.Set(runner, cfg)).To(Succeed())
	return createResource(ctx, k8sClient, runner)
}

var _ = Describe("Networking Plugin Endpoints (annotation-driven) Integration Tests", Serial, Ordered, func() {
	var (
		k8sManager     ctrl.Manager
		k8sClient      client.Client
		managerCtx     context.Context
		managerCancel  context.CancelFunc
		endpointsNs    = "runai-endpoints-test"
		controllerName = "networking-endpoints-plugin-test-controller"
	)

	BeforeAll(func() {
		managerCtx, managerCancel = context.WithCancel(ctx)

		_ = os.Setenv(workloadbuilders.ExternalUrlAuthContainerImageEnv, "quay.io/oauth2-proxy/oauth2-proxy:v7.6.0")
		_ = os.Setenv(workloadconfig.EnvVarOauth2ProxyOidcIssuerUrl, "https://auth.example.com")
		_ = os.Setenv(workloadconfig.EnvVarControlPlaneEndpoint, "https://cp.example.com")
		_ = os.Setenv(workloadbuilders.WorkloadAuthSecretNameEnv, "oauth2-proxy-secret")
		_ = os.Setenv(workloadconfig.EnvVarOauth2ProxyUsernameLoaderImage, "busybox:latest")

		absCertDir, err := filepath.Abs(webhookCertDir)
		Expect(err).NotTo(HaveOccurred())

		var managerErr error
		k8sManager, managerErr = ctrl.NewManager(cfg, ctrl.Options{
			Scheme:  scheme.Scheme,
			Metrics: metricsserver.Options{BindAddress: "0"},
			WebhookServer: webhook.NewServer(webhook.Options{
				CertDir: absCertDir,
				Port:    webhookPort,
			}),
		})
		Expect(managerErr).NotTo(HaveOccurred())

		k8sClient = k8sManager.GetClient()

		networkingPlugin := networking.NewPlugin(k8sClient)
		registry := framework.NewRegistry(networkingPlugin)
		fw := framework.NewFramework(registry, nil)
		ctrlInst := controller.NewWorkloadRunnerController(k8sClient, fw, k8sManager.GetEventRecorderFor(controllerName)).WithName(controllerName)
		Expect(ctrlInst.SetupWithManager(managerCtx, k8sManager)).To(Succeed())

		setupWebhooks(fw, k8sClient, managerCtx, k8sManager)

		go func() {
			defer GinkgoRecover()
			Expect(k8sManager.Start(managerCtx)).To(Succeed())
		}()

		Expect(k8sClient.Create(managerCtx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: endpointsNs},
		})).To(Succeed())
		waitForSetupToComplete(k8sClient, managerCtx, webhookPort)
	})

	AfterAll(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &appsv1.DeploymentList{})
		deleteAllForType(managerCtx, k8sClient, &networkingv1.IngressList{})
		deleteAllForType(managerCtx, k8sClient, &corev1.ServiceList{})
		managerCancel()
		waitForWebhookToBeClosed(webhookUrlValidate)

		_ = os.Unsetenv(workloadbuilders.ExternalUrlAuthContainerImageEnv)
		_ = os.Unsetenv(workloadconfig.EnvVarOauth2ProxyOidcIssuerUrl)
		_ = os.Unsetenv(workloadconfig.EnvVarControlPlaneEndpoint)
		_ = os.Unsetenv(workloadbuilders.WorkloadAuthSecretNameEnv)
		_ = os.Unsetenv(workloadconfig.EnvVarOauth2ProxyUsernameLoaderImage)
	})

	AfterEach(func() {
		deleteAllTestResources(managerCtx, k8sClient)
		deleteAllForType(managerCtx, k8sClient, &appsv1.DeploymentList{})
		deleteAllForType(managerCtx, k8sClient, &networkingv1.IngressList{})
		deleteAllForType(managerCtx, k8sClient, &corev1.ServiceList{})
	})

	Context("single endpoint", func() {
		It("reconciles a Service + auth proxy + Ingress end-to-end", func() {
			workloadName := "endpoints-single"
			jobset := test.NewJobSetBuilder(workloadName, endpointsNs).WithDefault().Build()

			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			_ = createEndpointsRunner(managerCtx, k8sClient, "runner-endpoints-single", endpointsNs, jobset,
				&endpoints.Config{
					Endpoints: []endpoints.Endpoint{{
						Services: []endpoints.Service{{
							Create: makeEndpointService("jupyter", 8888, workloadName).Create,
							Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
						}},
						ExternalAccess: &endpoints.ExternalAccess{
							Ingress: makeEndpointIngress("jupyter-ing", "jupyter.example.com", "jupyter", 8888),
						},
					}},
				})

			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: workloadName, Namespace: endpointsNs}, &jobsetv1alpha2.JobSet{})
			}, timeout, interval).Should(Succeed())

			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "jupyter", Namespace: endpointsNs}, &corev1.Service{})
			}, timeout, interval).Should(Succeed())

			expectedAuthProxy := "jupyter-auth-proxy"
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: expectedAuthProxy, Namespace: endpointsNs}, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			ingress := &networkingv1.Ingress{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "jupyter-ing", Namespace: endpointsNs}, ingress)
			}, timeout, interval).Should(Succeed())
			Expect(ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal(expectedAuthProxy))
		})
	})

	Context("two endpoints", func() {
		It("reconciles two endpoints independently with distinct auth-proxy resources", func() {
			workloadName := "endpoints-multi"
			jobset := test.NewJobSetBuilder(workloadName, endpointsNs).WithDefault().Build()

			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			_ = createEndpointsRunner(managerCtx, k8sClient, "runner-endpoints-multi", endpointsNs, jobset,
				&endpoints.Config{
					Endpoints: []endpoints.Endpoint{
						{
							Services: []endpoints.Service{{
								Create: makeEndpointService("jupyter", 8888, workloadName).Create,
								Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
							}},
							ExternalAccess: &endpoints.ExternalAccess{
								Ingress: makeEndpointIngress("jupyter-ing", "jupyter.example.com", "jupyter", 8888),
							},
						},
						{
							Services: []endpoints.Service{{
								Create: makeEndpointService("tensorboard", 6006, workloadName).Create,
							}},
						},
					},
				})

			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "jupyter", Namespace: endpointsNs}, &corev1.Service{})
			}, timeout, interval).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "tensorboard", Namespace: endpointsNs}, &corev1.Service{})
			}, timeout, interval).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "jupyter-auth-proxy", Namespace: endpointsNs}, &appsv1.Deployment{})
			}, timeout, interval).Should(Succeed())

			// tensorboard endpoint has no auth → no auth-proxy Deployment for it
			Consistently(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "tensorboard-auth-proxy", Namespace: endpointsNs}, &appsv1.Deployment{})
			}, timeout/2, interval).Should(WithTransform(errors.IsNotFound, BeTrue()))
		})
	})

	Context("internal-only endpoint", func() {
		It("creates only the workload Service when no ExternalAccess is set", func() {
			workloadName := "endpoints-internal"
			jobset := test.NewJobSetBuilder(workloadName, endpointsNs).WithDefault().Build()

			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			_ = createEndpointsRunner(managerCtx, k8sClient, "runner-endpoints-internal", endpointsNs, jobset,
				&endpoints.Config{
					Endpoints: []endpoints.Endpoint{{
						Services: []endpoints.Service{{
							Create: makeEndpointService("internal-svc", 9000, workloadName).Create,
						}},
					}},
				})

			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "internal-svc", Namespace: endpointsNs}, &corev1.Service{})
			}, timeout, interval).Should(Succeed())

			Consistently(func() int {
				deployList := &appsv1.DeploymentList{}
				_ = k8sClient.List(managerCtx, deployList, client.InNamespace(endpointsNs))
				return len(deployList.Items)
			}, timeout/2, interval).Should(Equal(0))

			Consistently(func() int {
				ingList := &networkingv1.IngressList{}
				_ = k8sClient.List(managerCtx, ingList, client.InNamespace(endpointsNs))
				return len(ingList.Items)
			}, timeout/2, interval).Should(Equal(0))
		})
	})

	Context("metrics Service lifecycle", func() {
		It("uses the attached workload identity, preserves labels, recreates deletion, and prunes removal", func() {
			workloadName := "endpoints-metrics"
			runnerName := "runner-endpoints-metrics"
			metricsName := workloadName + "-metrics"
			jobset := test.NewJobSetBuilder(workloadName, endpointsNs).WithDefault().Build()

			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))
			runner := createEndpointsRunner(managerCtx, k8sClient, runnerName, endpointsNs, jobset,
				&endpoints.Config{Endpoints: []endpoints.Endpoint{
					{Services: []endpoints.Service{makeEndpointService("serving", 8000, workloadName)}},
					{Services: []endpoints.Service{
						makeMetricsEndpointService(metricsName, workloadName, clusterconstants.LabelNIMWorkloadMetricsService),
					}},
				}})

			metricsKey := types.NamespacedName{Name: metricsName, Namespace: endpointsNs}
			metrics := &corev1.Service{}
			Eventually(func() error {
				return k8sClient.Get(managerCtx, metricsKey, metrics)
			}, timeout, interval).Should(Succeed())
			Expect(metrics.Labels).To(HaveKeyWithValue(clusterconstants.LabelWorkloadMetricsService, "true"))
			Expect(metrics.Labels).To(HaveKeyWithValue(clusterconstants.LabelNIMWorkloadMetricsService, "true"))
			Expect(metrics.Labels).To(HaveKeyWithValue(networking.LabelWorkloadRunner, runnerName))
			Expect(metrics.OwnerReferences).To(HaveLen(1))
			Expect(metrics.OwnerReferences[0].Name).To(Equal(workloadName))

			Expect(k8sClient.Delete(managerCtx, metrics)).To(Succeed())
			Eventually(func() error {
				return k8sClient.Get(managerCtx, metricsKey, &corev1.Service{})
			}, timeout, interval).Should(Succeed())

			Expect(k8sClient.Get(managerCtx, types.NamespacedName{Name: runnerName, Namespace: endpointsNs}, runner)).To(Succeed())
			Expect(endpoints.Set(runner, &endpoints.Config{Endpoints: []endpoints.Endpoint{
				{Services: []endpoints.Service{makeEndpointService("renamed-serving", 8000, workloadName)}},
			}})).To(Succeed())
			Expect(k8sClient.Update(managerCtx, runner)).To(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(managerCtx, metricsKey, &corev1.Service{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
			// Ordinary endpoint Services are deliberately outside metrics pruning.
			Expect(k8sClient.Get(managerCtx,
				types.NamespacedName{Name: "serving", Namespace: endpointsNs},
				&corev1.Service{})).To(Succeed())
		})
	})

	Context("mutual exclusivity", func() {
		It("rejects a runner that sets both spec.instructions.networking and the run.ai/endpoints annotation", func() {
			workloadName := "endpoints-mutex"
			jobset := test.NewJobSetBuilder(workloadName, endpointsNs).WithDefault().Build()

			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))

			runner := &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runner-endpoints-mutex",
					Namespace: endpointsNs,
					Annotations: map[string]string{
						constants.ControlPlaneWorkloadIdAnnotationForWorkload: uuid.New().String(),
					},
				},
				Spec: dromosv1alpha1.WorkloadRunnerSpec{
					Template: marshalRawExtension(jobset),
					Instructions: &dromosv1alpha1.Instructions{
						Networking: &dromosv1alpha1.NetworkingConfig{
							Services: []dromosv1alpha1.NetworkingServiceItem{{
								Create: &corev1.Service{
									ObjectMeta: metav1.ObjectMeta{Name: "legacy-svc"},
									Spec: corev1.ServiceSpec{
										Selector: map[string]string{"app": workloadName},
										Ports:    []corev1.ServicePort{{Port: 80}},
									},
								},
							}},
						},
					},
				},
			}
			Expect(endpoints.Set(runner, &endpoints.Config{
				Endpoints: []endpoints.Endpoint{{
					Services: []endpoints.Service{makeEndpointService("svc-a", 80, workloadName)},
				}},
			})).To(Succeed())

			err := k8sClient.Create(managerCtx, runner)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("mutually exclusive"))
		})
	})

	Context("coexistence", func() {
		It("a legacy CR and an annotation CR can live in the same namespace without interfering", func() {
			workloadLegacy := "endpoints-coexist-legacy"
			workloadAnnot := "endpoints-coexist-annot"
			jobsetLegacy := test.NewJobSetBuilder(workloadLegacy, endpointsNs).WithDefault().Build()
			jobsetAnnot := test.NewJobSetBuilder(workloadAnnot, endpointsNs).WithDefault().Build()

			_ = createKartaWithStatus(managerCtx, k8sClient, newKartaWithValidStatus(test.NewValidJobSetKarta()))

			_ = createWorkloadRunner(managerCtx, k8sClient, "runner-coexist-legacy", endpointsNs, jobsetLegacy,
				&dromosv1alpha1.Instructions{
					Networking: &dromosv1alpha1.NetworkingConfig{
						Services: []dromosv1alpha1.NetworkingServiceItem{{
							Create: &corev1.Service{
								ObjectMeta: metav1.ObjectMeta{Name: "legacy-only-svc"},
								Spec: corev1.ServiceSpec{
									Selector: map[string]string{"app": workloadLegacy},
									Ports:    []corev1.ServicePort{{Port: 80}},
								},
							},
						}},
					},
				})

			_ = createEndpointsRunner(managerCtx, k8sClient, "runner-coexist-annot", endpointsNs, jobsetAnnot,
				&endpoints.Config{
					Endpoints: []endpoints.Endpoint{{
						Services: []endpoints.Service{makeEndpointService("annot-only-svc", 8080, workloadAnnot)},
					}},
				})

			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "legacy-only-svc", Namespace: endpointsNs}, &corev1.Service{})
			}, timeout, interval).Should(Succeed())
			Eventually(func() error {
				return k8sClient.Get(managerCtx, types.NamespacedName{Name: "annot-only-svc", Namespace: endpointsNs}, &corev1.Service{})
			}, timeout, interval).Should(Succeed())
		})
	})
})
