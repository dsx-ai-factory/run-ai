/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import (
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Native connection names", func() {
	DescribeTable("preserves legacy service grouping and global port indexes", func(nick, serviceName, portName string) {
		ports := field_types.NewPortsField(map[string]field_types.PortItemAndRules{
			"a-loadbalancer": {Value: &field_types.PortItemValue{ServiceType: convutils.Ref("LoadBalancer"), ContainerPort: convutils.Ref(8080)}},
			"b-invalid":      {Value: &field_types.PortItemValue{ServiceType: convutils.Ref("invalid")}},
			"c-nodeport":     {Value: &field_types.PortItemValue{ServiceType: convutils.Ref("NodePort"), ContainerPort: convutils.Ref(8081)}},
			"d-serving":      {Value: &field_types.PortItemValue{ServiceType: convutils.Ref("ServingPort")}},
			"e-default":      {Value: &field_types.PortItemValue{ContainerPort: convutils.Ref(8082)}},
			"f-loadbalancer": {Value: &field_types.PortItemValue{ServiceType: convutils.Ref("LoadBalancer"), ContainerPort: convutils.Ref(8083)}},
			"g-nil":          {},
		})
		names := PortResourceNames(workloadsApi.WORKLOADTYPE_TRAINING, "train", ports)
		Expect(names[nick].ServiceName).To(Equal(serviceName))
		Expect(names[nick].PortName).To(Equal(portName))
	},
		Entry("first load balancer", "a-loadbalancer", "tw-train-1-loadbalancer", "port-train-8080-0"),
		Entry("node port keeps its global index", "c-nodeport", "tw-train-2-nodeport", "port-train-8081-1"),
		Entry("default service type sorts first", "e-default", "tw-train-0-clusterip", "port-train-8082-2"),
		Entry("second load balancer shares the service", "f-loadbalancer", "tw-train-1-loadbalancer", "port-train-8083-3"),
		Entry("invalid type is skipped", "b-invalid", "", ""),
		Entry("serving port is skipped", "d-serving", "", ""),
		Entry("nil item is skipped", "g-nil", "", ""),
	)

})
