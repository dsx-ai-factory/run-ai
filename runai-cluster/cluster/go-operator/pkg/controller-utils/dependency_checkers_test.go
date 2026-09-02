/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller_utils

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ocpconf "github.com/openshift/api/config/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("DependencyCheckers", func() {
	var (
		k8s rclient.Client
	)

	BeforeEach(func() {
		k8s = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
	})

	Context("Prometheus operator", func() {
		When("prometheus operator exists", func() {
			possibleOperatorLocations := []struct {
				deploymentName      string
				deploymentNamespace string
			}{
				{
					deploymentName:      consts.PrometheusOperatorDeploymentName,
					deploymentNamespace: consts.MonitoringNamespace,
				},
				{
					deploymentName:      consts.RancherMonitoringOperatorDeploymentName,
					deploymentNamespace: consts.RancherMonitoringSystemNamespace,
				},
			}

			for _, location := range possibleOperatorLocations {
				It(fmt.Sprintf("should return its namespace(%s) correctly", location.deploymentNamespace), func() {
					Expect(k8s.Create(apiAccessContext, &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: location.deploymentNamespace,
						},
					})).To(Succeed())

					Expect(k8s.Create(apiAccessContext, &appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      location.deploymentName,
							Namespace: location.deploymentNamespace,
						},
					})).To(Succeed())

					ns, err := GetPrometheusOperatorNamespace(apiAccessContext, k8s)
					Expect(err).ToNot(HaveOccurred())
					Expect(ns).To(Equal(location.deploymentNamespace))
				})
			}
		})

		It("should fail if prometheus operator is not found", func() {
			_, err := GetPrometheusOperatorNamespace(apiAccessContext, k8s)
			Expect(err).To(HaveOccurred())
		})

		Context("OCP", func() {
			BeforeEach(func() {
				k8s = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(
					&ocpconf.ClusterVersion{
						ObjectMeta: metav1.ObjectMeta{
							Name: "version",
						},
					},
					&v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: consts.OpenshiftMonitoringNamespace,
						},
					},
					&appsv1.Deployment{
						ObjectMeta: metav1.ObjectMeta{
							Name:      consts.PrometheusOperatorDeploymentName,
							Namespace: consts.OpenshiftMonitoringNamespace,
						},
					}).Build()

			})

			It("should return the openshift-monitoring namespace when a single operator is installed", func() {
				ns, err := GetPrometheusOperatorNamespace(apiAccessContext, k8s)
				Expect(err).ToNot(HaveOccurred())
				Expect(ns).To(Equal(consts.OpenshiftMonitoringNamespace))
			})

			It("should return the openshift-monitoring namespace when multiple operators are installed", func() {
				Expect(k8s.Create(apiAccessContext, &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "openshift-user-workload-monitoring",
					},
				})).To(Succeed())

				Expect(k8s.Create(apiAccessContext, &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      consts.PrometheusOperatorDeploymentName,
						Namespace: "openshift-user-workload-monitoring",
					},
				})).To(Succeed())

				ns, err := GetPrometheusOperatorNamespace(apiAccessContext, k8s)
				Expect(err).ToNot(HaveOccurred())
				Expect(ns).To(Equal(consts.OpenshiftMonitoringNamespace))
			})
		})
	})

	Context("DCGM exporter", func() {
		It("should return dcgm-exporter's service if it exists", func() {
			Expect(k8s.Create(apiAccessContext, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: consts.GPUOperatorNamespace,
				},
			})).To(Succeed())

			Expect(k8s.Create(apiAccessContext, &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.DCGMExporterServiceName,
					Namespace: consts.GPUOperatorNamespace,
					Labels: map[string]string{
						"app": "nvidia-dcgm-exporter",
					},
				},
			})).To(Succeed())

			svc, err := GetDCGMExporterService(apiAccessContext, k8s)
			Expect(svc.Name).To(Not(BeNil()))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should fail if dcgm-exporter's service is not found", func() {
			_, err := GetDCGMExporterService(apiAccessContext, k8s)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("DCGM exporter status check", func() {
		BeforeEach(func() {
			httpmock.Activate()
		})

		It("should return dcgm-exporter's service exist and no missing metrics", func() {
			Expect(k8s.Create(apiAccessContext, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: consts.GPUOperatorNamespace,
				},
			})).To(Succeed())

			Expect(k8s.Create(apiAccessContext, &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.DCGMExporterServiceName,
					Namespace: consts.GPUOperatorNamespace,
					Labels: map[string]string{
						"app": "nvidia-dcgm-exporter",
					},
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "100.0.0.127",
					Ports: []v1.ServicePort{
						{Port: 12345},
					},
				},
			})).To(Succeed())

			svc, err := GetDCGMExporterService(apiAccessContext, k8s)
			httpmock.RegisterResponder("GET", "http://100.0.0.127:12345/metrics", func(request *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
					"DCGM_FI_DEV_FB_USED": 1, "DCGM_FI_DEV_FB_FREE": 1, "DCGM_FI_DEV_GPU_UTIL": 1,
				})
				return resp, err
			})

			err, missingMetrics := VerifyNvidiaStackExistsInRightState(svc)
			Expect(err == nil)
			Expect(missingMetrics == "")
		})

		It("should return dcgm-exporter's service exist and there missing metric", func() {
			Expect(k8s.Create(apiAccessContext, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: consts.GPUOperatorNamespace,
				},
			})).To(Succeed())

			Expect(k8s.Create(apiAccessContext, &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      consts.DCGMExporterServiceName,
					Namespace: consts.GPUOperatorNamespace,
					Labels: map[string]string{
						"app": "nvidia-dcgm-exporter",
					},
				},
				Spec: v1.ServiceSpec{
					ClusterIP: "100.0.0.127",
					Ports: []v1.ServicePort{
						{Port: 12345},
					},
				},
			})).To(Succeed())

			Expect(k8s.Create(apiAccessContext, &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "dcgm-exporter-sdchjk123",
					Namespace: consts.GPUOperatorNamespace,
					Labels: map[string]string{
						"app": "nvidia-dcgm-exporter",
					},
				},
			})).To(Succeed())

			httpmock.RegisterResponder("GET", "http://100.0.0.127:12345/metrics", func(request *http.Request) (*http.Response, error) {
				resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
					"DCGM_FI_DEV_FB_USED": 1, "DCGM_FI_DEV_FB_FREE": 1,
				})
				return resp, err
			})

			svc, err := GetDCGMExporterService(apiAccessContext, k8s)
			err, missingMetrics := VerifyNvidiaStackExistsInRightState(svc)
			Expect(err == nil)
			Expect(strings.Contains(missingMetrics, "DCGM_FI_DEV_GPU_UTIL") == true)
		})
	})

	Context("GetComponentNamespace", func() {
		const (
			gatewayNamespace = "kourier-system"
			gatewayName      = "3scale-kourier-gateway"
		)

		var componentRegex = regexp.MustCompile("kourier.*gateway")

		BeforeEach(func() {
			Expect(k8s.Create(apiAccessContext, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: gatewayNamespace,
				},
			})).To(Succeed())
		})

		It("should return the namespace when the gateway is a Deployment", func() {
			Expect(k8s.Create(apiAccessContext, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatewayName,
					Namespace: gatewayNamespace,
				},
			})).To(Succeed())

			ns, err := GetComponentNamespace(apiAccessContext, k8s, componentRegex, "kourier gateway", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(ns).To(Equal(gatewayNamespace))
		})

		It("should return the namespace when the gateway is a DaemonSet", func() {
			Expect(k8s.Create(apiAccessContext, &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatewayName,
					Namespace: gatewayNamespace,
				},
			})).To(Succeed())

			ns, err := GetComponentNamespace(apiAccessContext, k8s, componentRegex, "kourier gateway", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(ns).To(Equal(gatewayNamespace))
		})

		It("should honor the expected namespace when the gateway is a DaemonSet", func() {
			Expect(k8s.Create(apiAccessContext, &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gatewayName,
					Namespace: gatewayNamespace,
				},
			})).To(Succeed())

			ns, err := GetComponentNamespace(apiAccessContext, k8s, componentRegex, "kourier gateway", gatewayNamespace)
			Expect(err).ToNot(HaveOccurred())
			Expect(ns).To(Equal(gatewayNamespace))

			_, err = GetComponentNamespace(apiAccessContext, k8s, componentRegex, "kourier gateway", "other-namespace")
			Expect(err).To(HaveOccurred())
		})

		It("should fail when neither a Deployment nor a DaemonSet matches", func() {
			_, err := GetComponentNamespace(apiAccessContext, k8s, componentRegex, "kourier gateway", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found in the cluster"))
		})
	})
})
