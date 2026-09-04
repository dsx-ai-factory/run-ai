/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/cluster/stream"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options Exec", func() {
	Describe("Options", func() {
		It("should convert to k8s exec options", func() {
			options := stream.ExecOptions{
				Options: stream.Options{
					Stdin:     true,
					Stdout:    true,
					Stderr:    true,
					TTY:       true,
					Pod:       "not_reflected",
					Container: "container",
					Command:   []string{"this", "is", "command"},
				},
			}

			Expect(options.Validate()).NotTo(HaveOccurred())
			k8sOptions, err := options.ToK8sPodExecOptions()

			Expect(err).NotTo(HaveOccurred())
			Expect(k8sOptions.Stdin).To(Equal(true))
			Expect(k8sOptions.Stdout).To(Equal(true))
			Expect(k8sOptions.Stderr).To(Equal(true))
			Expect(k8sOptions.TTY).To(Equal(true))
			Expect(k8sOptions.Container).To(Equal("container"))
			Expect(len(k8sOptions.Command)).To(Equal(3))
			Expect(k8sOptions.Command).To(Equal([]string{"this", "is", "command"}))
		})

		It("should validate options", func() {
			validOptions := stream.ExecOptions{
				Options: stream.Options{Command: []string{"this", "is", "command"}},
			}
			Expect(validOptions.Validate()).NotTo(HaveOccurred())

			invalidOptions := []stream.ExecOptions{{
				Options: stream.Options{
					Stdin:     true,
					Stdout:    true,
					Stderr:    true,
					TTY:       true,
					Pod:       "not_reflected",
					Container: "container",
				},
			}}

			for _, options := range invalidOptions {
				Expect(options.Validate()).To(HaveOccurred())
			}
		})
	})

	Describe("ExecOptionsFromRequest", func() {
		It("should parse options from GET request without TTY set", func() {
			req, err := http.NewRequest(http.MethodGet,
				"https://example.com/exec?stdin=true&stdout=false&stderr=true&tty=false&pod=podder&container=container",
				nil)
			Expect(err).NotTo(HaveOccurred())

			cmd := []string{"ls", "/users", "|", "grep", "this is sparta"}
			req.Header.Set(HeaderCommand, encodeCommandString(cmd))

			options, err := stream.ExecOptionsFromRequest(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Stdin).To(Equal(true))
			Expect(options.Stdout).To(Equal(false))
			Expect(options.Stderr).To(Equal(true))
			Expect(options.TTY).To(Equal(false))
			Expect(options.Pod).To(Equal("podder"))
			Expect(options.Container).To(Equal("container"))
			Expect(options.Command).To(Equal(cmd))
			Expect(options.TTYWidth).To(Equal(uint16(0)))
			Expect(options.TTYHeight).To(Equal(uint16(0)))
		})

		It("should parse options from GET request with TTY size set", func() {
			req, err := http.NewRequest(http.MethodGet, "https://example.com/exec?stdout=true&tty=true&height=11&width=34", nil)
			Expect(err).NotTo(HaveOccurred())
			cmd := []string{"this is sparta 2"}
			req.Header.Set(HeaderCommand, encodeCommandString(cmd))

			options, err := stream.ExecOptionsFromRequest(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Stdin).To(Equal(false))
			Expect(options.Stdout).To(Equal(true))
			Expect(options.Stderr).To(Equal(false))
			Expect(options.TTY).To(Equal(true))
			Expect(options.Pod).To(Equal(""))
			Expect(options.Container).To(Equal(""))
			Expect(options.Command).To(Equal([]string{"this is sparta 2"}))
			Expect(options.TTYWidth).To(Equal(uint16(34)))
			Expect(options.TTYHeight).To(Equal(uint16(11)))
		})

		It("should parse options from GET request using terminal bash", func() {
			req, err := http.NewRequest(http.MethodGet,
				fmt.Sprintf("https://example.com/exec?command=%s", encodeCommandString([]string{"/bin/bash"})),
				nil)
			Expect(err).NotTo(HaveOccurred())

			options, err := stream.ExecOptionsFromRequest(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Command).To(Equal([]string{"/bin/bash"}))
		})

		It("should parse options from GET request using terminal sh", func() {
			req, err := http.NewRequest(http.MethodGet,
				fmt.Sprintf("https://example.com/exec?command=%s", encodeCommandString([]string{"/bin/sh"})),
				nil)
			Expect(err).NotTo(HaveOccurred())

			options, err := stream.ExecOptionsFromRequest(req)

			Expect(err).NotTo(HaveOccurred())
			Expect(options.Command).To(Equal([]string{"/bin/sh"}))
		})

		It("should fail when command in query is unknown", func() {
			req, err := http.NewRequest(http.MethodGet,
				"https://example.com/exec?command=bla",
				nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = stream.ExecOptionsFromRequest(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to retrieve command: failed to decode command (bla): illegal base64 data at input byte 0"))
		})

		It("should fail when command in query is empty", func() {
			req, err := http.NewRequest(http.MethodGet,
				"https://example.com/exec",
				nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = stream.ExecOptionsFromRequest(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to retrieve command: command header and command query parameters are both empty"))
		})

		It("should fail when command is set in header and in query", func() {
			req, err := http.NewRequest(http.MethodGet,
				"https://example.com/exec?command=bla",
				nil)
			Expect(err).NotTo(HaveOccurred())
			cmd := []string{"this is sparta"}
			req.Header.Set(HeaderCommand, encodeCommandString(cmd))

			_, err = stream.ExecOptionsFromRequest(req)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("failed to retrieve command: command header and command query parameters are mutually exclusive"))
		})

		It("should return an error for invalid GET options request", func() {
			req, err := http.NewRequest(http.MethodGet, "https://example.com/exec", nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = stream.ExecOptionsFromRequest(req)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for invalid Post options request", func() {
			req, err := http.NewRequest(http.MethodPost, "https://example.com/exec", nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = stream.ExecOptionsFromRequest(req)
			Expect(err).To(HaveOccurred())
		})

		It("should return an error for invalid Post options request", func() {
			br, _ := json.Marshal(stream.Options{
				Stdin:     true,
				Stdout:    false,
				Stderr:    true,
				TTY:       false,
				Pod:       "podder",
				Container: "container",
			})

			req, err := http.NewRequest(http.MethodPost, "https://example.com/exec", bytes.NewBuffer(br))
			Expect(err).NotTo(HaveOccurred())

			_, err = stream.ExecOptionsFromRequest(req)
			Expect(err).To(HaveOccurred())
		})
	})
})

func encodeCommandString(cmd []string) string {
	strByte, err := json.Marshal(cmd)
	if err != nil {
		return ""
	}

	return base64.StdEncoding.EncodeToString(strByte)
}
