/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options", func() {
	var (
		pod        string
		container  string
		tailLines  int64
		timestamps bool
		follow     bool
		previous   bool
		limitBytes int64
		since      time.Duration
		sinceTime  string
	)

	BeforeEach(func() {
		pod = "example-pod"
		container = "example-container"
		tailLines = 100
		timestamps = true
		follow = true
		previous = true
		limitBytes = 1024
		since = 300 * time.Second
		sinceTime = "2023-05-29T12:00:00Z"
	})

	Describe("NewOptions", func() {
		It("creates a new Options instance with the provided values", func() {
			options := NewOptions(pod, container, sinceTime, tailLines, timestamps, follow, previous, limitBytes, since)

			Expect(options.Pod).To(Equal(pod))
			Expect(options.Container).To(Equal(container))
			Expect(options.TailLines).To(Equal(tailLines))
			Expect(options.Timestamps).To(BeTrue())
			Expect(options.Follow).To(BeTrue())
			Expect(options.Previous).To(BeTrue())
			Expect(options.LimitBytes).To(Equal(limitBytes))
			Expect(options.Since).To(Equal(since))
			Expect(options.SinceTime).To(Equal(sinceTime))
		})
	})

	Describe("ToQueryParams", func() {
		var options *Options

		BeforeEach(func() {
			options = NewOptions(pod, container, sinceTime, tailLines, timestamps, follow, previous, limitBytes, 0)
		})

		It("converts the options to query parameters", func() {
			queryParams, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())

			Expect(queryParams.Get(queryParamPod)).To(Equal(pod))
			Expect(queryParams.Get(queryParamContainer)).To(Equal(container))
			Expect(queryParams.Get(queryParamTailLines)).To(Equal(strconv.FormatInt(tailLines, 10)))
			Expect(queryParams.Get(queryParamTimestamps)).To(Equal(queryParamTrue))
			Expect(queryParams.Get(queryParamFollow)).To(Equal(queryParamTrue))
			Expect(queryParams.Get(queryParamPrevious)).To(Equal(queryParamTrue))
			Expect(queryParams.Get(queryParamLimitBytes)).To(Equal(strconv.FormatInt(limitBytes, 10)))
			Expect(queryParams.Get(queryParamSinceTime)).To(Equal(sinceTime))
		})

		It("omits empty or default values from the query parameters", func() {
			options.Pod = ""
			options.Container = ""
			options.TailLines = -1
			options.Timestamps = false
			options.Follow = false
			options.Previous = false
			options.LimitBytes = 0
			options.Since = 0
			options.SinceTime = ""

			queryParams, err := options.ToQueryParams()
			Expect(err).NotTo(HaveOccurred())

			Expect(queryParams.Get(queryParamPod)).To(BeEmpty())
			Expect(queryParams.Get(queryParamContainer)).To(BeEmpty())
			Expect(queryParams.Get(queryParamTailLines)).To(BeEmpty())
			Expect(queryParams.Get(queryParamTimestamps)).To(BeEmpty())
			Expect(queryParams.Get(queryParamFollow)).To(BeEmpty())
			Expect(queryParams.Get(queryParamPrevious)).To(BeEmpty())
			Expect(queryParams.Get(queryParamLimitBytes)).To(BeEmpty())
			Expect(queryParams.Get(queryParamSinceSeconds)).To(BeEmpty())
			Expect(queryParams.Get(queryParamSinceTime)).To(BeEmpty())
		})
		It("returns an error when the options are invalid", func() {
			options.TailLines = -2
			_, err := options.ToQueryParams()
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("Validate", func() {
		var options *Options

		BeforeEach(func() {
			options = NewOptions(pod, container, "", tailLines, timestamps, follow, previous, limitBytes, since)
		})

		It("returns nil when the options are valid", func() {
			err := options.Validate()
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error when the tail lines are invalid", func() {
			options.TailLines = -2
			err := options.Validate()
			Expect(err).To(HaveOccurred())
		})
		It("returns an error when the limitBytes are invalid", func() {
			options.LimitBytes = -2
			err := options.Validate()
			Expect(err).To(HaveOccurred())
		})
		It("returns an error when the sinceSeconds are invalid", func() {
			options.Since = -2
			err := options.Validate()
			Expect(err).To(HaveOccurred())
		})
		It("returns an error when both sinceTime and sinceSeconds are specified", func() {
			options.Since = 1
			options.SinceTime = "2023-05-29T12:00:00Z"
			err := options.Validate()
			Expect(err).To(HaveOccurred())
		})
	})
})
