/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSubmit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Application Submit Suite")
}

var _ = Describe("BuildValues", func() {
	var tempDir string

	BeforeEach(func() {
		tempDir = GinkgoT().TempDir()
	})

	writeValuesFile := func(content string) string {
		path := filepath.Join(tempDir, "values.yaml")
		Expect(os.WriteFile(path, []byte(content), 0o600)).To(Succeed())
		return path
	}

	It("returns nil when no values file and no set overrides are given", func() {
		values, err := BuildValues("", nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(BeNil())
	})

	It("parses a values file into a map", func() {
		path := writeValuesFile("admin:\n  user: bob\nreplicas: 2\n")

		values, err := BuildValues(path, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveKeyWithValue("replicas", 2))
		Expect(values["admin"]).To(HaveKeyWithValue("user", "bob"))
	})

	It("applies set overrides without a values file", func() {
		values, err := BuildValues("", []string{"persistence.enabled=false"})

		Expect(err).NotTo(HaveOccurred())
		Expect(values["persistence"]).To(HaveKeyWithValue("enabled", false))
	})

	It("merges set overrides on top of the values file", func() {
		path := writeValuesFile("admin:\n  password: fromfile\n  user: bob\n")

		values, err := BuildValues(path, []string{"admin.password=secret123"})

		Expect(err).NotTo(HaveOccurred())
		admin := values["admin"]
		Expect(admin).To(HaveKeyWithValue("password", "secret123"))
		Expect(admin).To(HaveKeyWithValue("user", "bob"))
	})

	It("applies multiple set overrides in order", func() {
		values, err := BuildValues("", []string{"a=1", "a=2", "nested.key=v"})

		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveKeyWithValue("a", int64(2)))
		Expect(values["nested"]).To(HaveKeyWithValue("key", "v"))
	})

	It("fails when the values file does not exist", func() {
		_, err := BuildValues(filepath.Join(tempDir, "missing.yaml"), nil)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to read values file"))
	})

	It("fails when the values file is not valid YAML", func() {
		path := writeValuesFile("admin:\n\tbad: [unclosed\n")

		_, err := BuildValues(path, nil)

		Expect(err).To(MatchError(ContainSubstring("is not valid YAML")))
	})

	It("fails on an invalid set expression", func() {
		_, err := BuildValues("", []string{"a=1,b"})

		Expect(err).To(MatchError(ContainSubstring("invalid --set value")))
	})

	// Helm-specific --set syntax, matching helm's own strvals semantics.

	It("parses list values", func() {
		values, err := BuildValues("", []string{"servers={alpha,beta,gamma}"})

		Expect(err).NotTo(HaveOccurred())
		Expect(values["servers"]).To(Equal([]interface{}{"alpha", "beta", "gamma"}))
	})

	It("parses list index assignments", func() {
		values, err := BuildValues("", []string{"servers[0].port=80", "servers[1].port=8080"})

		Expect(err).NotTo(HaveOccurred())
		servers := values["servers"].([]interface{})
		Expect(servers[0]).To(HaveKeyWithValue("port", int64(80)))
		Expect(servers[1]).To(HaveKeyWithValue("port", int64(8080)))
	})

	It("honors escaped commas in values", func() {
		values, err := BuildValues("", []string{`tolerations=a\,b`})

		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveKeyWithValue("tolerations", "a,b"))
	})

	It("honors escaped dots in keys", func() {
		values, err := BuildValues("", []string{`nodeSelector.kubernetes\.io/role=master`})

		Expect(err).NotTo(HaveOccurred())
		Expect(values["nodeSelector"]).To(HaveKeyWithValue("kubernetes.io/role", "master"))
	})

	It("keeps values with spaces intact", func() {
		values, err := BuildValues("", []string{"greeting=hello world"})

		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveKeyWithValue("greeting", "hello world"))
	})

	It("converts typed scalars the way helm does", func() {
		values, err := BuildValues("", []string{"enabled=true", "replicas=3", "tag=null"})

		Expect(err).NotTo(HaveOccurred())
		Expect(values).To(HaveKeyWithValue("enabled", true))
		Expect(values).To(HaveKeyWithValue("replicas", int64(3)))
		Expect(values).To(HaveKeyWithValue("tag", BeNil()))
	})

	It("fails when a nested override targets a scalar from the values file", func() {
		// same as helm: --set admin.password=x over a scalar admin key is
		// rejected by strvals rather than silently replacing the scalar
		path := writeValuesFile("admin: plainstring\n")

		_, err := BuildValues(path, []string{"admin.password=secret"})

		Expect(err).To(MatchError(ContainSubstring(`invalid --set value "admin.password=secret"`)))
	})
})

var _ = Describe("BuildCreationRequest", func() {
	It("builds a request with chart URL, values and credential auth", func() {
		request := BuildCreationRequest(CreationOptions{
			ClusterId:         "cluster-1",
			ProjectId:         "4500001",
			Name:              "my-grafana",
			ChartURL:          "https://charts.example.com/grafana-11.0.0.tgz",
			Values:            map[string]interface{}{"replicas": 2},
			CredentialAssetId: "asset-1",
			UserCredential:    true,
		})

		Expect(request.ClusterId).To(Equal("cluster-1"))
		Expect(request.ProjectId).To(Equal("4500001"))
		Expect(request.Name).To(Equal("my-grafana"))
		Expect(request.Spec.ChartRef.GetChartURL()).To(Equal("https://charts.example.com/grafana-11.0.0.tgz"))
		Expect(request.Spec.Values).To(HaveKeyWithValue("replicas", 2))
		Expect(request.Spec.Auth.GetAssetId()).To(Equal("asset-1"))
		Expect(request.Spec.Auth.GetUserCredential()).To(BeTrue())
	})

	It("omits auth when no credential is given", func() {
		request := BuildCreationRequest(CreationOptions{
			ClusterId: "cluster-1",
			ProjectId: "4500001",
			Name:      "my-grafana",
			ChartURL:  "https://charts.example.com/grafana-11.0.0.tgz",
		})

		Expect(request.Spec.Auth).To(BeNil())
		Expect(request.Spec.Values).To(BeNil())
	})
})
