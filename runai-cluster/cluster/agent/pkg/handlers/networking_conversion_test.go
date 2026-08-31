/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloadsManagerApi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func mapToNetworkingConfig(m map[string]interface{}) []workloadsManagerApi.NetworkingEndpoint {
	data, _ := json.Marshal([]map[string]interface{}{m})
	var config []workloadsManagerApi.NetworkingEndpoint
	json.Unmarshal(data, &config)
	return config
}

func mapToAgentNetworkingConfig(m map[string]interface{}) workloadsManagerApi.AgentNetworkingConfiguration {
	data, _ := json.Marshal(m)
	var config workloadsManagerApi.AgentNetworkingConfiguration
	json.Unmarshal(data, &config)
	return config
}

func endpointsConfigFromRunner(runner *dromosv1alpha1.WorkloadRunner) *endpoints.Config {
	config, err := endpoints.Get(runner)
	Expect(err).NotTo(HaveOccurred())
	Expect(config).NotTo(BeNil())
	return config
}

var _ = Describe("Networking Configuration Conversion", func() {
	var (
		manifest        *unstructured.Unstructured
		dummyWorkloadId = "dummy-workload-id"
	)

	BeforeEach(func() {
		manifest = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name":      "test-job",
					"namespace": "test-namespace",
				},
			},
		}
	})

	It("should set the endpoints annotation when configuration has endpoint-array networking", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "my-service",
						"selector": map[string]interface{}{"app": "my-app"},
						"ports": []interface{}{
							map[string]interface{}{
								"port":       float64(8000),
								"targetPort": float64(8000),
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		Expect(workloadRunner.Spec.Instructions).To(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services[0].Create).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Name).To(Equal("my-service"))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports[0].Port).To(Equal(int32(8000)))
	})

	It("should fall back to legacy networking when endpoint-array networking is not set", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"create": map[string]interface{}{
						"port":     float64(8080),
						"selector": map[string]interface{}{"app": "my-app"},
					},
				},
			},
		}
		legacyNetworking := mapToAgentNetworkingConfig(networking)
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{}
		capabilities.SetNetworking(legacyNetworking)

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		Expect(endpoints.Has(workloadRunner)).To(BeFalse())
		Expect(workloadRunner.Spec.Instructions.Networking).NotTo(BeNil())
		Expect(workloadRunner.Spec.Instructions.Networking.Services).To(HaveLen(1))
		Expect(workloadRunner.Spec.Instructions.Networking.Services[0].Create).NotTo(BeNil())
		Expect(workloadRunner.Spec.Instructions.Networking.Services[0].Create.Spec.Ports).To(HaveLen(1))
		Expect(workloadRunner.Spec.Instructions.Networking.Services[0].Create.Spec.Ports[0].Port).To(Equal(int32(8080)))
		Expect(workloadRunner.Spec.Instructions.Networking.Services[0].Create.Spec.Selector).To(Equal(map[string]string{"app": "my-app"}))
	})

	It("should prefer endpoint-array networking over legacy networking when both are set", func() {
		legacyNetworking := mapToAgentNetworkingConfig(map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"ref": map[string]interface{}{
						"name": "legacy-service",
						"port": float64(7000),
					},
				},
			},
		})
		endpointNetworking := mapToNetworkingConfig(map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "endpoint-service",
						"selector": map[string]interface{}{"app": "endpoint"},
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(8000),
							},
						},
					},
				},
			},
		})
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: endpointNetworking,
		}
		capabilities.SetNetworking(legacyNetworking)

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner.Spec.Instructions).To(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services[0].Create).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Name).To(Equal("endpoint-service"))
	})

	It("should set endpoints services[0].create when configuration has spec", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "my-service",
						"type":     "NodePort",
						"selector": map[string]interface{}{"app": "my-app"},
						"ports": []interface{}{
							map[string]interface{}{
								"name":       "http",
								"port":       float64(9090),
								"targetPort": float64(8080),
								"nodePort":   float64(30090),
								"protocol":   "UDP",
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services[0].Create).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Name).To(Equal("my-service"))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports[0].Name).To(Equal("http"))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports[0].Port).To(Equal(int32(9090)))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(8080)))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports[0].NodePort).To(Equal(int32(30090)))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolUDP))
		Expect(endpointConfig.Endpoints[0].Services[0].Create.Spec.Selector).To(Equal(map[string]string{"app": "my-app"}))
	})

	It("preserves labeled full services across multiple endpoints and the runner annotation", func() {
		apiConfig := []workloadsManagerApi.NetworkingEndpoint{}
		for _, endpoint := range []map[string]interface{}{
			{"services": []interface{}{map[string]interface{}{"spec": map[string]interface{}{
				"name":     "nim-serving",
				"type":     "ClusterIP",
				"selector": map[string]interface{}{"app": "model"},
				"labels": map[string]interface{}{
					"run.ai/nim-workload":                 "true",
					"run.ai/nim-workload-metrics-service": "true",
					"ordinary":                            "preserve",
				},
				"ports": []interface{}{map[string]interface{}{
					"name": "http", "port": float64(8000), "targetPort": float64(8001), "protocol": "TCP",
				}},
			}}}},
			{"services": []interface{}{map[string]interface{}{"spec": map[string]interface{}{
				"name":     "vllm-metrics",
				"type":     "NodePort",
				"selector": map[string]interface{}{"app": "model"},
				"labels": map[string]interface{}{
					"run.ai/vllm-workload":                 "true",
					"run.ai/vllm-workload-metrics-service": "true",
					"run.ai/workload-metrics-service":      "true",
					"shared":                               "metrics",
				},
				"ports": []interface{}{map[string]interface{}{
					"name": "metrics", "port": float64(9090), "targetPort": float64(9091),
					"nodePort": float64(30090), "protocol": "UDP",
				}},
			}}}},
		} {
			apiConfig = append(apiConfig, mapToNetworkingConfig(endpoint)[0])
		}

		converted, err := convertNetworkingEndpoints(apiConfig, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(converted.Endpoints).To(HaveLen(2))
		Expect(converted.Endpoints[0].Services[0].Create.Labels).To(Equal(map[string]string{
			"run.ai/nim-workload":                 "true",
			"run.ai/nim-workload-metrics-service": "true",
			"ordinary":                            "preserve",
		}))
		Expect(converted.Endpoints[1].Services[0].Create.Labels).To(Equal(map[string]string{
			"run.ai/vllm-workload":                 "true",
			"run.ai/vllm-workload-metrics-service": "true",
			"run.ai/workload-metrics-service":      "true",
			"shared":                               "metrics",
		}))
		Expect(converted.Endpoints[1].Services[0].Create.Spec.Type).To(Equal(corev1.ServiceTypeNodePort))
		Expect(converted.Endpoints[1].Services[0].Create.Spec.Ports[0].NodePort).To(Equal(int32(30090)))
		Expect(converted.Endpoints[1].Services[0].Create.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolUDP))

		runner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId,
			&workloadsManagerApi.WorkloadV2UpdateConfiguration{NetworkingEndpoints: apiConfig})
		Expect(err).NotTo(HaveOccurred())
		roundTripped := endpointsConfigFromRunner(runner)
		Expect(roundTripped.Endpoints).To(HaveLen(2))
		Expect(roundTripped.Endpoints[0].Services[0].Create.Labels).To(
			HaveKeyWithValue("ordinary", "preserve"))
		Expect(roundTripped.Endpoints[1].Services[0].Create.Labels).To(
			HaveKeyWithValue("run.ai/vllm-workload-metrics-service", "true"))
		Expect(roundTripped.Endpoints[1].Services[0].Create.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(9091)))
	})

	It("returns error when spec has port but no selector", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name": "my-service",
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(8080),
							},
						},
					},
				},
			},
		}
		apiConfig := mapToNetworkingConfig(networking)

		_, err := convertNetworkingEndpoints(apiConfig, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("selector is required"))
	})

	It("returns error when port is invalid", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "my-service",
						"selector": map[string]interface{}{"app": "my-app"},
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(0),
							},
						},
					},
				},
			},
		}
		apiConfig := mapToNetworkingConfig(networking)

		_, err := convertNetworkingEndpoints(apiConfig, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("port must be greater than 0"))
	})

	It("returns error when both spec and ref are set", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "my-service",
						"selector": map[string]interface{}{"app": "my-app"},
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(8080),
							},
						},
					},
					"ref": map[string]interface{}{
						"name": "existing-service",
						"port": float64(9090),
					},
				},
			},
		}
		apiConfig := mapToNetworkingConfig(networking)

		_, err := convertNetworkingEndpoints(apiConfig, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec and ref are mutually exclusive"))
	})

	It("returns error when neither spec nor ref are set", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"authorization": map[string]interface{}{
						"authorizationType": "public",
					},
				},
			},
		}
		apiConfig := mapToNetworkingConfig(networking)

		_, err := convertNetworkingEndpoints(apiConfig, false)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("either spec or ref must be set"))
	})

	It("should set endpoints services[0].ref when configuration has ref", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"ref": map[string]interface{}{
						"name": "my-existing-service",
						"port": float64(8080),
					},
					"authorization": map[string]interface{}{
						"authorizationType": "authenticatedUsers",
					},
				},
			},
			"externalAccess": map[string]interface{}{
				"ingress": map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"host": "existing.example.com",
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path":     "/",
											"pathType": "Prefix",
											"backend": map[string]interface{}{
												"service": map[string]interface{}{
													"name": "my-existing-service",
													"port": map[string]interface{}{
														"number": float64(8080),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints[0].Services).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services[0].Ref).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].Services[0].Ref.Name).To(Equal("my-existing-service"))
		Expect(endpointConfig.Endpoints[0].Services[0].Ref.Port).To(Equal(int32(8080)))
		Expect(endpointConfig.Endpoints[0].Services[0].Create).To(BeNil())
	})

	It("returns error from wrapWithWorkloadRunner when networking conversion fails", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name": "my-service",
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(8080),
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		_, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("selector is required"))
	})

	It("should not add Instructions when configuration has empty networking array", func() {
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: []workloadsManagerApi.NetworkingEndpoint{},
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		Expect(workloadRunner.Spec.Instructions).To(BeNil())
		Expect(endpoints.Has(workloadRunner)).To(BeFalse())
	})

	It("should keep storage instructions when endpoint-array networking is set", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "my-service",
						"selector": map[string]interface{}{"app": "my-app"},
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(8080),
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
			Pvcs:                []workloadsManagerApi.WorkloadV2UpdateConfigurationPvc{*newDummyAPIPvc()},
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())
		Expect(workloadRunner.Spec.Instructions).NotTo(BeNil())
		Expect(workloadRunner.Spec.Instructions.Networking).To(BeNil())
		Expect(workloadRunner.Spec.Instructions.Storage).NotTo(BeNil())
		Expect(workloadRunner.Spec.Instructions.Storage.NewPvcs).To(HaveLen(1))
		Expect(endpoints.Has(workloadRunner)).To(BeTrue())
	})

	It("should set endpoint auth when configuration uses authorization arrays", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"ref": map[string]interface{}{
						"name": "my-service",
						"port": float64(8000),
					},
					"authorization": map[string]interface{}{
						"authorizationType": "authorizedUsers",
						"authorizedUsers":   []interface{}{"user1@example.com", "user2@example.com"},
					},
				},
			},
			"externalAccess": map[string]interface{}{
				"ingress": map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"host": "auth.example.com",
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path":     "/",
											"pathType": "Prefix",
											"backend": map[string]interface{}{
												"service": map[string]interface{}{
													"name": "my-service",
													"port": map[string]interface{}{
														"number": float64(8000),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints[0].Services).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].Services[0].Auth).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].Services[0].Auth.AuthorizationType).To(Equal(endpoints.AuthorizationTypeAuthorizedUsers))
		Expect(endpointConfig.Endpoints[0].Services[0].Auth.AuthorizedUsers).To(Equal("user1@example.com,user2@example.com"))
		Expect(endpointConfig.Endpoints[0].Services[0].Auth.AuthorizedGroups).To(BeEmpty())
	})

	It("should set endpoint externalAccess when configuration has full ingress", func() {
		networking := map[string]interface{}{
			"externalAccess": map[string]interface{}{
				"ingress": map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"host": "myhost.example.com",
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path":     "/my-path",
											"pathType": "Prefix",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints[0].ExternalAccess).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].ExternalAccess.Ingress).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].ExternalAccess.Ingress.Spec.Rules).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].ExternalAccess.Ingress.Spec.Rules[0].Host).To(Equal("myhost.example.com"))
		Expect(endpointConfig.Endpoints[0].ExternalAccess.Ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/my-path"))
	})

	It("should not set externalAccess when it has no supported resource", func() {
		networking := map[string]interface{}{
			"services": []interface{}{
				map[string]interface{}{
					"spec": map[string]interface{}{
						"name":     "my-service",
						"selector": map[string]interface{}{"app": "my-app"},
						"ports": []interface{}{
							map[string]interface{}{
								"port": float64(8080),
							},
						},
					},
				},
			},
			"externalAccess": map[string]interface{}{},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints[0].ExternalAccess).To(BeNil())
	})

	It("should pass through full ingress (pathType ImplementationSpecific)", func() {
		networking := map[string]interface{}{
			"externalAccess": map[string]interface{}{
				"ingress": map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"metadata": map[string]interface{}{
						"annotations": map[string]interface{}{
							"nginx.ingress.kubernetes.io/rewrite-target": "/$2",
						},
					},
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"host": "host.example.com",
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path":     "/api/v1(/|$)(.*)",
											"pathType": "ImplementationSpecific",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints[0].ExternalAccess.Ingress).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].ExternalAccess.Ingress.Spec.Rules[0].HTTP.Paths[0].Path).To(Equal("/api/v1(/|$)(.*)"))
		Expect(*endpointConfig.Endpoints[0].ExternalAccess.Ingress.Spec.Rules[0].HTTP.Paths[0].PathType).To(Equal(networkingv1.PathTypeImplementationSpecific))
	})

	It("should set endpoint externalAccess.HTTPRoute when configuration has httpRoute", func() {
		networking := map[string]interface{}{
			"externalAccess": map[string]interface{}{
				"httpRoute": map[string]interface{}{
					"apiVersion": "gateway.networking.k8s.io/v1",
					"kind":       "HTTPRoute",
					"spec": map[string]interface{}{
						"parentRefs": []interface{}{
							map[string]interface{}{
								"name":      "my-gateway",
								"namespace": "gateway-ns",
							},
						},
						"hostnames": []interface{}{"myhost.example.com"},
						"rules": []interface{}{
							map[string]interface{}{
								"matches": []interface{}{
									map[string]interface{}{
										"path": map[string]interface{}{
											"type":  "PathPrefix",
											"value": "/my-path",
										},
									},
								},
								"backendRefs": []interface{}{
									map[string]interface{}{
										"name": "my-service",
										"port": float64(8080),
									},
								},
							},
						},
					},
				},
			},
		}
		capabilities := &workloadsManagerApi.WorkloadV2UpdateConfiguration{
			NetworkingEndpoints: mapToNetworkingConfig(networking),
		}

		workloadRunner, err := wrapWithWorkloadRunner(manifest, dummyWorkloadId, capabilities)

		Expect(err).NotTo(HaveOccurred())

		Expect(workloadRunner).NotTo(BeNil())
		endpointConfig := endpointsConfigFromRunner(workloadRunner)
		Expect(endpointConfig.Endpoints[0].ExternalAccess).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].ExternalAccess.HTTPRoute).NotTo(BeNil())
		Expect(endpointConfig.Endpoints[0].ExternalAccess.HTTPRoute.Spec.Hostnames).To(HaveLen(1))
		Expect(string(endpointConfig.Endpoints[0].ExternalAccess.HTTPRoute.Spec.Hostnames[0])).To(Equal("myhost.example.com"))
		Expect(endpointConfig.Endpoints[0].ExternalAccess.HTTPRoute.Spec.Rules).To(HaveLen(1))
		Expect(endpointConfig.Endpoints[0].ExternalAccess.HTTPRoute.Spec.Rules[0].BackendRefs).To(HaveLen(1))
		Expect(string(endpointConfig.Endpoints[0].ExternalAccess.HTTPRoute.Spec.Rules[0].BackendRefs[0].Name)).To(Equal("my-service"))
	})

	It("should keep HTTPRoute when selecting between Ingress and Route", func() {
		networking := map[string]interface{}{
			"externalAccess": map[string]interface{}{
				"ingress": map[string]interface{}{
					"apiVersion": "networking.k8s.io/v1",
					"kind":       "Ingress",
					"spec": map[string]interface{}{
						"rules": []interface{}{
							map[string]interface{}{
								"host": "ingress.example.com",
								"http": map[string]interface{}{
									"paths": []interface{}{
										map[string]interface{}{
											"path":     "/",
											"pathType": "Prefix",
										},
									},
								},
							},
						},
					},
				},
				"route": map[string]interface{}{
					"apiVersion": "route.openshift.io/v1",
					"kind":       "Route",
					"spec": map[string]interface{}{
						"host": "route.example.com",
					},
				},
				"httpRoute": map[string]interface{}{
					"apiVersion": "gateway.networking.k8s.io/v1",
					"kind":       "HTTPRoute",
					"spec": map[string]interface{}{
						"hostnames": []interface{}{"gateway.example.com"},
						"rules": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
		}
		apiConfig := mapToNetworkingConfig(networking)

		standardConfig, err := convertNetworkingEndpoints(apiConfig, false)

		Expect(err).NotTo(HaveOccurred())
		Expect(standardConfig.Endpoints[0].ExternalAccess.Ingress).NotTo(BeNil())
		Expect(standardConfig.Endpoints[0].ExternalAccess.Route).To(BeNil())
		Expect(standardConfig.Endpoints[0].ExternalAccess.HTTPRoute).NotTo(BeNil())

		openShiftConfig, err := convertNetworkingEndpoints(apiConfig, true)

		Expect(err).NotTo(HaveOccurred())
		Expect(openShiftConfig.Endpoints[0].ExternalAccess.Ingress).To(BeNil())
		Expect(openShiftConfig.Endpoints[0].ExternalAccess.Route).NotTo(BeNil())
		Expect(openShiftConfig.Endpoints[0].ExternalAccess.HTTPRoute).NotTo(BeNil())
	})
})
