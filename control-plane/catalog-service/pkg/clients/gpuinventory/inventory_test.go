/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpuinventory

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	clusterapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type nodePoolsFunc func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error)

func (f nodePoolsFunc) GetNodePoolsMinimalByCluster(ctx context.Context, id string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
	return f(ctx, id)
}

func (f nodePoolsFunc) GetClusterMinimalDataById(context.Context, string) (*clusterapi.ClustersMinimalData, error) {
	return &clusterapi.ClustersMinimalData{}, nil
}

const clusterID = "4b94efde-b05a-4118-a03a-1dedff81fa8e"

var _ = It("inventory merges per node counts by maximum within each capacity", func() {
	small, large, two, four, six := int64(40_000_000_000), int64(80_000_000_000), int64(2), int64(4), int64(6)
	client := New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		return &clusterapi.GetNodepoolsMinimal200Response{Nodepools: []clusterapi.NodepoolMinimal{
			{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{
				{Type: "A100", GpuMemoryBytes: &large, TotalDevices: 8, MaxDevicesPerNode: &four},
				{Type: "A100", GpuMemoryBytes: &small, TotalDevices: 2, MaxDevicesPerNode: &two},
			}},
			{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{
				{Type: "A100", GpuMemoryBytes: &large, TotalDevices: 6, MaxDevicesPerNode: &six},
				{Type: "A100", GpuMemoryBytes: &large, TotalDevices: 20}, // Unknown count cannot invent a 20-GPU node.
				{Type: "T4", TotalDevices: 8},
			}},
		}}, nil
	}), 30*time.Second)
	got, err := client.Get(context.Background(), clusterID)
	Expect(err).To(BeNil())
	Expect(got).To(HaveLen(3))
	Expect(got[0].MaxDevicesPerNode).To(BeNumerically("==", 2))
	Expect(got[1].MaxDevicesPerNode).To(BeNumerically("==", 6))
	Expect(got[2].MaxDevicesPerNode).To(BeNumerically("==", 0))
})

var _ = It("inventory rejects impossible per node counts", func() {
	for _, count := range []int64{-1, 0, 3} {
		client := New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
			return &clusterapi.GetNodepoolsMinimal200Response{Nodepools: []clusterapi.NodepoolMinimal{{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{
				{Type: "A100", TotalDevices: 2, MaxDevicesPerNode: &count},
			}}}}, nil
		}), 30*time.Second)
		{
			got, err := client.Get(context.Background(), clusterID)
			Expect(err).NotTo(BeNil())
			Expect(got).To(BeNil())
		}
	}
})

var _ = It("inventory keeps distinct capacities without claiming availability", func() {
	small, large := int64(40_000_000_000), int64(80_000_000_000)
	client := New(nodePoolsFunc(func(_ context.Context, id string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		Expect(id).To(Equal(clusterID))
		return &clusterapi.GetNodepoolsMinimal200Response{Nodepools: []clusterapi.NodepoolMinimal{
			{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{
				{Type: "A100", GpuMemoryBytes: &large, TotalDevices: 8},
				{Type: "A100", GpuMemoryBytes: &small, TotalDevices: 1},
			}},
			{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{
				{Type: "A100", GpuMemoryBytes: &large, TotalDevices: 4},
				{Type: "T4", TotalDevices: 1},
			}},
		}}, nil
	}), 30*time.Second)
	got, err := client.Get(context.Background(), clusterID)
	Expect(err).To(BeNil())
	Expect(got).To(HaveLen(3))
	Expect(got[0].Type).To(Equal("A100"))
	Expect(got[0].MemoryBytes).NotTo(BeNil())
	Expect(*got[0].MemoryBytes).To(Equal(small))
	Expect(got[1].MemoryBytes).NotTo(BeNil())
	Expect(*got[1].MemoryBytes).To(Equal(large))
	Expect(got[2].Type).To(Equal("T4"))
	Expect(got[2].MemoryBytes).To(BeNil())
})

var _ = It("inventory failure is not an empty cluster", func() {
	upstream := errors.New("cluster service unavailable")
	client := New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		return nil, upstream
	}), 30*time.Second)
	{
		_, err := client.Get(context.Background(), clusterID)
		Expect(err).NotTo(BeNil())
		Expect(err.Details).To(Equal(upstream.Error()))
	}
	client = New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		return nil, nil
	}), 30*time.Second)
	{
		_, err := client.Get(context.Background(), clusterID)
		Expect(err).NotTo(BeNil())
	}
})

var _ = It("inventory validates cluster before building filter", func() {
	client := New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		Fail("invalid cluster reached inventory")
		return nil, nil
	}), 30*time.Second)
	for _, id := range []string{"", "cluster-id,tenantId==other"} {
		{
			_, err := client.Get(context.Background(), id)
			Expect(err).NotTo(BeNil())
		}
	}
})

var _ = Describe("inventory reads every nodepool page", func() {
	for _, tc := range []struct {
		name       string
		total      int
		failSecond bool
	}{
		{"GPU on second page", 51, false}, {"exact full page", 50, false}, {"empty cluster", 0, false}, {"second page fails", 51, true},
	} {
		It(tc.name, func() {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				if r.URL.Path == "/api/v1/clusters/"+clusterID+"/minimal" {
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte(`{}`))
					Expect(err).To(BeNil())
					return
				}
				calls++
				Expect(r.URL.Query().Get("filterBy")).To(Equal("clusterId==" + clusterID))
				offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
				limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
				if limit == 0 {
					limit = 50
				}
				if offset > 0 && tc.failSecond {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				pools := []clusterapi.NodepoolMinimal{}
				for i := offset; i < tc.total && i < offset+limit; i++ {
					pool := clusterapi.NodepoolMinimal{}
					if i == tc.total-1 {
						memory := int64(80_000_000_000)
						pool.GpuTypes = []clusterapi.NodepoolGpuTypeDetails{{Type: "A100", GpuMemoryBytes: &memory, TotalDevices: 1}}
					}
					pools = append(pools, pool)
				}
				w.Header().Set("Content-Type", "application/json")
				{
					err := json.NewEncoder(w).Encode(clusterapi.GetNodepoolsMinimal200Response{Nodepools: pools})
					Expect(err).To(BeNil())
				}
			}))
			defer server.Close()
			hardware, err := New(clusterclient.NewClusterClient(server.URL), 30*time.Second).Get(context.Background(), clusterID)
			if tc.failSecond {
				Expect(err).NotTo(BeNil())
				Expect(hardware).To(BeNil())
				return
			}
			Expect(err).To(BeNil())
			if tc.total > 0 {
				Expect(hardware).To(HaveLen(1))
				Expect(hardware[0].Type).To(Equal("A100"))
			}
			if tc.total == 50 {
				Expect(calls).To(BeNumerically("==", 2))
			}
			if tc.total == 0 {
				Expect(hardware).To(HaveLen(0))
				Expect(calls).To(BeNumerically("==", 1))
			}
		})
	}
})

var _ = It("inventory rejects malformed records without partial success", func() {
	memory := int64(80_000_000_000)
	for _, invalid := range []clusterapi.NodepoolGpuTypeDetails{
		{Type: "A100", TotalDevices: -1}, {Type: "", TotalDevices: 1}, {Type: "  ", TotalDevices: 1},
	} {
		client := New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
			return &clusterapi.GetNodepoolsMinimal200Response{Nodepools: []clusterapi.NodepoolMinimal{{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{
				{Type: "A100", GpuMemoryBytes: &memory, TotalDevices: 2}, invalid,
			}}}}, nil
		}), 30*time.Second)
		{
			hardware, err := client.Get(context.Background(), clusterID)
			Expect(err).NotTo(BeNil())
			Expect(hardware).To(BeNil())
		}
	}
})

var _ = It("inventory adds deadline for caller without one", func() {
	client := New(nodePoolsFunc(func(ctx context.Context, _ string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		if _, bounded := ctx.Deadline(); !bounded {
			Fail("inventory upstream received an unbounded context")
			return nil, errors.New("missing deadline")
		}
		return &clusterapi.GetNodepoolsMinimal200Response{}, nil
	}), 30*time.Second)
	{
		_, err := client.Get(context.Background(), clusterID)
		Expect(err).To(BeNil())
	}
})

var _ = It("inventory timeout cancels blocked dependency", func() {
	for _, callerDeadline := range []bool{false, true} {
		ctx := context.Background()
		if callerDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 5*time.Millisecond)
			defer cancel()
		}
		client := New(nodePoolsFunc(func(ctx context.Context, _ string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		}), 20*time.Millisecond)
		start := time.Now()
		{
			result, err := client.Get(ctx, clusterID)
			Expect(result).To(BeNil())
			Expect(err).NotTo(BeNil())
			Expect(err.Details).To(ContainSubstring(context.DeadlineExceeded.Error()))
			Expect(time.Since(start)).To(BeNumerically("<=", time.Second))
		}
	}
})

var _ = It("inventory zero devices remain absent", func() {
	client := New(nodePoolsFunc(func(context.Context, string) (*clusterapi.GetNodepoolsMinimal200Response, error) {
		return &clusterapi.GetNodepoolsMinimal200Response{Nodepools: []clusterapi.NodepoolMinimal{{GpuTypes: []clusterapi.NodepoolGpuTypeDetails{{TotalDevices: 0}, {Type: "A100", TotalDevices: 0}}}}}, nil
	}), time.Second)
	{
		result, err := client.Get(context.Background(), clusterID)
		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(0))
	}
})
