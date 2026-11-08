/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpu_pci_inventory

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	securityv1 "github.com/openshift/api/security/v1"
	yamlreader "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

var _ = Describe("OpenShift inventory SCC schema", func() {
	var document []byte
	BeforeEach(func() {
		path := filepath.Join("..", "..", "..", "..", "..", "runai-chart", "templates", "ocp", "gpu-pci-inventory-scc.yaml")
		source, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		// This template only uses the OpenShift condition; enable it without needing
		// Helm inside the operator's Go test image.
		tmpl, err := template.New("inventory-scc").Funcs(template.FuncMap{
			"include": func(name string, _ any) (string, error) {
				if name != "isOpenShift" {
					return "", fmt.Errorf("unexpected template helper %q", name)
				}
				return "true", nil
			},
		}).Parse(string(source))
		Expect(err).NotTo(HaveOccurred())
		var rendered bytes.Buffer
		values := map[string]any{"Values": map[string]any{"openShift": map[string]any{"securityContextConstraints": map[string]any{"create": true}}}}
		Expect(tmpl.Execute(&rendered, values)).To(Succeed())
		document, err = yamlreader.NewYAMLReader(bufio.NewReader(&rendered)).Read()
		Expect(err).NotTo(HaveOccurred())
	})

	It("strictly decodes the rendered SCC against the existing OpenShift API type", func() {
		scc, err := decodeInventorySCC(document)
		Expect(err).NotTo(HaveOccurred())
		Expect(scc.Kind).To(Equal("SecurityContextConstraints"))
		Expect(scc.APIVersion).To(Equal("security.openshift.io/v1"))
		Expect(scc.Name).To(Equal("gpu-pci-inventory"))
		Expect(scc.AllowHostDirVolumePlugin).To(BeTrue())
		Expect(scc.Users).To(BeEmpty())
		Expect(scc.Groups).To(BeEmpty())
	})

	DescribeTable("rejects fields outside the OpenShift SCC API", func(field string, value any) {
		var object map[string]any
		Expect(yaml.UnmarshalStrict(document, &object)).To(Succeed())
		object[field] = value
		invalid, err := yaml.Marshal(object)
		Expect(err).NotTo(HaveOccurred())
		_, err = decodeInventorySCC(invalid)
		Expect(err).To(MatchError(ContainSubstring("unknown field \"" + field + "\"")))
	},
		Entry("unsupported host-path restrictions", "allowedHostPaths", []any{map[string]any{"pathPrefix": "/sys", "readOnly": true}}),
		Entry("unknown fields", "unrecognizedSCCField", true),
	)
})

func decodeInventorySCC(document []byte) (*securityv1.SecurityContextConstraints, error) {
	data, err := yaml.YAMLToJSONStrict(document)
	if err != nil {
		return nil, err
	}
	var scc securityv1.SecurityContextConstraints
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&scc); err != nil {
		return nil, err
	}
	return &scc, nil
}
