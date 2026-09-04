/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("WorkloadsQueueRecord.ToClusterUpdate", func() {
	It("includes creator identity", func() {
		record := WorkloadsQueueRecord{
			ID:              "workload-id",
			Name:            "workload-name",
			CreatedBy:       "owner-uid",
			ResourceGroup:   "batch",
			ResourceVersion: "v1",
			ResourceKind:    "Job",
			DesiredPhase:    string(api.DESIREDPHASE_RUNNING),
		}

		Expect(record.ToClusterUpdate().CreatedBy).To(Equal("owner-uid"))
	})
})

var _ = Describe("NetworkingOwnershipContract", func() {
	It("round-trips through JSON", func() {
		configuration := WorkloadConfiguration{
			NetworkingOwnershipContract: &NetworkingOwnershipContract{Version: "v1"},
		}

		payload, err := json.Marshal(configuration)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(payload)).To(Equal(`{"networkingOwnershipContract":{"version":"v1"}}`))

		var decoded WorkloadConfiguration
		Expect(json.Unmarshal(payload, &decoded)).To(Succeed())
		Expect(decoded.NetworkingOwnershipContract).ToNot(BeNil())
		Expect(decoded.NetworkingOwnershipContract.Version).To(Equal("v1"))
	})
})

var _ = Describe("NetworkingConfiguration", func() {
	It("unmarshals a single endpoint object", func() {
		var cfg WorkloadConfiguration
		Expect(json.Unmarshal([]byte(`{
			"networking": {
				"services": [
					{
						"ref": {
							"name": "existing-service",
							"port": 8000
						}
					}
				]
			}
		}`), &cfg)).To(Succeed())
		Expect(cfg.Networking).ToNot(BeNil())

		endpoints := cfg.Networking.ToAPI()
		Expect(endpoints).ToNot(BeNil())
		Expect(*endpoints).To(HaveLen(1))
		services := (*endpoints)[0].Services
		Expect(services).ToNot(BeNil())
		Expect(*services).To(HaveLen(1))
		Expect((*services)[0].Ref).ToNot(BeNil())
		Expect((*services)[0].Ref.Name).To(Equal("existing-service"))
		Expect((*services)[0].Ref.Port).To(Equal(int32(8000)))
	})

	DescribeTable("unmarshals legacy create networking",
		func(payload string) {
			var cfg WorkloadConfiguration
			Expect(json.Unmarshal([]byte(payload), &cfg)).To(Succeed())
			Expect(cfg.Networking).ToNot(BeNil())

			endpoints := cfg.Networking.ToAPI("legacy-workload")
			Expect(endpoints).ToNot(BeNil())
			Expect(*endpoints).To(HaveLen(1))
			services := (*endpoints)[0].Services
			Expect(services).ToNot(BeNil())
			Expect(*services).To(HaveLen(1))
			spec := (*services)[0].Spec
			Expect(spec).ToNot(BeNil())
			Expect(spec.Name).To(Equal("legacy-workload"))
			Expect(spec.Type).ToNot(BeNil())
			Expect(*spec.Type).To(Equal(api.SERVICETYPE_CLUSTER_IP))
			Expect(spec.Selector).ToNot(BeNil())
			Expect((*spec.Selector)["app"]).To(Equal("legacy-workload"))
			Expect(spec.Ports).To(HaveLen(1))
			Expect(spec.Ports[0].Port).To(Equal(int32(8080)))

			agentConfig := cfg.Networking.ToAgentAPI("legacy-workload")
			Expect(agentConfig).ToNot(BeNil())
			Expect(agentConfig.Services).ToNot(BeNil())
			Expect(*agentConfig.Services).To(HaveLen(1))
			create := (*agentConfig.Services)[0].Create
			Expect(create).ToNot(BeNil())
			Expect(create.Port).To(Equal(int32(8080)))
			Expect(create.Selector).ToNot(BeNil())
			Expect((*create.Selector)["app"]).To(Equal("legacy-workload"))
		},
		Entry("single endpoint object", `{
			"networking": {
				"services": [
					{
						"create": {
							"port": 8080,
							"selector": {
								"app": "legacy-workload"
							}
						}
					}
				]
			}
		}`),
		Entry("endpoint array", `{
			"networking": [
				{
					"services": [
						{
							"create": {
								"port": 8080,
								"selector": {
									"app": "legacy-workload"
								}
							}
						}
					]
				}
			]
		}`),
	)

	It("unmarshals an endpoint array", func() {
		var cfg WorkloadConfiguration
		Expect(json.Unmarshal([]byte(`{
			"networking": [
				{
					"services": [
						{
							"ref": {
								"name": "existing-service",
								"port": 8000
							}
						}
					]
				}
			]
		}`), &cfg)).To(Succeed())
		Expect(cfg.Networking).ToNot(BeNil())

		endpoints := cfg.Networking.ToAPI()
		Expect(endpoints).ToNot(BeNil())
		Expect(*endpoints).To(HaveLen(1))
	})

	It("projects compatible service specs to agent API", func() {
		portName := "http"
		nodePort := int32(30080)
		serviceType := api.SERVICETYPE_NODE_PORT
		selector := map[string]string{"app": "workload"}
		config := NetworkingConfiguration{{
			Services: &[]api.NetworkingService{{
				Spec: &api.ServiceSpec{
					Name:     "workload",
					Type:     &serviceType,
					Selector: &selector,
					Ports: []api.ServicePort{{
						Name:     &portName,
						Port:     8080,
						NodePort: &nodePort,
					}},
				},
			}},
		}}

		Expect(config.ToAgentAPI("workload")).To(BeNil())

		config[0].Services = &[]api.NetworkingService{{
			Spec: &api.ServiceSpec{
				Name:     "workload",
				Selector: &selector,
				Ports: []api.ServicePort{{
					Port: 8080,
				}},
			},
			Ref: &api.ServiceRef{Name: "workload", Port: 8080},
		}}
		Expect(config.AgentNetworkingCompatibilityError("workload")).To(HaveOccurred())

		config[0].Services = &[]api.NetworkingService{{}}
		Expect(config.AgentNetworkingCompatibilityError("workload")).To(HaveOccurred())

		serviceType = api.SERVICETYPE_CLUSTER_IP
		config[0].Services = &[]api.NetworkingService{{
			Spec: &api.ServiceSpec{
				Name:     "workload",
				Type:     &serviceType,
				Selector: &selector,
				Ports: []api.ServicePort{{
					Port: 8080,
				}},
			},
		}}

		agentConfig := config.ToAgentAPI("workload")
		Expect(agentConfig).ToNot(BeNil())
		Expect(agentConfig.Services).ToNot(BeNil())
		Expect(*agentConfig.Services).To(HaveLen(1))
		create := (*agentConfig.Services)[0].Create
		Expect(create).ToNot(BeNil())
		Expect(create.Port).To(Equal(int32(8080)))
		Expect(create.Selector).ToNot(BeNil())
		Expect((*create.Selector)["app"]).To(Equal("workload"))
	})

	It("round-trips service labels without mutating stored configuration", func() {
		payload := []byte(`{
			"networking": [
				{"services":[{"spec":{
					"name":"serving","selector":{"app":"model"},"labels":{
						"run.ai/nim-workload":"true",
						"run.ai/nim-workload-metrics-service":"true",
						"ordinary":"keep"
					},"ports":[{"port":8000}]
				}}]},
				{"services":[{"spec":{
					"name":"metrics","selector":{"app":"model"},"labels":{
						"run.ai/vllm-workload":"true",
						"run.ai/vllm-workload-metrics-service":"true",
						"shared":"metrics"
					},"ports":[{"name":"metrics","port":9090,"targetPort":9091}]
				}}]}
			]
		}`)

		var cfg WorkloadConfiguration
		Expect(json.Unmarshal(payload, &cfg)).To(Succeed())

		first := cfg.Networking.ToAPI("workload")
		Expect(first).ToNot(BeNil())
		Expect(*first).To(HaveLen(2))
		Expect((*(*first)[0].Services)[0].Spec.Labels["ordinary"]).To(Equal("keep"))
		Expect((*(*first)[1].Services)[0].Spec.Labels["run.ai/vllm-workload-metrics-service"]).To(Equal("true"))

		(*(*first)[0].Services)[0].Spec.Labels["ordinary"] = "mutated"
		(*(*first)[0].Services)[0].Spec.Ports[0].Port = 1
		second := cfg.Networking.ToAPI("workload")
		Expect((*(*second)[0].Services)[0].Spec.Labels["ordinary"]).To(Equal("keep"))
		Expect((*(*second)[0].Services)[0].Spec.Ports[0].Port).To(Equal(int32(8000)))

		record := WorkloadsQueueRecord{Name: "workload", Configuration: &cfg}
		update := record.ToClusterUpdate()
		Expect(update.Configuration.NetworkingEndpoints).ToNot(BeNil())
		Expect(*update.Configuration.NetworkingEndpoints).To(HaveLen(2))
		Expect((*(*update.Configuration.NetworkingEndpoints)[1].Services)[0].Spec.Labels["shared"]).To(Equal("metrics"))
	})
})
