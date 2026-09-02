/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Tests")
}

var _ = Describe("Submit Command", func() {
	var (
		tmpDir       string
		manifestPath string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "submit-test-*")
		Expect(err).To(BeNil())

		// Create a valid test manifest
		manifestContent := `apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test-container
    image: nginx:latest`

		manifestPath = filepath.Join(tmpDir, "test-manifest.yaml")
		err = os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Context("Command structure", func() {
		It("should have correct command name", func() {
			Expect(workload.Submit.Use).To(Equal("submit [NAME]"))
		})

		It("should have a short description", func() {
			Expect(workload.Submit.Short).NotTo(BeEmpty())
		})

		It("should have a long description", func() {
			Expect(workload.Submit.Long).NotTo(BeEmpty())
		})

		It("should have examples", func() {
			Expect(workload.Submit.Example).NotTo(BeEmpty())
		})

		It("should accept maximum 1 argument", func() {
			err := workload.Submit.Args(workload.Submit, []string{"arg1"})
			Expect(err).To(BeNil())

			err = workload.Submit.Args(workload.Submit, []string{"arg1", "arg2"})
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Flag registration", func() {
		It("should have file flag", func() {
			flag := workload.Submit.Flags().Lookup("file")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("f"))
		})

		It("should have project flag", func() {
			flag := workload.Submit.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have priority flag", func() {
			flag := workload.Submit.Flags().Lookup("priority")
			Expect(flag).NotTo(BeNil())
		})

		It("should have category flag", func() {
			flag := workload.Submit.Flags().Lookup("category")
			Expect(flag).NotTo(BeNil())
		})

		It("should have name-prefix flag", func() {
			flag := workload.Submit.Flags().Lookup("name-prefix")
			Expect(flag).NotTo(BeNil())
		})
	})

	Context("Command help and documentation", func() {
		It("should have meaningful help text", func() {
			Expect(workload.Submit.Short).To(ContainSubstring("submit"))
			Expect(workload.Submit.Short).To(ContainSubstring("workload"))
		})

		It("should have comprehensive examples", func() {
			Expect(workload.Submit.Example).To(ContainSubstring("runai workload submit"))
			Expect(workload.Submit.Example).To(ContainSubstring("--file"))
			Expect(workload.Submit.Example).To(ContainSubstring("--project"))
			Expect(workload.Submit.Example).To(ContainSubstring("--name-prefix"))
			Expect(workload.Submit.Example).To(ContainSubstring("--priority"))
			Expect(workload.Submit.Example).To(ContainSubstring("--category"))
		})
	})
})
