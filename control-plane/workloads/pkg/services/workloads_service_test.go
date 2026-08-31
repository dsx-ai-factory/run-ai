/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestIsPendingOnly(t *testing.T) {
	t.Run("returns true for single Pending phase", func(t *testing.T) {
		assert.True(t, isPendingOnly([]api.Phase{api.PHASE_PENDING}))
	})

	t.Run("returns false for single Running phase", func(t *testing.T) {
		assert.False(t, isPendingOnly([]api.Phase{api.PHASE_RUNNING}))
	})

	t.Run("returns false for multiple phases including Pending", func(t *testing.T) {
		assert.False(t, isPendingOnly([]api.Phase{api.PHASE_PENDING, api.PHASE_RUNNING}))
	})

	t.Run("returns false for empty phases", func(t *testing.T) {
		assert.False(t, isPendingOnly([]api.Phase{}))
	})
}

var _ = Describe("suspendSupportLookupKind", func() {
	DescribeTable("resolves the concrete lookup kind",
		func(typeName, group, version, distributedFramework, expectedKind string) {
			Expect(suspendSupportLookupKind(typeName, group, version, distributedFramework)).To(Equal(expectedKind))
		},
		Entry("training v1 maps to Job", "Training", "batch", "v1", "", "Job"),
		Entry("workspace v1 maps to Job", "Workspace", "batch", "v1", "", "Job"),
		Entry("training non-v1 keeps public type", "Training", "batch", "v2", "", "Training"),
		Entry("PyTorch v1 maps to PyTorchJob", "Distributed", "kubeflow.org", "v1", string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH), "PyTorchJob"),
		Entry("TF v1 maps to TFJob", "Distributed", "kubeflow.org", "v1", string(api.DISTRIBUTEDFRAMEWORK_TF), "TFJob"),
		Entry("XGBoost v1 maps to XGBoostJob", "Distributed", "kubeflow.org", "v1", string(api.DISTRIBUTEDFRAMEWORK_XG_BOOST), "XGBoostJob"),
		Entry("JAX v1 maps to JAXJob", "Distributed", "kubeflow.org", "v1", string(api.DISTRIBUTEDFRAMEWORK_JAX), "JAXJob"),
		Entry("MPI v1 maps to MPIJob", "Distributed", "kubeflow.org", "v1", string(api.DISTRIBUTEDFRAMEWORK_MPI), "MPIJob"),
		Entry("MPI v2beta1 maps to MPIJob", "Distributed", "kubeflow.org", "v2beta1", string(api.DISTRIBUTEDFRAMEWORK_MPI), "MPIJob"),
		Entry("known framework with unregistered version still resolves concrete kind", "Distributed", "kubeflow.org", "v9", string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH), "PyTorchJob"),
		Entry("unknown framework keeps public type", "Distributed", "kubeflow.org", "v1", "Unknown", "Distributed"),
		Entry("job-suffixed framework keeps public type", "Distributed", "kubeflow.org", "v1", "PyTorchJob", "Distributed"),
		Entry("missing framework keeps public type", "Distributed", "kubeflow.org", "v1", "", "Distributed"),
		Entry("wrong group keeps public type", "Distributed", "example.org", "v2beta1", string(api.DISTRIBUTEDFRAMEWORK_MPI), "Distributed"),
	)
})

func TestValidateNodepoolGroupByPhase(t *testing.T) {
	t.Run("RequestedNodepools with Pending phase is valid", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{RequestedNodepools}, []api.Phase{api.PHASE_PENDING})
		assert.NoError(t, err)
	})

	t.Run("RequestedNodepools without Pending phase returns error", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{RequestedNodepools}, []api.Phase{api.PHASE_RUNNING})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "RequestedNodepools")
	})

	t.Run("RequestedNodepools with no phase filter returns error", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{RequestedNodepools}, []api.Phase{})
		assert.Error(t, err)
	})

	t.Run("CurrentNodepools with Pending phase returns error", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{CurrentNodepools}, []api.Phase{api.PHASE_PENDING})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "CurrentNodepools")
	})

	t.Run("CurrentNodepools with Running phase is valid", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{CurrentNodepools}, []api.Phase{api.PHASE_RUNNING})
		assert.NoError(t, err)
	})

	t.Run("CurrentNodepools with no phase filter is valid", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{CurrentNodepools}, []api.Phase{})
		assert.NoError(t, err)
	})

	t.Run("no nodepool groupBy is always valid", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{ClusterId, DepartmentId}, []api.Phase{api.PHASE_PENDING})
		assert.NoError(t, err)
	})

	t.Run("RequestedNodepools with multiple phases returns error", func(t *testing.T) {
		err := validateNodepoolGroupByPhase([]string{RequestedNodepools}, []api.Phase{api.PHASE_PENDING, api.PHASE_RUNNING})
		assert.Error(t, err)
	})
}

func TestConnectionToEndpoint(t *testing.T) {
	t.Run("https URL parses host, path, protocol, tls", func(t *testing.T) {
		conn := api.Connection{Url: "https://chat.example.com/v1/completions"}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "https://chat.example.com/v1/completions", ep.Url)
		assert.Equal(t, "chat.example.com", ep.Host)
		assert.Equal(t, "/v1/completions", ep.Path)
		assert.Equal(t, "HTTPS", ep.Protocol)
		assert.Equal(t, ptr.To(true), ep.TlsEnabled)
	})

	t.Run("http URL parses protocol, tls is false", func(t *testing.T) {
		conn := api.Connection{Url: "http://example.com:8080/api"}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "example.com", ep.Host)
		assert.Equal(t, int32(8080), ep.Port)
		assert.Equal(t, "/api", ep.Path)
		assert.Equal(t, "HTTP", ep.Protocol)
		assert.Equal(t, ptr.To(false), ep.TlsEnabled)
	})

	t.Run("ip:port URL parses host and port", func(t *testing.T) {
		conn := api.Connection{Url: "1.2.3.4:5678"}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "1.2.3.4", ep.Host)
		assert.Equal(t, int32(5678), ep.Port)
		assert.Equal(t, "", ep.Path)
		assert.Equal(t, "", ep.Protocol)
		assert.Nil(t, ep.TlsEnabled)
	})

	t.Run("URL without port falls back to ContainerPort", func(t *testing.T) {
		conn := api.Connection{Url: "https://example.com", ContainerPort: 3000}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "example.com", ep.Host)
		assert.Equal(t, int32(3000), ep.Port)
		assert.Equal(t, "HTTPS", ep.Protocol)
		assert.Equal(t, ptr.To(true), ep.TlsEnabled)
	})

	t.Run("URL with root path only strips path", func(t *testing.T) {
		conn := api.Connection{Url: "https://example.com/"}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "/", ep.Path)
	})

	t.Run("empty URL returns empty endpoint", func(t *testing.T) {
		conn := api.Connection{Url: ""}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "", ep.Url)
		assert.Equal(t, "", ep.Host)
		assert.Equal(t, "", ep.Path)
		assert.Equal(t, "", ep.Protocol)
		assert.Nil(t, ep.TlsEnabled)
	})

	t.Run("URL with port ignores ContainerPort", func(t *testing.T) {
		conn := api.Connection{Url: "https://example.com:443/path", ContainerPort: 8080}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, int32(443), ep.Port)
	})

	// Regression: url.Parse does not error on "host:port" — it parses as scheme=host, opaque=port.
	// The fix checks parsed.Opaque to detect this and falls back to net.SplitHostPort.
	t.Run("hostname:port without scheme uses SplitHostPort fallback", func(t *testing.T) {
		conn := api.Connection{Url: "myhost:9090"}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "myhost", ep.Host)
		assert.Equal(t, int32(9090), ep.Port)
		assert.Equal(t, "", ep.Protocol, "bare host:port should have no protocol")
		assert.Nil(t, ep.TlsEnabled)
		assert.Equal(t, "", ep.Path)
	})

	t.Run("ip:port with ContainerPort prefers parsed port", func(t *testing.T) {
		conn := api.Connection{Url: "10.0.0.1:8080", ContainerPort: 3000}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "10.0.0.1", ep.Host)
		assert.Equal(t, int32(8080), ep.Port, "should use port from URL, not ContainerPort")
	})

	t.Run("ip:port falls back to ContainerPort when SplitHostPort fails", func(t *testing.T) {
		conn := api.Connection{Url: "not-a-valid-url", ContainerPort: 4000}
		ep := connectionToEndpoint(conn)
		assert.Equal(t, "", ep.Host)
		assert.Equal(t, int32(4000), ep.Port, "should fall back to ContainerPort")
	})
}

func TestEndpointsFromConnections(t *testing.T) {
	t.Run("extracts endpoints from connections with URLs, skips empty", func(t *testing.T) {
		envs := []api.Environment{
			{
				Connections: []api.Connection{
					{Url: "https://a.com/path", ContainerPort: 80},
					{Url: ""},
					{Url: "10.0.0.1:8080"},
				},
			},
			{
				Connections: []api.Connection{
					{Url: "http://b.com:3000"},
				},
			},
		}
		endpoints := endpointsFromConnections(envs)
		assert.Len(t, endpoints, 3)

		assert.Equal(t, "a.com", endpoints[0].Host)
		assert.Equal(t, "/path", endpoints[0].Path)

		assert.Equal(t, "10.0.0.1", endpoints[1].Host)
		assert.Equal(t, int32(8080), endpoints[1].Port)

		assert.Equal(t, "b.com", endpoints[2].Host)
		assert.Equal(t, int32(3000), endpoints[2].Port)
	})

	t.Run("returns nil for no connections with URLs", func(t *testing.T) {
		envs := []api.Environment{
			{Connections: []api.Connection{{Url: ""}}},
		}
		endpoints := endpointsFromConnections(envs)
		assert.Nil(t, endpoints)
	})

	t.Run("returns nil for empty environments", func(t *testing.T) {
		endpoints := endpointsFromConnections([]api.Environment{})
		assert.Nil(t, endpoints)
	})
}

func TestContainsGroupBy(t *testing.T) {
	t.Run("finds existing key", func(t *testing.T) {
		assert.True(t, containsGroupBy([]string{ClusterId, CurrentNodepools}, CurrentNodepools))
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		assert.False(t, containsGroupBy([]string{ClusterId, DepartmentId}, CurrentNodepools))
	})

	t.Run("returns false for empty slice", func(t *testing.T) {
		assert.False(t, containsGroupBy([]string{}, CurrentNodepools))
	})
}
