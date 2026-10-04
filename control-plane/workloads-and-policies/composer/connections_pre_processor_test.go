/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"strings"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConnectionsPreProcessor", func() {
	Describe("ComposeNetworkingResourceName", func() {
		It("keeps networking resource names within the Kubernetes limit", func() {
			name := ComposeNetworkingResourceName(
				NetworkingResourcePrefixInference,
				"workload-20510b80092a-b54b0b",
				"url-authenticated-external-url",
			)

			Expect(name).To(Equal("infw-workload-20510b80092a-b54b0b-url-authenticated--e2e17e943f"))
			Expect(name).To(Equal(ComposeNetworkingResourceName(
				NetworkingResourcePrefixInference,
				"workload-20510b80092a-b54b0b",
				"url-authenticated-external-url",
			)))
		})

		It("does not collide when long names have the same truncated prefix", func() {
			first := ComposeNetworkingResourceName("infw", "workload-20510b80092a-b54b0b", "url-authenticated-external-url")
			second := ComposeNetworkingResourceName("infw", "workload-20510b80092a-b54b0b", "url-authenticated-external-user")

			Expect(first).NotTo(Equal(second))
		})

		It("leaves names at the limit unchanged", func() {
			name := "a" + strings.Repeat("b", 62)

			Expect(ComposeNetworkingResourceName("", "", name)).To(Equal(name))
		})
	})

	It("correlates Kubeflow worker port connections to WLR service and port resources", func() {
		cp := &ConnectionsPreProcessor{
			WorkloadName:                  "dist",
			ProjectName:                   "project",
			ResourceNamePrefix:            NetworkingResourcePrefixKubeflow,
			WorkloadV2NetworkingSupported: true,
		}
		replicaType := workloadsApi.REPLICATYPE_WORKER
		framework := workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH

		connections := []workloadsApi.Connection{{
			ConnectionType:  string(openapi.INTERNALCONNECTIONTYPE_LOAD_BALANCER),
			CorrelationName: convutils.Ref("port-worker"),
		}}
		cp.HandlePorts(connections, portsField("port-worker", "LoadBalancer", 8081), &replicaType, &framework)

		Expect(connections[0].GetServiceResourceName()).To(Equal("kf-dist-lb-worker"))
		Expect(connections[0].GetPortResourceName()).To(Equal("port-8081"))
	})

	It("correlates MPI master port connections to launcher WLR services", func() {
		cp := &ConnectionsPreProcessor{
			WorkloadName:                  "mpi",
			ProjectName:                   "project",
			ResourceNamePrefix:            NetworkingResourcePrefixKubeflow,
			WorkloadV2NetworkingSupported: true,
		}
		replicaType := workloadsApi.REPLICATYPE_MASTER
		framework := workloadsApi.DISTRIBUTEDFRAMEWORK_MPI

		connections := []workloadsApi.Connection{{
			ConnectionType:  string(openapi.INTERNALCONNECTIONTYPE_NODE_PORT),
			CorrelationName: convutils.Ref("port-master"),
		}}
		cp.HandlePorts(connections, portsField("port-master", "NodePort", 8082), &replicaType, &framework)

		Expect(connections[0].GetServiceResourceName()).To(Equal("kf-mpi-np-launcher"))
		Expect(connections[0].GetPortResourceName()).To(Equal("port-8082"))
	})

	It("correlates Kubeflow URL connections to WLR URL services", func() {
		cp := &ConnectionsPreProcessor{
			WorkloadName:                  "dist",
			ProjectName:                   "project",
			ResourceNamePrefix:            NetworkingResourcePrefixKubeflow,
			WorkloadV2NetworkingSupported: true,
		}
		replicaType := workloadsApi.REPLICATYPE_MASTER
		framework := workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH

		connections := []workloadsApi.Connection{{
			ConnectionType:  string(openapi.INTERNALCONNECTIONTYPE_EXTERNAL_URL),
			CorrelationName: convutils.Ref("url-master-1"),
		}}
		cp.HandleExposedUrls(connections, exposedURLsField("url-master-1", 8888), &replicaType, &framework)

		Expect(connections[0].GetServiceResourceName()).To(Equal("kf-dist-url-master-1"))
		Expect(connections[0].GetExposedUrlPath()).To(Equal("/project/dist"))
	})

	It("correlates inference URL connections to WLR URL services", func() {
		cp := &ConnectionsPreProcessor{
			WorkloadName:       "inference",
			ProjectName:        "project",
			ResourceNamePrefix: NetworkingResourcePrefixInference,
		}

		connections := []workloadsApi.Connection{{
			ConnectionType:  string(openapi.INTERNALCONNECTIONTYPE_EXTERNAL_URL),
			CorrelationName: convutils.Ref("url-1"),
		}}
		cp.HandleExposedUrls(connections, exposedURLsField("url-1", 8888), nil, nil)

		Expect(connections[0].GetServiceResourceName()).To(Equal("infw-inference-url-1"))
		Expect(connections[0].GetExposedUrlPath()).To(Equal("/project/inference"))
	})
})

func portsField(nick string, serviceType string, containerPort int) *field_types.PortsField {
	return field_types.NewPortsField(map[string]field_types.PortItemAndRules{
		nick: {
			Value: &field_types.PortItemValue{
				ServiceType:   convutils.Ref(serviceType),
				ContainerPort: convutils.Ref(containerPort),
			},
		},
	})
}

func exposedURLsField(nick string, containerPort int32) *field_types.ExposedUrlsField {
	return field_types.NewExposedUrlField(map[string]field_types.ExposedUrlItemAndRules{
		nick: {
			Value: &field_types.ExposedUrlItemValue{
				ContainerPort: convutils.Ref(containerPort),
			},
		},
	})
}
