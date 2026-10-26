/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
)

func TestClusterService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Service Client")
}

const clusterUuid = "cluster-uuid"

var _ = Describe("Cluster Service Client", func() {
	var (
		server         *httptest.Server
		clusterService *ClusterService
		requestedQuery []url.Values
	)

	serve := func(path string, pages func(offset, limit int32) any) {
		requestedQuery = nil
		mux := http.NewServeMux()
		mux.HandleFunc(path, func(writer http.ResponseWriter, request *http.Request) {
			query := request.URL.Query()
			requestedQuery = append(requestedQuery, query)

			var offset, limit int32
			_, err := fmt.Sscan(query.Get("offset"), &offset)
			Expect(err).NotTo(HaveOccurred())
			_, err = fmt.Sscan(query.Get("limit"), &limit)
			Expect(err).NotTo(HaveOccurred())

			writer.Header().Set("Content-Type", "application/json")
			Expect(json.NewEncoder(writer).Encode(pages(offset, limit))).To(Succeed())
		})

		server = httptest.NewServer(mux)
		DeferCleanup(server.Close)

		var err error
		clusterService, err = NewClusterService(server.URL, "token")
		Expect(err).NotTo(HaveOccurred())
	}

	Context("GetNodeList", func() {
		It("requests exactly the given offset and limit, and returns that single page", func() {
			const totalNodes = 8
			serve("/api/v1/nodes", func(offset, limit int32) any {
				return openapi.Nodes{Nodes: nodePage(offset, limit, totalNodes)}
			})

			nodeList, err := clusterService.GetNodeList(context.Background(), clusterUuid, 5, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeNames(nodeList.GetNodes())).To(Equal(expectedNodeNames(totalNodes)[5:8]))

			Expect(requestedQuery).To(HaveLen(1))
			Expect(requestedQuery[0].Get("offset")).To(Equal("5"))
			Expect(requestedQuery[0].Get("limit")).To(Equal("3"))
			Expect(requestedQuery[0].Get("sortBy")).To(Equal(string(openapi.NODEFILTERSORTFIELDS_NAME)))
			Expect(requestedQuery[0].Get("filterBy")).To(Equal(fmt.Sprintf("clusterUuid==%s", clusterUuid)))
		})

		It("makes exactly one request, even when the page comes back full", func() {
			const totalNodes = 20
			serve("/api/v1/nodes", func(offset, limit int32) any {
				return openapi.Nodes{Nodes: nodePage(offset, limit, totalNodes)}
			})

			nodeList, err := clusterService.GetNodeList(context.Background(), clusterUuid, 0, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeList.GetNodes()).To(HaveLen(5))
			Expect(requestedQuery).To(HaveLen(1))
		})

		It("returns no nodes for an empty page", func() {
			serve("/api/v1/nodes", func(offset, limit int32) any {
				return openapi.Nodes{Nodes: []openapi.Node{}}
			})

			nodeList, err := clusterService.GetNodeList(context.Background(), clusterUuid, 0, 100)
			Expect(err).NotTo(HaveOccurred())
			Expect(nodeList.GetNodes()).To(BeEmpty())
		})

		It("propagates a failure from the server", func() {
			serve("/api/v1/nodes", func(offset, limit int32) any {
				return openapi.Nodes{Nodes: nodePage(offset, limit, 3)}
			})
			server.Config.Handler = http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
				writer.WriteHeader(http.StatusForbidden)
			})

			_, err := clusterService.GetNodeList(context.Background(), clusterUuid, 0, 100)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("GetNodepoolList", func() {
		It("collects every page of node pools", func() {
			const totalNodepools = int(maxPageSize) + 2
			serve("/api/v1/node-pools", func(offset, limit int32) any {
				return openapi.GetNodepools200Response{Nodepools: nodepoolPage(offset, limit, totalNodepools)}
			})

			nodepoolList, err := clusterService.GetNodepoolList(context.Background(),
				filter.NewBuilder().WithClusterID(clusterUuid).StringList())
			Expect(err).NotTo(HaveOccurred())
			Expect(nodepoolList.GetNodepools()).To(HaveLen(totalNodepools))
			Expect(requestedQuery).To(HaveLen(2))
			Expect(requestedQuery[1].Get("offset")).To(Equal("500"))
		})
	})

	Context("GetNodepoolMinimalList", func() {
		It("collects every page of minimal node pools", func() {
			const totalNodepools = int(maxPageSize) + 1
			serve("/api/v1/node-pools/minimal", func(offset, limit int32) any {
				return openapi.GetNodepoolsMinimal200Response{Nodepools: nodepoolMinimalPage(offset, limit, totalNodepools)}
			})

			nodepoolList, err := clusterService.GetNodepoolMinimalList(context.Background(),
				filter.NewBuilder().WithClusterID(clusterUuid).StringList())
			Expect(err).NotTo(HaveOccurred())
			Expect(nodepoolList.GetNodepools()).To(HaveLen(totalNodepools))
			Expect(requestedQuery).To(HaveLen(2))
		})
	})
})

func pageBounds(offset, limit int32, total int) (int, int) {
	start := min(int(offset), total)
	return start, min(start+int(limit), total)
}

func nodePage(offset, limit int32, total int) []openapi.Node {
	start, end := pageBounds(offset, limit, total)
	page := make([]openapi.Node, 0, end-start)
	for i := start; i < end; i++ {
		page = append(page, openapi.Node{Name: nodeName(i), ClusterUuid: clusterUuid, NodePool: "default"})
	}
	return page
}

func nodepoolPage(offset, limit int32, total int) []openapi.Nodepool {
	start, end := pageBounds(offset, limit, total)
	page := make([]openapi.Nodepool, 0, end-start)
	for i := start; i < end; i++ {
		page = append(page, openapi.Nodepool{Name: fmt.Sprintf("nodepool-%04d", i)})
	}
	return page
}

func nodepoolMinimalPage(offset, limit int32, total int) []openapi.NodepoolMinimal {
	start, end := pageBounds(offset, limit, total)
	page := make([]openapi.NodepoolMinimal, 0, end-start)
	for i := start; i < end; i++ {
		page = append(page, openapi.NodepoolMinimal{Name: fmt.Sprintf("nodepool-%04d", i)})
	}
	return page
}

func nodeName(index int) string {
	return fmt.Sprintf("node-%04d", index)
}

func nodeNames(nodes []openapi.Node) []string {
	names := make([]string, 0, len(nodes))
	for _, node := range nodes {
		names = append(names, node.Name)
	}
	return names
}

func expectedNodeNames(total int) []string {
	names := make([]string, 0, total)
	for i := range total {
		names = append(names, nodeName(i))
	}
	return names
}
