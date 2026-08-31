/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package url_test

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/apis"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/logs"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("url.BuildClusterUrl", func() {
	var (
		baseUrl  string
		workload *apis.Workload
		action   string
		options  *logs.Options
	)

	BeforeEach(func() {
		baseUrl = "https://example.com"
		workload = &apis.Workload{
			Project:   "example-project",
			Type:      api.Workspace,
			Name:      "example-workload",
			Framework: api.Runai,
		}
		action = "logs"
		options = logs.NewOptions("example-pod", "example-container", "2023-05-29T12:00:00Z", 100, true, true, true, 1024, 0)
	})

	It("builds the cluster URL with the provided parameters", func() {
		query, err := options.ToQueryParams()
		Expect(err).NotTo(HaveOccurred())
		expectedEndpoint := fmt.Sprintf(url.Template, workload.Project, workload.Type.String(), workload.Framework.String(), workload.Name, action)
		expectedUrl := fmt.Sprintf("%s%s?%s", baseUrl, expectedEndpoint, query.Encode())

		resultUrl, err := url.BuildClusterUrl(baseUrl, workload, action, query)

		Expect(err).NotTo(HaveOccurred())
		Expect(resultUrl.String()).To(Equal(expectedUrl))
	})

	It("returns an error if the base URL is invalid", func() {
		invalidBaseUrl := ":"
		query, err := options.ToQueryParams()
		Expect(err).ToNot(HaveOccurred())
		resultUrl, err := url.BuildClusterUrl(invalidBaseUrl, workload, action, query)

		Expect(err).To(HaveOccurred())
		Expect(resultUrl).To(BeNil())
	})

	It("builds the cluster URL without query parameters if options is nil", func() {
		expectedEndpoint := fmt.Sprintf(url.Template, workload.Project, workload.Type.String(), workload.Framework.String(), workload.Name, action)
		expectedUrl := fmt.Sprintf("%s%s", baseUrl, expectedEndpoint)

		resultUrl, err := url.BuildClusterUrl(baseUrl, workload, action, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(resultUrl.String()).To(Equal(expectedUrl))
	})

	Context("External workloads", func() {
		BeforeEach(func() {
			workload = &apis.Workload{
				Project:   "example-project",
				Type:      api.Training,
				Name:      "external-workload-name",
				Framework: api.NimService,
			}
		})

		It("builds the cluster URL with external/external pattern for NimService", func() {
			expectedEndpoint := fmt.Sprintf(url.Template, workload.Project, url.External, url.External, workload.Name, action)
			expectedUrl := fmt.Sprintf("%s%s", baseUrl, expectedEndpoint)

			resultUrl, err := url.BuildClusterUrl(baseUrl, workload, action, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(resultUrl.String()).To(Equal(expectedUrl))
			Expect(resultUrl.String()).To(ContainSubstring("/external/external/"))
		})

		It("builds the cluster URL with query parameters for external workloads", func() {
			query, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())
			expectedEndpoint := fmt.Sprintf(url.Template, workload.Project, url.External, url.External, workload.Name, action)
			expectedUrl := fmt.Sprintf("%s%s?%s", baseUrl, expectedEndpoint, query.Encode())

			resultUrl, err := url.BuildClusterUrl(baseUrl, workload, action, query)

			Expect(err).NotTo(HaveOccurred())
			Expect(resultUrl.String()).To(Equal(expectedUrl))
			Expect(resultUrl.String()).To(ContainSubstring("/external/external/"))
			Expect(resultUrl.String()).To(ContainSubstring("pod=example-pod"))
		})
	})
})
