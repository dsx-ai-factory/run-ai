/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package notifications

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	assetsApi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Connection naming ownership", func() {
	DescribeTable("selects native names only for native workloads",
		func(workloadType workloadsApi.WorkloadType, prefix, expectedService, expectedPort string) {
			nsw := NewNotificationsComposer(string(workloadType), "model", "", "", "", "", "", time.Time{}, nil,
				composer.FieldsComposerOptions{WorkloadV2NetworkingSupported: true}).WithConnectionsPreProcessor("team", "")
			// Submission starts with native metadata, then selects the V2 composer when supported.
			if prefix != "" {
				nsw.WithConnectionsPreProcessor("team", prefix)
			}
			connections := []workloadsApi.Connection{{
				ConnectionType:  string(assetsApi.INTERNALCONNECTIONTYPE_LOAD_BALANCER),
				CorrelationName: ptr.To("port-1"),
			}}
			ports := field_types.NewPortsField(map[string]field_types.PortItemAndRules{
				"port-1": {Value: &field_types.PortItemValue{ServiceType: ptr.To("LoadBalancer"), ContainerPort: ptr.To(8080)}},
			})
			nsw.connectionsPreProcessor.HandlePorts(connections, ports, nil, nil)
			Expect(connections[0].GetServiceResourceName()).To(Equal(expectedService))
			Expect(connections[0].GetPortResourceName()).To(Equal(expectedPort))
		},
		Entry("V1 workspace", workloadsApi.WORKLOADTYPE_WORKSPACE, "", "iw-model-0-loadbalancer", "port-model-8080-0"),
		Entry("V1 training", workloadsApi.WORKLOADTYPE_TRAINING, "", "tw-model-0-loadbalancer", "port-model-8080-0"),
		Entry("V1 distributed", workloadsApi.WORKLOADTYPE_DISTRIBUTED, "", "dw-model-0-loadbalancer", "port-model-8080-0"),
		Entry("V1 inference", workloadsApi.WORKLOADTYPE_INFERENCE, "", "infw-model-0-loadbalancer", "port-model-8080-0"),
		Entry("V2 job overrides native names", workloadsApi.WORKLOADTYPE_TRAINING, composer.NetworkingResourcePrefixJob, "job-model-lb", "port-8080"),
		Entry("V2 Kubeflow overrides native names", workloadsApi.WORKLOADTYPE_DISTRIBUTED, composer.NetworkingResourcePrefixKubeflow, "kf-model-lb-worker", "port-8080"),
		Entry("V2 inference does not reference native port Services", workloadsApi.WORKLOADTYPE_INFERENCE, composer.NetworkingResourcePrefixInference, "", ""),
	)
})
