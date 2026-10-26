/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer

import (
	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Automatic exposed URLs", func() {
	DescribeTable("preserves the URL used by networking and authentication",
		func(domain, workload, suffix string, subdomain bool, expected string) {
			actual, err := ComposeAutoExposedURL(domain, "team", workload, suffix, subdomain)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual.String()).To(Equal(expected))
		},
		Entry("first subdomain URL", "https://example.com", "job", "", true, "https://team-job.example.com/"),
		Entry("subdomain suffix", "https://example.com", "job", "metrics", true, "https://team-job-metrics.example.com/"),
		Entry("path suffix", "https://example.com", "job", "metrics", false, "https://example.com/team/job/metrics"),
		Entry("configured path, query and fragment", "https://example.com/base?key=value#section", "job", "metrics", false,
			"https://example.com/team/job/metrics//base?key=value#section"),
		Entry("subdomain preserves port and path", "https://example.com:8443/base", "job", "", true,
			"https://team-job.example.com:8443/base"),
		Entry("long hostname uses the existing stable hash", "https://example.com", strings.Repeat("x", 60), "metrics", true,
			"https://team-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx-9e3eddb17a.example.com/"),
	)

	DescribeTable("rejects an unusable cluster domain",
		func(domain string) {
			actual, err := ComposeAutoExposedURL(domain, "team", "job", "", true)
			Expect(err).To(HaveOccurred())
			Expect(actual).To(BeNil())
		},
		Entry("empty", ""),
		Entry("malformed escape", "https://example.com/%zz"),
	)
})

var _ = Describe("Exposed URL connection names", func() {
	It("assigns the unsuffixed URL to the first automatic nickname, skipping nil and custom entries", func() {
		urls := field_types.NewExposedUrlField(map[string]field_types.ExposedUrlItemAndRules{
			"a-nil":    {},
			"b-custom": {Value: &field_types.ExposedUrlItemValue{CustomUrl: ptr.To(true), Url: ptr.To("https://custom.example/path")}},
			"c-auto":   {Value: &field_types.ExposedUrlItemValue{}},
			"d-auto":   {Value: &field_types.ExposedUrlItemValue{}},
		})
		actual, err := exposedURLsPerNick(urls, "team", "job", "https://example.com", false)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(HaveLen(3))
		Expect(actual["b-custom"].String()).To(Equal("https://custom.example/path"))
		Expect(actual["c-auto"].String()).To(Equal("https://example.com/team/job"))
		Expect(actual["d-auto"].String()).To(Equal("https://example.com/team/job/d-auto"))
	})

	DescribeTable("preserves validation of custom URLs",
		func(custom bool, rawURL, expectedError string) {
			urls := field_types.NewExposedUrlField(map[string]field_types.ExposedUrlItemAndRules{
				"url": {Value: &field_types.ExposedUrlItemValue{CustomUrl: ptr.To(custom), Url: ptr.To(rawURL)}},
			})
			actual, err := exposedURLsPerNick(urls, "team", "job", "https://example.com", true)
			Expect(err).To(MatchError(ContainSubstring(expectedError)))
			Expect(actual).To(BeNil())
		},
		Entry("custom URL is missing", true, "", "url is mandatory when customUrl is enabled"),
		Entry("custom URL flag is disabled", false, "https://custom.example", "url should only be provided for customUrl"),
		Entry("custom URL is malformed", true, "https://custom.example/%zz", "invalid URL escape"),
	)

	It("keeps long subdomains distinct when their prefixes match", func() {
		first, err := ComposeAutoExposedURL("https://example.com", "team", strings.Repeat("x", 60), "a", true)
		Expect(err).NotTo(HaveOccurred())
		second, err := ComposeAutoExposedURL("https://example.com", "team", strings.Repeat("x", 60), "b", true)
		Expect(err).NotTo(HaveOccurred())
		Expect(first.Host).NotTo(Equal(second.Host))
		Expect(strings.Split(first.Host, ".")[0]).To(HaveLen(63))
		Expect(strings.Split(second.Host, ".")[0]).To(HaveLen(63))
	})
})
